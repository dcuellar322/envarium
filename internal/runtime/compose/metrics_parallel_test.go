package compose

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type delayedMetricEngine struct {
	*fakeEngine
	active  atomic.Int64
	maximum atomic.Int64
}

func (e *delayedMetricEngine) ContainerStats(ctx context.Context, id string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error) {
	active := e.active.Add(1)
	defer e.active.Add(-1)
	for {
		maximum := e.maximum.Load()
		if active <= maximum || e.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	select {
	case <-ctx.Done():
		return client.ContainerStatsResult{}, ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}
	return e.fakeEngine.ContainerStats(ctx, id, options)
}
func TestMetricReadsUseBoundedConcurrencyAndDoNotStarveLaterServices(t *testing.T) {
	t.Parallel()
	e := &delayedMetricEngine{fakeEngine: &fakeEngine{stats: map[string][]byte{}, inspects: map[string]container.InspectResponse{}}}
	for i := range 16 {
		id := fmt.Sprintf("service-%d", i)
		e.containers = append(e.containers, container.Summary{ID: id, Labels: map[string]string{labelProject: "fixture", labelService: id}})
		e.stats[id] = mustJSON(t, containerStatsFixture(time.Now(), 1, 2, 100, 0, 0, 0, 0, 1))
	}
	driver := &Driver{engine: fakeConnector{engine: e}}
	sink := &composeMetricSink{}
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()
	if err := driver.streamMetrics(ctx, domain.MetricRequest{Project: domain.ProjectRuntime{ProjectID: "one"}}, normalizedConfig{ProjectName: "fixture"}, sink); err != nil {
		t.Fatal(err)
	}
	if len(sink.samples) != 16 || e.maximum.Load() != 8 || e.active.Load() != 0 {
		t.Fatalf("samples=%d max=%d active=%d", len(sink.samples), e.maximum.Load(), e.active.Load())
	}
	for _, s := range sink.samples {
		if !s.CPUAvailable || !s.MemoryAvailable || s.Partial {
			t.Fatalf("missing service: %#v", s)
		}
	}
}
