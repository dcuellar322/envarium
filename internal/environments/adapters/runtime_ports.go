package adapters

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"envarium.dev/envarium/internal/environments/application"
	"envarium.dev/envarium/internal/environments/domain"
	runtimeDomain "envarium.dev/envarium/internal/runtime/domain"
)

type nativeAllocationGate interface {
	WithIdleNativeRuntime(context.Context, string, func() error) error
}

type runtimePortResolver interface {
	ResolveRuntime(context.Context, string) (runtimeDomain.ProjectRuntime, error)
}

// RuntimePortDeclarations reads the selected checkout, not another branch's ports.
type RuntimePortDeclarations struct {
	projects runtimePortResolver
	runtime  nativeAllocationGate
}

// NewRuntimePortDeclarations reads current checkout manifests and owned runtime state.
func NewRuntimePortDeclarations(projects runtimePortResolver, runtime nativeAllocationGate) *RuntimePortDeclarations {
	return &RuntimePortDeclarations{projects: projects, runtime: runtime}
}

// DeclarationsForEnvironment reads the manifest in the registered checkout.
func (s *RuntimePortDeclarations) DeclarationsForEnvironment(ctx context.Context, id string) ([]application.PortDeclaration, error) {
	return s.Declarations(ctx, id)
}

// Declarations returns exact logical port definitions from the selected scope.
func (s *RuntimePortDeclarations) Declarations(ctx context.Context, id string) ([]application.PortDeclaration, error) {
	project, err := s.projects.ResolveRuntime(ctx, id)
	if err != nil {
		return nil, err
	}
	result := []application.PortDeclaration{}
	for id, port := range project.Ports {
		target := port.Target
		if target == 0 {
			target = port.Host
		}
		result = append(result, application.PortDeclaration{ID: id, Protocol: port.Protocol, TargetPort: target})
	}
	return result, nil
}

// WithLeaseChange holds the runtime's root gate through the allocation write.
func (s *RuntimePortDeclarations) WithLeaseChange(ctx context.Context, environment domain.Environment, leases []domain.PortLease, change func() error) error {
	if sameLeases(environment.Allocation.PortLeases, leases) {
		return change()
	}
	if s.runtime == nil {
		return fmt.Errorf("runtime allocation gate is unavailable")
	}
	return s.runtime.WithIdleNativeRuntime(ctx, environment.ID, change)
}

func sameLeases(previous, current []domain.PortLease) bool {
	previous = slices.Clone(previous)
	current = slices.Clone(current)
	sort.Slice(previous, func(i, j int) bool { return previous[i].PortID < previous[j].PortID })
	sort.Slice(current, func(i, j int) bool { return current[i].PortID < current[j].PortID })
	return slices.Equal(previous, current)
}
