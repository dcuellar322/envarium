package application

import (
	"context"
	"errors"
	"testing"

	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

type statusSource struct {
	runtime                   runtime.Observation
	health                    observability.ProjectHealth
	runtimeErr, healthErr     error
	runtimeCalls, healthCalls int
}

func (s *statusSource) Inspect(context.Context, string) (runtime.Observation, error) {
	s.runtimeCalls++
	return s.runtime, s.runtimeErr
}

func (s *statusSource) Get(context.Context, string) (observability.ProjectHealth, error) {
	s.healthCalls++
	return s.health, s.healthErr
}

func TestSingleAndBatchedRuntimePreserveHealthDerivedStates(t *testing.T) {
	t.Parallel()
	for _, item := range []struct {
		state     runtime.ProjectState
		health    observability.HealthStatus
		connected bool
		want      runtime.ProjectState
	}{
		{runtime.StateRunning, "unhealthy", true, runtime.StateDegraded},
		{runtime.StateRunningExternal, "unhealthy", true, runtime.StateDegraded},
		{runtime.StatePartiallyRunning, "unhealthy", true, runtime.StateDegraded},
		{runtime.StateRunning, "unknown", true, runtime.StateStarting},
		{runtime.StateRunningExternal, "unknown", true, runtime.StateStarting},
		{runtime.StateRunning, "unknown", false, runtime.StateRunning},
		{runtime.StateStopped, "unhealthy", true, runtime.StateStopped},
		{runtime.StateRunning, "healthy", true, runtime.StateRunning},
	} {
		source := &statusSource{runtime: runtime.Observation{ProjectID: "one", State: item.state}, health: observability.ProjectHealth{Status: item.health, Results: []observability.HealthResult{{CheckID: "ready"}}}}
		if item.connected {
			source.health.ObserverState = "connected"
		}
		observed, err := ObserveRuntime(context.Background(), source, source, "one")
		if err != nil || observed.State != item.want || observed.Services == nil {
			t.Fatalf("single %s/%s = %#v, %v", item.state, item.health, observed, err)
		}
		service := NewService(source, nil, nil, source)
		for _, kind := range []string{"runtime", "status"} {
			batch, err := service.Read(context.Background(), kind, []string{"one"})
			if err != nil || batch.Projects[0].Runtime == nil || batch.Projects[0].Runtime.State != item.want {
				t.Fatalf("batch %s %s/%s = %#v, %v", kind, item.state, item.health, batch, err)
			}
			if (batch.Projects[0].Health != nil) != (kind == "status") {
				t.Fatalf("unexpected health scope: %#v", batch.Projects[0])
			}
		}
	}
}

func TestStatusBatchRetainsPartialEvidenceWithoutLeakingErrors(t *testing.T) {
	t.Parallel()
	for _, failed := range []string{"runtime", "health"} {
		source := &statusSource{runtime: runtime.Observation{State: runtime.StateRunning}}
		failure := errors.New("private source failure")
		if failed == "runtime" {
			source.runtimeErr = failure
		} else {
			source.healthErr = failure
		}
		batch, err := NewService(source, nil, nil, source).Read(context.Background(), "status", []string{"one"})
		if err != nil {
			t.Fatal(err)
		}
		item := batch.Projects[0]
		if item.Warning == "" || item.Warning == failure.Error() || (item.Runtime == nil) != (failed == "runtime") || (item.Health == nil) != (failed == "health") {
			t.Fatalf("partial status = %#v", item)
		}
	}
	source := &statusSource{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ObserveRuntime(ctx, source, source, "one"); !errors.Is(err, context.Canceled) || source.runtimeCalls != 0 || source.healthCalls != 0 {
		t.Fatalf("canceled read = %v, calls=%d/%d", err, source.runtimeCalls, source.healthCalls)
	}
}
