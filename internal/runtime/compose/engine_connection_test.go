package compose

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
)

func TestEngineConnectionNegotiatesWithoutDiscardedVersionRequest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a local Unix socket")
	}
	t.Parallel()
	// A short path also fits Darwin's Unix-domain socket pathname limit.
	directory, err := os.MkdirTemp("/tmp", "envarium-engine-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	address := filepath.Join(directory, "engine.sock")
	listener, err := net.Listen("unix", address)
	if err != nil {
		t.Fatal(err)
	}
	var pings, other, connections atomic.Int32
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_ping" {
			other.Add(1)
			http.Error(w, "metadata unavailable", http.StatusInternalServerError)
			return
		}
		version := "1.55"
		if pings.Add(1) > 1 {
			version = "1.54"
		}
		w.Header().Set("API-Version", version)
		w.Header().Set("OSType", "linux")
		w.WriteHeader(http.StatusOK)
	}), ConnState: func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}}
	t.Cleanup(func() { _ = server.Close() })
	go func() { _ = server.Serve(listener) }()
	factory := engineFactory{clients: newEngineClients()}
	t.Cleanup(func() { _ = factory.clients.close() })
	for _, version := range []string{"1.55", "1.54"} {
		engine, ping, err := factory.Connect(context.Background(), dockerConnection{Host: "unix://" + address})
		if err != nil {
			t.Fatal(err)
		}
		if ping.APIVersion != version {
			t.Fatalf("negotiation=%#v, want fresh %s", ping, version)
		}
		_ = engine.Close()
	}
	if pings.Load() != 2 || other.Load() != 0 || connections.Load() != 1 {
		t.Fatalf("ping requests=%d other requests=%d physical connections=%d", pings.Load(), other.Load(), connections.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := factory.Connect(ctx, dockerConnection{Host: "unix://" + address}); err == nil || factory.clients.count != 0 {
		t.Fatalf("canceled negotiation=%v retained clients=%d", err, factory.clients.count)
	}
}
