package compose

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func TestMetricSamplesReuseFiniteStatsConnection(t *testing.T) {
	var connections atomic.Int32
	stats := mustJSON(t, containerStatsFixture(time.Now().UTC(), 100, 200, 1000, 1, 2, 3, 4, 2))
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/stats"):
			if r.URL.Query().Get("stream") != "false" {
				t.Error("expected finite statistics response")
			}
			_, _ = w.Write(stats)
			w.(http.Flusher).Flush()
			// The terminating chunk can arrive after the JSON object. Decoding
			// that object alone must not discard the underlying connection.
			time.Sleep(20 * time.Millisecond)
		case strings.HasSuffix(r.URL.Path, "/json"):
			_, _ = w.Write([]byte(`{"RestartCount":2}`))
		default:
			http.NotFound(w, r)
		}
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	defer server.Close()
	engine, err := client.New(client.WithHost(server.URL), client.WithAPIVersion("1.45"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = engine.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sink := &composeMetricSink{}
	for range 3 {
		if err := sampleContainer(ctx, engine, domain.ProjectRuntime{ProjectID: "project"}, container.Summary{ID: "api"}, sink); err != nil {
			t.Fatal(err)
		}
	}
	if got := connections.Load(); got != 1 {
		t.Fatalf("three samples opened %d connections; expected reuse of one", got)
	}
	if len(sink.samples) != 3 || sink.samples[0].CPUPercent != 100 || sink.samples[0].RestartCount != 2 {
		t.Fatalf("fresh samples changed: %#v", sink.samples)
	}
}
