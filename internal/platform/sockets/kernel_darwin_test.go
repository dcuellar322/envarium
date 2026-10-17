//go:build darwin

package sockets

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"testing"

	gnet "github.com/shirou/gopsutil/v4/net"
	"golang.org/x/sys/unix"
)

func pcbFixture(protocol string, family uint32, ip net.IP, pid int32) []byte {
	raw := make([]byte, 24)
	binary.LittleEndian.PutUint32(raw, 24)
	binary.LittleEndian.PutUint64(raw[8:], 100)
	kinds, sizes := []uint32{16, 1, 2, 4, 8}, []int{104, 104, 32, 32, 136}
	if protocol == "tcp" {
		kinds = append(kinds, 32)
		sizes = append(sizes, 204)
	}
	for index, kind := range kinds {
		record := make([]byte, (sizes[index]+7)&^7)
		binary.LittleEndian.PutUint32(record, uint32(sizes[index]))
		binary.LittleEndian.PutUint32(record[4:], kind)
		switch kind {
		case 16:
			binary.BigEndian.PutUint16(record[18:], 43210)
			binary.LittleEndian.PutUint64(record[28:], 99)
			if family == unix.AF_INET {
				record[44] = 1
				copy(record[76:80], ip.To4())
			} else {
				record[44] = 2
				copy(record[64:80], ip.To16())
			}
		case 1:
			proto := uint32(unix.IPPROTO_UDP)
			if protocol == "tcp" {
				proto = unix.IPPROTO_TCP
			}
			binary.LittleEndian.PutUint32(record[36:], proto)
			binary.LittleEndian.PutUint32(record[40:], family)
			binary.LittleEndian.PutUint32(record[68:], uint32(pid))
		case 32:
			binary.LittleEndian.PutUint32(record[36:], 1)
		}
		raw = append(raw, record...)
	}
	trailer := make([]byte, 24)
	binary.LittleEndian.PutUint32(trailer, 24)
	return append(raw, trailer...)
}

func TestKernelPCBFramingAddressAndOwner(t *testing.T) {
	t.Parallel()
	for _, proto := range []string{"tcp", "udp"} {
		for _, host := range []string{"127.0.0.1", "::1", "::"} {
			family := uint32(unix.AF_INET)
			if net.ParseIP(host).To4() == nil {
				family = unix.AF_INET6
			}
			raw := pcbFixture(proto, family, net.ParseIP(host), 123)
			got, err := parsePCBs(raw, proto)
			if err != nil || len(got) != 1 || got[0].Pid != 123 || got[0].Laddr.IP != host || got[0].Laddr.Port != 43210 || got[0].Family != family {
				t.Fatalf("%s %s: %v %v", proto, host, got, err)
			}
		}
	}
	raw := pcbFixture("tcp", unix.AF_INET, net.ParseIP("127.0.0.1"), 123)
	for _, mutate := range []func([]byte) []byte{
		func(b []byte) []byte { return b[:10] },
		func(b []byte) []byte { return b[:len(b)-24] },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[24:], 103); return b },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[28:], 99); return b },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[128+68:], 0xffffffff); return b },
	} {
		if _, err := parsePCBs(mutate(append([]byte(nil), raw...)), "tcp"); err == nil {
			t.Fatal("malformed PCB data accepted")
		}
	}
	// New-generation PCBs and established connections are excluded like netstat.
	binary.LittleEndian.PutUint64(raw[24+28:], 101)
	if values, err := parsePCBs(raw, "tcp"); err != nil || len(values) != 0 {
		t.Fatal("generation filter lost")
	}
	raw = pcbFixture("tcp", unix.AF_INET, net.ParseIP("127.0.0.1"), 123)
	binary.LittleEndian.PutUint32(raw[432+36:], 4)
	if values, err := parsePCBs(raw, "tcp"); err != nil || len(values) != 0 {
		t.Fatal("established TCP connection included")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := kernelConnections(ctx, "tcp"); err != context.Canceled {
		t.Fatalf("cancel: %v", err)
	}
}

func TestKernelSocketInventoryMatchesNetstatForLiveSockets(t *testing.T) {
	if os.Getenv("ENVARIUM_SOCKET_PARITY") != "1" {
		t.Skip("opt-in live macOS kernel/netstat comparison")
	}
	for _, network := range []string{"tcp4", "tcp6", "udp4", "udp6"} {
		t.Run(network, func(t *testing.T) {
			var port int
			host := "127.0.0.1:0"
			if network[len(network)-1] == '6' {
				host = "[::1]:0"
			}
			if network[:3] == "tcp" {
				listener, err := net.Listen(network, host)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = listener.Close() }()
				port = listener.Addr().(*net.TCPAddr).Port
			} else {
				listener, err := net.ListenPacket(network, host)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = listener.Close() }()
				port = listener.LocalAddr().(*net.UDPAddr).Port
			}
			protocol := network[:3]
			native, err := kernelConnections(context.Background(), protocol)
			if err != nil {
				t.Fatal(err)
			}
			fallback, err := netstatConnections(context.Background(), protocol)
			if err != nil {
				t.Fatal(err)
			}
			find := func(values []gnet.ConnectionStat) gnet.ConnectionStat {
				for _, v := range values {
					if v.Laddr.Port == uint32(port) && v.Pid == int32(os.Getpid()) {
						return v
					}
				}
				t.Fatal("live listener or owner absent")
				return gnet.ConnectionStat{}
			}
			a, b := find(native), find(fallback)
			if a.Laddr != b.Laddr || a.Family != b.Family || a.Type != b.Type || a.Status != b.Status {
				t.Fatalf("native=%v fallback=%v", a, b)
			}
		})
	}
}

func TestKernelFallbackPreservesEvidenceErrorsAndCancellation(t *testing.T) {
	t.Parallel()
	calls := 0
	denied := func(context.Context, string) ([]gnet.ConnectionStat, error) { return nil, unix.EPERM }
	fallback := func(context.Context, string) ([]gnet.ConnectionStat, error) {
		calls++
		return []gnet.ConnectionStat{{Pid: 42}}, nil
	}
	values, err := readSocketInventory(context.Background(), "tcp", denied, fallback)
	if err != nil || len(values) != 1 || values[0].Pid != 42 || calls != 1 {
		t.Fatal("kernel permission failure hid fallback evidence")
	}
	failed := func(context.Context, string) ([]gnet.ConnectionStat, error) { return nil, unix.EIO }
	if _, err := readSocketInventory(context.Background(), "tcp", denied, failed); !errors.Is(err, unix.EIO) || !errors.Is(err, unix.EPERM) {
		t.Fatal("both failures were not reported")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readSocketInventory(ctx, "tcp", denied, fallback); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal("cancellation launched fallback")
	}
}
