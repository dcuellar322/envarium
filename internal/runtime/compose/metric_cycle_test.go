package compose

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"errors"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"sync"
	"testing"
	"time"
)

type cycleCountingEngine struct {
	*fakeEngine
	mu           sync.Mutex
	lists, stats int
	reads        map[string]int
}

func (e *cycleCountingEngine) ContainerList(ctx context.Context, o client.ContainerListOptions) (client.ContainerListResult, error) {
	e.lists++
	return e.fakeEngine.ContainerList(ctx, o)
}
func (e *cycleCountingEngine) ContainerStats(ctx context.Context, id string, o client.ContainerStatsOptions) (client.ContainerStatsResult, error) {
	e.mu.Lock()
	e.stats++
	e.mu.Unlock()
	return e.fakeEngine.ContainerStats(ctx, id, o)
}
func (e *cycleCountingEngine) ContainerInspect(ctx context.Context, id string, o client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.reads[id]++
	result, err := e.fakeEngine.ContainerInspect(ctx, id, o)
	result.Container.RestartCount = e.stats
	return result, err
}
func TestMetricCycleReadsEachContainerOnceAndRetainsFreshRestartCounts(t *testing.T) {
	t.Parallel()
	engine := &cycleCountingEngine{fakeEngine: &fakeEngine{
		containers: []container.Summary{{ID: "running", State: container.StateRunning, Labels: map[string]string{labelProject: "fixture", labelService: "api"}}, {ID: "stopped", State: container.StateExited, Labels: map[string]string{labelProject: "fixture", labelService: "worker"}}},
		stats:      map[string][]byte{"running": mustJSON(t, containerStatsFixture(time.Now(), 1, 2, 100, 0, 0, 0, 0, 1))},
		inspects:   map[string]container.InspectResponse{"running": {State: &container.State{Running: true}}, "stopped": {State: &container.State{Running: false}}},
	}, reads: map[string]int{}}
	driver := &Driver{engine: fakeConnector{engine: engine}, managed: newManagedContainers()}
	project := domain.ProjectRuntime{ProjectID: "p", Services: []domain.ServiceDeclaration{{ID: "api", RuntimeName: "api"}, {ID: "worker", RuntimeName: "worker"}}}
	config := normalizedConfig{ProjectName: "fixture", Services: []string{"api", "worker"}}
	for cycle := 1; cycle <= 2; cycle++ {
		observation, samples, err := driver.observeMetricCycle(context.Background(), project, config)
		if err != nil || len(samples) != 1 || samples[0].Partial || !samples[0].CPUAvailable || samples[0].RestartCount != cycle || len(observation.Services) != 2 || observation.Services[0].Container.RestartCount != cycle {
			t.Fatalf("cycle=%d observation=%#v samples=%#v err=%v", cycle, observation, samples, err)
		}
	}
	if engine.lists != 2 || engine.stats != 2 || engine.reads["running"] != 2 || engine.reads["stopped"] != 2 {
		t.Fatalf("duplicate or inactive reads: %#v", engine)
	}
	failure := errors.New("stats denied")
	engine.statsErr = map[string]error{"running": failure}
	observation, samples, err := driver.observeMetricCycle(context.Background(), project, config)
	if !errors.Is(err, failure) || !observation.State.IsActive() || len(samples) != 1 || !samples[0].Partial {
		t.Fatalf("partial sampling lost status: %#v %v %v", observation, samples, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := engine.lists
	if _, _, err := driver.observeMetricCycle(ctx, project, config); !errors.Is(err, context.Canceled) || engine.lists != before {
		t.Fatal("canceled cycle read Engine", err)
	}
}
