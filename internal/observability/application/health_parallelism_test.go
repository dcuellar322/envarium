package application

import (
	"context"
	"fmt"
	goruntime "runtime"
	"sync/atomic"
	"testing"
	"time"

	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

type parallelProbeSource map[string]runtime.ProjectRuntime

func (s parallelProbeSource) ListRuntimeProjectIDs(ctx context.Context) ([]string, error) {
	ids := make([]string, 0, len(s))
	for id := range s {
		ids = append(ids, id)
	}
	return ids, ctx.Err()
}
func (s parallelProbeSource) ResolveRuntime(ctx context.Context, id string) (runtime.ProjectRuntime, error) {
	return s[id], ctx.Err()
}

type barrierProbeEvaluator struct {
	ready   chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (e *barrierProbeEvaluator) Evaluate(ctx context.Context, project runtime.ProjectRuntime, _ runtime.Observation, _ runtime.HealthCheckDefinition) observability.HealthResult {
	if e.calls.Add(1) <= 4 {
		e.ready <- struct{}{}
	}
	select {
	case <-e.release:
	case <-ctx.Done():
	}
	return observability.HealthResult{ProjectID: project.ProjectID, ServiceID: "api", CheckID: "probe", Status: observability.StatusHealthy, ObservedAt: time.Now().UTC()}
}

func TestHealthIOConcurrencyWithSingleCPUExecutionSlot(t *testing.T) {
	previous := goruntime.GOMAXPROCS(1)
	defer goruntime.GOMAXPROCS(previous)
	source := parallelProbeSource{}
	for i := range 32 {
		id := fmt.Sprint(i)
		source[id] = runtime.ProjectRuntime{ProjectID: id, ManifestHash: "checks", Services: []runtime.ServiceDeclaration{{ID: "api", HealthChecks: []runtime.HealthCheckDefinition{{ID: "probe", ServiceID: "api", Type: "process", IntervalSeconds: 1}}}}}
	}
	evaluator := &barrierProbeEvaluator{ready: make(chan struct{}, 4), release: make(chan struct{})}
	repository := &healthRepositoryStub{}
	service := NewHealthService(source, &healthObserverStub{}, repository, evaluator)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.runDue(ctx, func(_ string, err error) { t.Errorf("scheduler: %v", err) })
	}()
	// All four workers must reach an I/O wait concurrently before any may finish.
	for range 4 {
		select {
		case <-evaluator.ready:
		case <-ctx.Done():
			t.Fatal("probe worker concurrency lost")
		}
	}
	close(evaluator.release)
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("probe completion was delayed")
	}
	if evaluator.calls.Load() != 32 || len(repository.results) != 32 {
		t.Fatal("probe result lost")
	}
	for _, schedule := range service.nextDue {
		if schedule.due.After(time.Now().Add(1200 * time.Millisecond)) {
			t.Fatal("one-second probe interval changed")
		}
	}
}
