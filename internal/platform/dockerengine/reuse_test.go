package dockerengine

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moby/moby/client"
)

func TestConcurrentObservationWavesReuseConnections(t *testing.T) {
	const parallel = 8
	var connections atomic.Int32
	var mu sync.Mutex
	var gate chan struct{}
	arrivals := make(chan struct{}, parallel)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_ping" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		mu.Lock()
		current := gate
		mu.Unlock()
		arrivals <- struct{}{}
		<-current
		w.Header().Set("API-Version", "1.45")
		w.WriteHeader(http.StatusOK)
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	defer server.Close()
	t.Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	t.Setenv("DOCKER_TLS_VERIFY", "")
	t.Setenv("DOCKER_CERT_PATH", "")
	engine, err := (Connection{FromEnv: true}).NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = engine.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for wave := 0; wave < 2; wave++ {
		current := make(chan struct{})
		mu.Lock()
		gate = current
		mu.Unlock()
		errors := make(chan error, parallel)
		for range parallel {
			go func() { _, err := engine.Ping(ctx, client.PingOptions{}); errors <- err }()
		}
		for range parallel {
			select {
			case <-arrivals:
			case <-ctx.Done():
				close(current)
				t.Fatal(ctx.Err())
			}
		}
		close(current)
		for range parallel {
			if err := <-errors; err != nil {
				t.Fatal(err)
			}
		}
		if got := connections.Load(); got != parallel {
			t.Fatalf("wave %d opened %d connections; expected reuse of %d", wave, got, parallel)
		}
	}
}
