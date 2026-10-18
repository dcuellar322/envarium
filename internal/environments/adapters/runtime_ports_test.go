package adapters

import (
	"context"
	"envarium.dev/envarium/internal/environments/domain"
	"errors"
	"testing"
)

type allocationGateStub struct {
	calls int
	err   error
}

func (g *allocationGateStub) WithIdleNativeRuntime(_ context.Context, _ string, change func() error) error {
	g.calls++
	if g.err != nil {
		return g.err
	}
	return change()
}
func TestNativeLeaseWriteRunsInsideOwnershipGate(t *testing.T) {
	environment := validLifecycleEnvironment()
	gate := &allocationGateStub{}
	source := NewRuntimePortDeclarations(nil, gate)
	called := false
	change := func() error { called = true; return nil }
	if err := source.WithLeaseChange(context.Background(), environment, environment.Allocation.PortLeases, change); err != nil || !called || gate.calls != 0 {
		t.Fatal("stable leases needlessly gated", err)
	}
	changed := []domain.PortLease{{PortID: "web", HostPort: 22081, TargetPort: 8080, Protocol: "tcp"}}
	called = false
	gate.err = errors.New("owned process active")
	if err := source.WithLeaseChange(context.Background(), environment, changed, change); err == nil || called || gate.calls != 1 {
		t.Fatal("lease write bypassed ownership gate", err)
	}
	gate.err = nil
	if err := source.WithLeaseChange(context.Background(), environment, changed, change); err != nil || !called {
		t.Fatal("idle allocation failed", err)
	}
}
