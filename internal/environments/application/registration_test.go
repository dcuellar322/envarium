package application

import (
	"context"
	"envarium.dev/envarium/internal/environments/domain"
	"errors"
	"slices"
	"testing"
)

func TestRegistrationCoordinatorAllocatesUniqueExactPortsAndPreservesThem(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	service := newTestService(repository, []WorktreeObservation{
		{Path: "/repo/main", Branch: "main"}, {Path: "/repo/feature", Branch: "feature"},
	})
	allocator := &allocatorStub{next: 21000}
	coordinator := NewRegistrationCoordinator(service, declarationStub{items: []PortDeclaration{
		{ID: "api", Protocol: "tcp", TargetPort: 8080}, {ID: "debug", Protocol: "tcp", TargetPort: 4000},
	}}, allocator)
	first, err := coordinator.RegisterWorktrees(context.Background(), "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Environments) != 2 || allocator.calls != 4 {
		t.Fatalf("registration=%#v allocator calls=%d", first, allocator.calls)
	}
	hosts := make(map[int]string)
	for _, environment := range first.Environments {
		if environment.State != "inactive" || len(environment.Allocation.PortLeases) != 2 {
			t.Fatalf("environment = %#v", environment)
		}
		for _, lease := range environment.Allocation.PortLeases {
			if owner, exists := hosts[lease.HostPort]; exists {
				t.Fatalf("host port %d shared by %s and %s", lease.HostPort, owner, environment.ID)
			}
			hosts[lease.HostPort] = environment.ID
		}
	}
	allocator.next = 31000
	second, err := coordinator.RegisterWorktrees(context.Background(), "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if allocator.calls != 4 {
		t.Fatalf("stable compatible leases triggered new allocations: %d", allocator.calls)
	}
	for index := range first.Environments {
		if !slices.Equal(first.Environments[index].Allocation.PortLeases, second.Environments[index].Allocation.PortLeases) {
			t.Fatalf("leases changed: %#v -> %#v", first.Environments[index].Allocation.PortLeases, second.Environments[index].Allocation.PortLeases)
		}
	}
}

func TestRegistrationCoordinatorDoesNotPersistPartialAllocationFailure(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	service := newTestService(repository, []WorktreeObservation{{Path: "/repo/main", Branch: "main"}})
	coordinator := NewRegistrationCoordinator(service, declarationStub{items: []PortDeclaration{{ID: "api", Protocol: "tcp", TargetPort: 8080}}}, &allocatorStub{err: errors.New("no port")})
	if _, err := coordinator.RegisterWorktrees(context.Background(), "project-1"); err == nil {
		t.Fatal("RegisterWorktrees() succeeded")
	}
	items, err := repository.List(context.Background())
	if err != nil || len(items) != 1 || len(items[0].Allocation.PortLeases) != 0 {
		t.Fatalf("items=%#v error=%v", items, err)
	}
}

type declarationStub struct{ items []PortDeclaration }

func (s declarationStub) Declarations(context.Context, string) ([]PortDeclaration, error) {
	return slices.Clone(s.items), nil
}

type allocatorStub struct {
	next  int
	calls int
	err   error
}

func (a *allocatorStub) Suggest(_ context.Context, _, _ int, _ string, _ string, excluded []int) (int, error) {
	if a.err != nil {
		return 0, a.err
	}
	a.calls++
	for slices.Contains(excluded, a.next) {
		a.next++
	}
	result := a.next
	a.next++
	return result, nil
}

type checkoutDeclarations struct {
	declarationStub
	scopes []string
}

func (s *checkoutDeclarations) DeclarationsForEnvironment(_ context.Context, id string) ([]PortDeclaration, error) {
	s.scopes = append(s.scopes, id)
	return []PortDeclaration{{ID: id, Protocol: "tcp", TargetPort: 8080}}, nil
}
func TestRegistrationReadsEachSelectedCheckoutDeclarations(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(repository, []WorktreeObservation{{Path: "/repo/main", Branch: "main"}, {Path: "/repo/feature", Branch: "feature"}})
	source := &checkoutDeclarations{}
	coordinator := NewRegistrationCoordinator(service, source, &allocatorStub{next: 21000})
	result, err := coordinator.RegisterWorktrees(context.Background(), "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(source.scopes) != 2 {
		t.Fatal("did not resolve both checkouts")
	}
	for _, environment := range result.Environments {
		if len(environment.Allocation.PortLeases) != 1 || environment.Allocation.PortLeases[0].PortID != environment.ID {
			t.Fatal("used another checkout's declarations")
		}
	}
}

type countedRuntimeWrites struct {
	*repositoryStub
	updates int
}

func (r *countedRuntimeWrites) Update(ctx context.Context, environment domain.Environment) error {
	r.updates++
	return r.repositoryStub.Update(ctx, environment)
}
func TestStableRegistrationDoesNotRewriteRuntimeState(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(repository, []WorktreeObservation{{Path: "/repo/main", Branch: "main"}})
	counted := &countedRuntimeWrites{repositoryStub: repository}
	service.repository = counted
	coordinator := NewRegistrationCoordinator(service, declarationStub{items: []PortDeclaration{{ID: "api", Protocol: "tcp", TargetPort: 8080}}}, &allocatorStub{next: 21000})
	if _, err := coordinator.RegisterWorktrees(context.Background(), "project-1"); err != nil {
		t.Fatal(err)
	}
	first := counted.updates
	if _, err := coordinator.RegisterWorktrees(context.Background(), "project-1"); err != nil {
		t.Fatal(err)
	}
	if first != 1 || counted.updates != first {
		t.Fatal("unchanged registration wrote runtime state", counted.updates)
	}
}
