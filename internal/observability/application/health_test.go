package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

type healthSourceStub struct{ project runtime.ProjectRuntime }

func (s healthSourceStub) ResolveRuntime(context.Context, string) (runtime.ProjectRuntime, error) {
	return s.project, nil
}
func (s healthSourceStub) ListRuntimeProjectIDs(context.Context) ([]string, error) {
	return []string{s.project.ProjectID}, nil
}

type healthObserverStub struct {
	mu    sync.Mutex
	calls int
}

func (s *healthObserverStub) Inspect(context.Context, string) (runtime.Observation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	state := "starting"
	if s.calls > 1 {
		state = "running"
	}
	return runtime.Observation{ProjectID: "project-1", State: runtime.StateRunning, Services: []runtime.ServiceObservation{{ID: "api", State: state}}, ObservedAt: time.Now().UTC()}, nil
}

type healthRepositoryStub struct {
	mu      sync.Mutex
	results []observability.HealthResult
}

func (s *healthRepositoryStub) AppendHealth(_ context.Context, result observability.HealthResult) error {
	s.mu.Lock()
	s.results = append(s.results, result)
	s.mu.Unlock()
	return nil
}
func (s *healthRepositoryStub) LatestHealth(context.Context, string) ([]observability.HealthResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]observability.HealthResult(nil), s.results...), nil
}
func (*healthRepositoryStub) PruneHealth(context.Context, time.Time) error { return nil }

type observationEvaluator struct{}

func (observationEvaluator) Evaluate(_ context.Context, _ runtime.ProjectRuntime, observation runtime.Observation, _ runtime.HealthCheckDefinition) observability.HealthResult {
	status := observability.StatusUnhealthy
	if observation.Services[0].State == "running" {
		status = observability.StatusHealthy
	}
	return observability.HealthResult{Status: status, Message: string(status), ObservedAt: time.Now().UTC()}
}

func TestWaitRequiredReobservesBetweenRetries(t *testing.T) {
	t.Parallel()
	project := runtime.ProjectRuntime{ProjectID: "project-1", Services: []runtime.ServiceDeclaration{{ID: "api", HealthChecks: []runtime.HealthCheckDefinition{{
		ID: "alive", ServiceID: "api", Type: "process", Required: true, Retries: 1,
	}}}}}
	observer := &healthObserverStub{}
	repository := &healthRepositoryStub{}
	service := NewHealthService(healthSourceStub{project: project}, observer, repository, observationEvaluator{})
	if err := service.WaitRequired(context.Background(), project.ProjectID); err != nil {
		t.Fatalf("WaitRequired() error = %v", err)
	}
	if observer.calls != 2 || len(repository.results) != 1 || repository.results[0].Status != observability.StatusHealthy {
		t.Fatalf("calls = %d, results = %#v", observer.calls, repository.results)
	}
}

type alwaysUnhealthy struct{}

func TestHealthDoesNotReportStoppedProjectsAsHealthyOrApplyOldFailures(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	for _, state := range []runtime.ProjectState{runtime.StateStopped, runtime.StatePaused} {
		result := aggregateHealth("project-1", runtime.Observation{State: state, ObservedAt: now}, []observability.HealthResult{
			{Status: observability.StatusUnhealthy, Required: true, ObservedAt: now},
		}, now)
		if result.Status != observability.StatusUnknown || len(result.Results) != 0 || result.ObserverState != observability.ObserverConnected {
			t.Fatalf("%s health = %#v", state, result)
		}
	}
	for _, state := range []runtime.ProjectState{runtime.StateFailed, runtime.StateDegraded, runtime.StatePartiallyRunning} {
		result := aggregateHealth("project-1", runtime.Observation{State: state, ObservedAt: now}, nil, now)
		if result.Status != observability.StatusUnhealthy {
			t.Fatalf("%s health = %#v", state, result)
		}
	}
}

func TestReadinessRetryStopsPromptlyOnCancellation(t *testing.T) {
	t.Parallel()
	service := NewHealthService(healthSourceStub{}, &healthObserverStub{}, &healthRepositoryStub{}, alwaysUnhealthy{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := service.runCheck(ctx, runtime.ProjectRuntime{}, runtime.Observation{}, runtime.HealthCheckDefinition{
		ID: "ready", Retries: 20, IntervalSeconds: 10,
	}, true)
	if result.Status != observability.StatusUnknown || result.Message != "health evaluation cancelled" {
		t.Fatalf("cancelled readiness = %#v", result)
	}
}

func (alwaysUnhealthy) Evaluate(context.Context, runtime.ProjectRuntime, runtime.Observation, runtime.HealthCheckDefinition) observability.HealthResult {
	return observability.HealthResult{Status: observability.StatusUnhealthy, Message: "not ready", ObservedAt: time.Now().UTC()}
}

func TestWaitRequiredReportsReadinessFailureWithoutRuntimeStop(t *testing.T) {
	t.Parallel()
	project := runtime.ProjectRuntime{ProjectID: "project-1", Services: []runtime.ServiceDeclaration{{ID: "api", HealthChecks: []runtime.HealthCheckDefinition{{
		ID: "ready", ServiceID: "api", Type: "http", Required: true,
	}}}}}
	service := NewHealthService(healthSourceStub{project: project}, &healthObserverStub{}, &healthRepositoryStub{}, alwaysUnhealthy{})
	err := service.WaitRequired(context.Background(), project.ProjectID)
	if !errors.Is(err, ErrRequiredHealthChecks) {
		t.Fatalf("WaitRequired() error = %v", err)
	}
}

func TestCompositeAnyUsesMemberResults(t *testing.T) {
	t.Parallel()
	results := map[string]observability.HealthResult{
		"api\x00one": {Status: observability.StatusUnhealthy},
		"api\x00two": {Status: observability.StatusHealthy},
	}
	result := compositeResult("project-1", runtime.HealthCheckDefinition{ID: "ready", ServiceID: "api", Mode: "any", Members: []string{"one", "two"}}, results, time.Now())
	if result.Status != observability.StatusHealthy || result.ProjectID != "project-1" {
		t.Fatalf("result = %#v", result)
	}
}

type fixedHealthObserver struct{ observation runtime.Observation }

func (s fixedHealthObserver) Inspect(context.Context, string) (runtime.Observation, error) {
	return s.observation, nil
}

type countingHealthEvaluator struct{ calls int }

func (e *countingHealthEvaluator) Evaluate(context.Context, runtime.ProjectRuntime, runtime.Observation, runtime.HealthCheckDefinition) observability.HealthResult {
	e.calls++
	return observability.HealthResult{Status: observability.StatusHealthy}
}

type countingHealthReads struct {
	healthRepositoryStub
	reads int
}

func (r *countingHealthReads) LatestHealth(context.Context, string) ([]observability.HealthResult, error) {
	r.reads++
	return nil, nil
}

func TestStoppedAndPausedHealthDoesNotExecuteOrPersistDiscardedProbes(t *testing.T) {
	t.Parallel()
	for _, state := range []runtime.ProjectState{runtime.StateStopped, runtime.StatePaused} {
		project := runtime.ProjectRuntime{ProjectID: "stopped", Services: []runtime.ServiceDeclaration{{ID: "api", HealthChecks: []runtime.HealthCheckDefinition{{ID: "command", ServiceID: "api", Type: "command", Command: []string{"must-not-run"}, Retries: 3, Required: true}}}}}
		observer := fixedHealthObserver{observation: runtime.Observation{State: state, ObservedAt: time.Now().UTC()}}
		repository := &countingHealthReads{}
		evaluator := &countingHealthEvaluator{}
		service := NewHealthService(healthSourceStub{project: project}, observer, repository, evaluator)
		for range 2 {
			result, err := service.EvaluateProject(context.Background(), project.ProjectID)
			if err != nil || result.Status != observability.StatusUnknown || len(result.Results) != 0 {
				t.Fatalf("stopped health: %v %v", result, err)
			}
			if _, err := service.Get(context.Background(), project.ProjectID); err != nil {
				t.Fatal(err)
			}
		}
		if evaluator.calls != 0 || len(repository.results) != 0 || repository.reads != 0 {
			t.Fatalf("discarded work still ran: probes=%d writes=%d reads=%d", evaluator.calls, len(repository.results), repository.reads)
		}
	}
	// Starting services still execute readiness probes.
	project := runtime.ProjectRuntime{ProjectID: "starting", Services: []runtime.ServiceDeclaration{{ID: "api", HealthChecks: []runtime.HealthCheckDefinition{{ID: "ready", ServiceID: "api", Type: "process"}}}}}
	evaluator := &countingHealthEvaluator{}
	repository := &healthRepositoryStub{}
	service := NewHealthService(healthSourceStub{project: project}, fixedHealthObserver{observation: runtime.Observation{State: runtime.StateStarting, ObservedAt: time.Now().UTC()}}, repository, evaluator)
	if _, err := service.EvaluateProject(context.Background(), project.ProjectID); err != nil {
		t.Fatal(err)
	}
	if evaluator.calls != 1 || len(repository.results) != 1 {
		t.Fatal("starting readiness was disabled")
	}
}
