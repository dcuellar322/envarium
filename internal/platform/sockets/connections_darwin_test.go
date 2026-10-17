//go:build darwin

package sockets

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	gnet "github.com/shirou/gopsutil/v4/net"
)

const inventoryHeader = "Active Internet connections (including servers)\nProto Recv-Q Send-Q Local Address Foreign Address (state) rxbytes txbytes rhiwat shiwat pid epid state options\n"

func TestParseMacOSSocketInventory(t *testing.T) {
	tcp := inventoryHeader + "tcp4 0 0 127.0.0.1.3000 *.* LISTEN 0 0 128 128 42 0 0\ntcp6 0 0 ::1.3001 *.* LISTEN 0 0 128 128 43 0 0\ntcp46 0 0 *.3002 *.* LISTEN 0 0 128 128 44 0 0\ntcp4 0 0 127.0.0.1.99 127.0.0.1.22 ESTABLISHED 0 0 128 128 45 0 0\n"
	values, err := parseNetstat(tcp, "tcp")
	if err != nil || len(values) != 3 {
		t.Fatalf("TCP: %v %v", values, err)
	}
	for i, host := range []string{"127.0.0.1", "::1", "::"} {
		if values[i].Laddr.IP != host || values[i].Pid != int32(42+i) || values[i].Laddr.Port != uint32(3000+i) {
			t.Fatalf("TCP row: %+v", values[i])
		}
	}
	udp := inventoryHeader + "udp6 0 0 fe80::1%lo0.5353 *.* 0 0 128 128 123 0 0\nudp4 0 0 *.123 *.* 0 0 128 128 0 0 0\n"
	values, err = parseNetstat(udp, "udp")
	if err != nil || len(values) != 2 || values[0].Pid != 123 || values[0].Laddr.IP != "fe80::1%lo0" || values[1].Laddr.IP != "0.0.0.0" {
		t.Fatalf("UDP: %v %v", values, err)
	}
	for _, input := range []string{"", "tcp4 0 0 *.3000 *.* LISTEN", inventoryHeader + "tcp4 0 0 invalid *.* LISTEN 0 0 128 128 42", inventoryHeader + "tcp4 0 0 *.99999 *.* LISTEN 0 0 128 128 42", inventoryHeader + "tcp4 0 0 *.3000 *.* LISTEN 0 0 128 128 invalid"} {
		if _, err := parseNetstat(input, "tcp"); err == nil {
			t.Errorf("accepted invalid inventory %q", input)
		}
	}
	if values, err := parseNetstat(inventoryHeader, "tcp"); err != nil || len(values) != 0 {
		t.Fatalf("empty inventory: %v %v", values, err)
	}
	var output socketOutput
	if _, err := output.Write([]byte(strings.Repeat("x", 8<<20+1))); err == nil {
		t.Fatal("unbounded command output")
	}
}

func TestFreshMacOSInventoryObservesSocketOpenAndClose(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	port := uint32(listener.Addr().(*net.TCPAddr).Port)
	values, err := Connections(context.Background(), "tcp")
	if err != nil {
		t.Fatal(err)
	}
	if !hasSocket(values, port, int32(os.Getpid())) {
		t.Fatalf("missing live listener %s with PID %d", strconv.Itoa(int(port)), os.Getpid())
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	values, err = Connections(context.Background(), "tcp")
	if err != nil {
		t.Fatal(err)
	}
	if hasSocket(values, port, int32(os.Getpid())) {
		t.Fatal("fresh inventory retained closed listener")
	}
	udp, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = udp.Close() }()
	values, err = Connections(context.Background(), "inet")
	if err != nil {
		t.Fatal(err)
	}
	if !hasSocket(values, uint32(udp.LocalAddr().(*net.UDPAddr).Port), int32(os.Getpid())) {
		t.Fatal("UDP binding missing")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Connections(ctx, "tcp"); err == nil {
		t.Fatal("cancelled inventory succeeded")
	}
}

func hasSocket(values []gnet.ConnectionStat, port uint32, pid int32) bool {
	for _, value := range values {
		if value.Laddr.Port == port && value.Pid == pid {
			return true
		}
	}
	return false
}
