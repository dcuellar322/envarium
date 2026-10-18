package adapters

import (
	"context"
	"testing"
	"time"

	"envarium.dev/envarium/internal/environments/application"
	"envarium.dev/envarium/internal/environments/domain"
	runtimeDomain "envarium.dev/envarium/internal/runtime/domain"
)

func TestLifecyclePublishesOnlyObservedExactLeaseAndClearsItOnStop(t *testing.T) {
	t.Parallel()
	repository := NewMemoryRepository()
	environment := validLifecycleEnvironment()
	if err := repository.ReplaceProject(context.Background(), environment.ProjectID, []domain.Environment{environment}); err != nil {
		t.Fatal(err)
	}
	service := application.NewService(nil, nil, repository)
	refreshes := 0
	lifecycle := NewLifecycle(service, runtimeObserverStub{observation: runtimeDomain.Observation{
		ProjectID: environment.ID,
		Services: []runtimeDomain.ServiceObservation{{Ports: []runtimeDomain.PublishedPort{
			{HostPort: 22080, ContainerPort: 8080, Protocol: "tcp"},
		}}},
	}}, func(context.Context) error { refreshes++; return nil })
	if err := lifecycle.Started(context.Background(), environment.ID); err != nil {
		t.Fatal(err)
	}
	active, _ := service.Get(context.Background(), environment.ID)
	if active.State != domain.StateActive || active.Target != "http://127.0.0.1:22080" || refreshes != 1 {
		t.Fatalf("active=%#v refreshes=%d", active, refreshes)
	}
	if err := lifecycle.Stopped(context.Background(), environment.ID); err != nil {
		t.Fatal(err)
	}
	stopped, _ := service.Get(context.Background(), environment.ID)
	if stopped.State != domain.StateInactive || stopped.Target != "" || refreshes != 2 {
		t.Fatalf("stopped=%#v refreshes=%d", stopped, refreshes)
	}
}

func TestLifecycleLeavesNonHTTPEnvironmentUnavailableWithoutFabricatedRoute(t *testing.T) {
	t.Parallel()
	repository := NewMemoryRepository()
	environment := validLifecycleEnvironment()
	if err := repository.ReplaceProject(context.Background(), environment.ProjectID, []domain.Environment{environment}); err != nil {
		t.Fatal(err)
	}
	service := application.NewService(nil, nil, repository)
	lifecycle := NewLifecycle(service, runtimeObserverStub{}, nil)
	if err := lifecycle.Started(context.Background(), environment.ID); err != nil {
		t.Fatal(err)
	}
	current, _ := service.Get(context.Background(), environment.ID)
	if current.State != domain.StateInactive || current.Target != "" {
		t.Fatalf("environment = %#v", current)
	}
}

type runtimeObserverStub struct{ observation runtimeDomain.Observation }

func (s runtimeObserverStub) Inspect(context.Context, string) (runtimeDomain.Observation, error) {
	return s.observation, nil
}

func validLifecycleEnvironment() domain.Environment {
	now := time.Unix(10, 0).UTC()
	return domain.Environment{
		ID: "env-0123456789", ProjectID: "project-1", Name: "feature", Path: "/repo/feature",
		Availability: domain.AvailabilityAvailable, State: domain.StateInactive, Hostname: "project-abc.localhost",
		Allocation: domain.RuntimeAllocation{
			ComposeProjectName: "sy-project-0123456789abcdef", PortLeaseNamespace: "worktree:0123456789", PortOffset: 10,
			PortLeases: []domain.PortLease{{PortID: "web", Protocol: "tcp", TargetPort: 8080, HostPort: 22080}},
		},
		RegisteredAt: now, LastObservedAt: now, UpdatedAt: now,
	}
}

type bindingObserverStub struct {
	runtimeObserverStub
	bindings runtimeDomain.Observation
}

func (s bindingObserverStub) InspectBindings(context.Context, string) (runtimeDomain.Observation, error) {
	return s.bindings, nil
}
func TestNativeLifecycleRequiresActualOwnedSocketBeforePublishingRoute(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryRepository()
	environment := validLifecycleEnvironment()
	if err := repository.ReplaceProject(ctx, environment.ProjectID, []domain.Environment{environment}); err != nil {
		t.Fatal(err)
	}
	service := application.NewService(nil, nil, repository)
	declared := runtimeDomain.Observation{Driver: runtimeDomain.KindProcess, Services: []runtimeDomain.ServiceObservation{{Ports: []runtimeDomain.PublishedPort{{HostPort: 22080, Protocol: "tcp"}}}}}
	if err := NewLifecycle(service, runtimeObserverStub{observation: declared}, nil).Started(ctx, environment.ID); err == nil {
		t.Fatal("unverified native bindings accepted")
	}
	lifecycle := NewLifecycle(service, bindingObserverStub{runtimeObserverStub: runtimeObserverStub{observation: declared}}, nil)
	if err := lifecycle.Started(ctx, environment.ID); err != nil {
		t.Fatal(err)
	}
	current, _ := service.Get(ctx, environment.ID)
	if current.Target != "" {
		t.Fatal("declared port fabricated route")
	}
	lifecycle = NewLifecycle(service, bindingObserverStub{runtimeObserverStub: runtimeObserverStub{observation: declared}, bindings: declared}, nil)
	if err := lifecycle.Started(ctx, environment.ID); err != nil {
		t.Fatal(err)
	}
	current, _ = service.Get(ctx, environment.ID)
	if current.Target != "http://127.0.0.1:22080" {
		t.Fatal("verified route missing")
	}
}
