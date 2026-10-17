//go:build darwin

package sockets

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"

	gnet "github.com/shirou/gopsutil/v4/net"
	"golang.org/x/sys/unix"
)

// PCB records are packed to four bytes and records are rounded to eight bytes.
// Validate the known ABI before reading offsets; unknown layouts use netstat.
// See ADR-0021 for Apple's exported layouts and compatibility policy.
func kernelConnections(ctx context.Context, protocol string) ([]gnet.ConnectionStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	raw, err := unix.SysctlRaw("net.inet." + protocol + ".pcblist_n")
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(raw) > 8<<20 {
		return nil, errors.New("kernel socket inventory exceeds 8 MiB")
	}
	return parsePCBs(raw, protocol)
}

type pcbGroup struct {
	mask             uint32
	inp, socket, tcp []byte
}

func parsePCBs(raw []byte, protocol string) ([]gnet.ConnectionStat, error) {
	if protocol != "tcp" && protocol != "udp" {
		return nil, errors.New("unsupported kernel socket protocol")
	}
	if len(raw) < 48 || binary.LittleEndian.Uint32(raw) != 24 {
		return nil, errors.New("invalid kernel socket header")
	}
	generation := binary.LittleEndian.Uint64(raw[8:])
	expected := uint32(31)
	if protocol == "tcp" {
		expected = 63
	}
	result := []gnet.ConnectionStat{}
	var group pcbGroup
	for at := 24; at < len(raw); {
		if len(raw)-at < 8 {
			return nil, errors.New("truncated kernel socket record")
		}
		length := int(binary.LittleEndian.Uint32(raw[at:]))
		kind := binary.LittleEndian.Uint32(raw[at+4:])
		if length == 24 {
			if at+24 != len(raw) || group.mask != 0 {
				return nil, errors.New("incomplete kernel socket group")
			}
			return result, nil
		}
		rounded := (length + 7) &^ 7
		if length < 8 || rounded > len(raw)-at {
			return nil, errors.New("invalid kernel socket record length")
		}
		record := raw[at : at+length]
		if err := group.add(record, kind); err != nil {
			return nil, err
		}
		at += rounded
		if group.mask != expected {
			continue
		}
		value, include, err := group.connection(protocol, generation)
		if err != nil {
			return nil, err
		}
		if include {
			result = append(result, value)
		}
		group = pcbGroup{}
	}
	return nil, errors.New("kernel socket trailer missing")
}

func (g *pcbGroup) add(record []byte, kind uint32) error {
	if kind == 0 || g.mask&kind != 0 {
		return errors.New("duplicate kernel socket record")
	}
	switch kind {
	case 16:
		if len(record) != 104 {
			return errors.New("unsupported kernel PCB layout")
		}
		g.inp = record
	case 1:
		if len(record) != 104 {
			return errors.New("unsupported kernel socket layout")
		}
		g.socket = record
	case 2, 4:
		if len(record) != 32 {
			return errors.New("unsupported kernel socket buffer layout")
		}
	case 8:
		if len(record) != 136 {
			return errors.New("unsupported kernel socket statistics layout")
		}
	case 32:
		if len(record) != 204 {
			return errors.New("unsupported kernel TCP layout")
		}
		g.tcp = record
	default:
		return errors.New("unknown kernel socket record kind")
	}
	g.mask |= kind
	return nil
}

func (g pcbGroup) connection(protocol string, generation uint64) (gnet.ConnectionStat, bool, error) {
	if len(g.inp) == 0 || len(g.socket) == 0 {
		return gnet.ConnectionStat{}, false, errors.New("kernel socket identity missing")
	}
	transport := uint32(unix.IPPROTO_UDP)
	socketType := uint32(unix.SOCK_DGRAM)
	status := ""
	if protocol == "tcp" {
		transport, socketType, status = unix.IPPROTO_TCP, unix.SOCK_STREAM, "LISTEN"
		if len(g.tcp) == 0 {
			return gnet.ConnectionStat{}, false, errors.New("kernel TCP state missing")
		}
		// External TCP records retain four timers; state follows at byte 36.
		if binary.LittleEndian.Uint32(g.tcp[36:]) != 1 {
			return gnet.ConnectionStat{}, false, nil
		}
	}
	if binary.LittleEndian.Uint64(g.inp[28:]) > generation || binary.LittleEndian.Uint32(g.socket[36:]) != transport {
		return gnet.ConnectionStat{}, false, nil
	}
	owner := binary.LittleEndian.Uint32(g.socket[68:])
	if owner > math.MaxInt32 {
		return gnet.ConnectionStat{}, false, errors.New("invalid kernel socket owner")
	}
	family := binary.LittleEndian.Uint32(g.socket[40:])
	flags := g.inp[44]
	var ip net.IP
	switch {
	case family == unix.AF_INET && flags&1 != 0:
		ip = net.IP(g.inp[76:80])
	case family == unix.AF_INET6 && flags&2 != 0:
		ip = net.IP(g.inp[64:80])
	default:
		return gnet.ConnectionStat{}, false, errors.New("inconsistent kernel socket address family")
	}
	address := ip.String()
	if family == unix.AF_INET6 && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast()) {
		scope := int(binary.BigEndian.Uint16(ip[2:4]))
		if scope != 0 {
			ip = append(net.IP(nil), ip...)
			ip[2], ip[3] = 0, 0
			zone := strconv.Itoa(scope)
			if iface, err := net.InterfaceByIndex(scope); err == nil {
				zone = iface.Name
			}
			address = fmt.Sprintf("%s%%%s", ip.String(), zone)
		}
	}
	return gnet.ConnectionStat{Family: family, Type: socketType, Pid: int32(owner), Status: status, Laddr: gnet.Addr{IP: address, Port: uint32(binary.BigEndian.Uint16(g.inp[18:]))}}, true, nil
}
