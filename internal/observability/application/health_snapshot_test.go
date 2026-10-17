package application

import (
	"context"
	"errors"
	"testing"

	runtime "envarium.dev/envarium/internal/runtime/domain"
)

type batchedHealthSource struct {
	RuntimeSource
	calls int
}

func (s *batchedHealthSource) ResolveRuntimeProjects(context.Context) (map[string]runtime.ProjectRuntime, map[string]error, error) {
	s.calls++
	return map[string]runtime.ProjectRuntime{"good": {ProjectID: "good"}}, map[string]error{"bad": errors.New("invalid source")}, nil
}

func TestHealthSchedulerUsesInjectedBatchSourceAndPreservesFailures(t *testing.T) {
	t.Parallel()
	source := &batchedHealthSource{}
	projects, failures, err := healthProjects(context.Background(), source)
	if err != nil || source.calls != 1 || projects["good"].ProjectID != "good" || failures["bad"] == nil {
		t.Fatal("batch source or per-project failure lost")
	}
	ids := healthProjectIDs(projects, failures)
	if len(ids) != 2 || ids[0] != "bad" || ids[1] != "good" {
		t.Fatal("unstable snapshot membership")
	}
}

func TestHealthSchedulerSkipsEmptyDeclarationsAndDetectsNewChecks(t *testing.T) {
	t.Parallel()
	source := &healthSourceStub{project: runtime.ProjectRuntime{ProjectID: "project-1", ManifestHash: "empty", Services: []runtime.ServiceDeclaration{{ID: "api"}}}}
	observer := &healthObserverStub{}
	repository := &healthRepositoryStub{}
	service := NewHealthService(source, observer, repository, observationEvaluator{})
	report := func(_ string, err error) { t.Errorf("scheduler: %v", err) }
	for range 3 {
		service.runDue(context.Background(), report)
	}
	if observer.calls != 0 || len(service.nextDue) != 0 {
		t.Fatal("empty probe declaration triggered a runtime read")
	}
	source.project.ManifestHash = "checks"
	source.project.Services[0].HealthChecks = []runtime.HealthCheckDefinition{{ID: "alive", ServiceID: "api", Type: "process"}}
	service.runDue(context.Background(), report)
	if observer.calls != 1 || len(repository.results) != 1 {
		t.Fatal("new check was not detected and evaluated immediately")
	}
	source.project.ManifestHash = "removed"
	source.project.Services[0].HealthChecks = nil
	service.runDue(context.Background(), report)
	if observer.calls != 1 || len(service.nextDue) != 0 {
		t.Fatal("removed checks retained a schedule")
	}
	if _, err := service.Get(context.Background(), "project-1"); err != nil {
		t.Fatal(err)
	}
	if observer.calls != 2 {
		t.Fatal("direct health read stopped inspecting built-in runtime health")
	}
}
