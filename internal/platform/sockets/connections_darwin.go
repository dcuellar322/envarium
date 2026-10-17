//go:build darwin

package sockets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	gnet "github.com/shirou/gopsutil/v4/net"
)

// Connections uses macOS's kernel socket inventory. Unlike lsof this does not
// walk every process's open files. Numeric addresses avoid DNS/service lookups.
// TCP callers need listeners only; UDP callers need every local binding.
func Connections(ctx context.Context, kind string) ([]gnet.ConnectionStat, error) {
	protocols := []string{kind}
	if kind == "inet" {
		protocols = []string{"tcp", "udp"}
	}
	result := []gnet.ConnectionStat{}
	for _, protocol := range protocols {
		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("unsupported socket inventory: %s", kind)
		}
		values, err := readSocketInventory(ctx, protocol, kernelConnections, netstatConnections)
		if err != nil {
			return nil, err
		}
		result = append(result, values...)
	}
	return result, nil
}

func netstatConnections(ctx context.Context, protocol string) ([]gnet.ConnectionStat, error) {
	command := exec.CommandContext(ctx, "/usr/sbin/netstat", "-anvlW", "-p", protocol)
	var output socketOutput
	command.Stdout = &output
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("read %s socket inventory: %w", protocol, err)
	}
	values, err := parseNetstat(output.String(), protocol)
	if err != nil {
		return nil, err
	}

	return values, nil
}

type socketOutput struct{ bytes.Buffer }

func (b *socketOutput) Write(value []byte) (int, error) {
	if b.Len()+len(value) > 8<<20 {
		return 0, errors.New("socket inventory exceeds 8 MiB")
	}
	return b.Buffer.Write(value)
}

func parseNetstat(output, protocol string) ([]gnet.ConnectionStat, error) {
	result := []gnet.ConnectionStat{}
	header := false
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "Proto" {
			// Fail explicitly if a future OS changes the layout; missing evidence must
			// not be represented as a successful empty inventory.
			header = len(fields) >= 13 && strings.Join(fields[8:13], " ") == "rxbytes txbytes rhiwat shiwat pid"
			continue
		}
		if !strings.HasPrefix(fields[0], protocol) {
			continue
		}
		pidColumn := 10
		if protocol == "udp" {
			pidColumn = 9
		}
		if !header || len(fields) <= pidColumn {
			return nil, errors.New("unrecognized macOS socket inventory layout")
		}
		if protocol == "tcp" && fields[5] != "LISTEN" {
			continue
		}
		address, err := netstatAddress(fields[3])
		if err != nil {
			return nil, err
		}
		pid, err := strconv.ParseInt(fields[pidColumn], 10, 32)
		if err != nil || pid < 0 {
			return nil, errors.New("invalid macOS socket owner PID")
		}
		family, socketType := uint32(syscall.AF_INET), uint32(syscall.SOCK_STREAM)
		if strings.Contains(fields[0], "6") {
			family = syscall.AF_INET6
		}
		if protocol == "udp" {
			socketType = syscall.SOCK_DGRAM
		}
		if address.IP == "*" {
			address.IP = "0.0.0.0"
			if family == syscall.AF_INET6 {
				address.IP = "::"
			}
		}
		status := ""
		if protocol == "tcp" {
			status = "LISTEN"
		}
		result = append(result, gnet.ConnectionStat{Family: family, Type: socketType, Laddr: address, Pid: int32(pid), Status: status})
	}
	if !header {
		return nil, errors.New("macOS socket inventory header missing")
	}
	return result, nil
}

func netstatAddress(value string) (gnet.Addr, error) {
	separator := strings.LastIndexByte(value, '.')
	if separator < 0 {
		return gnet.Addr{}, errors.New("invalid macOS socket address")
	}
	host, portText := value[:separator], value[separator+1:]
	if portText == "*" {
		return gnet.Addr{IP: host}, nil
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	ip, _, _ := strings.Cut(host, "%")
	if err != nil || host != "*" && net.ParseIP(ip) == nil {
		return gnet.Addr{}, errors.New("invalid macOS socket address")
	}
	return gnet.Addr{IP: host, Port: uint32(port)}, nil
}

type socketReader func(context.Context, string) ([]gnet.ConnectionStat, error)

func readSocketInventory(ctx context.Context, protocol string, kernel, fallback socketReader) ([]gnet.ConnectionStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	values, err := kernel(ctx, protocol)
	if err == nil {
		return values, nil
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	values, fallbackErr := fallback(ctx, protocol)
	if fallbackErr != nil {
		return nil, errors.Join(fmt.Errorf("kernel socket inventory: %w", err), fallbackErr)
	}
	return values, nil
}
