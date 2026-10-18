package application

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"errors"
	"testing"
	"time"
)

type leaseDriver struct {
	*driverStub
	observation domain.Observation
}

func (d *leaseDriver) Inspect(context.Context, domain.ProjectRuntime) (domain.Observation, error) {
	return d.observation, nil
}
func TestNativeAllocationUsesFreshOwnershipAndSerializesLifecycle(t *testing.T) {
	projects := projectSourceStub{projects: map[string]domain.ProjectRuntime{"env-f": {ProjectID: "env-f", Root: "/repo/f", Kind: domain.KindProcess}}}
	driver := &leaseDriver{driverStub: &driverStub{kind: domain.KindProcess}}
	service := NewService(projects, driver)
	if _, err := service.Inspect(context.Background(), "env-f"); err != nil {
		t.Fatal(err)
	}
	driver.observation = domain.Observation{Services: []domain.ServiceObservation{{Process: &domain.ProcessMetadata{RunID: "owned"}}}}
	called := false
	if err := service.WithIdleNativeRuntime(context.Background(), "env-f", func() error { called = true; return nil }); err == nil || called {
		t.Fatal("cached idle state authorized lease change")
	}
	driver.observation = domain.Observation{State: domain.StateRunningExternal, Services: []domain.ServiceObservation{{Process: &domain.ProcessMetadata{PID: 123}}}}
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- service.WithIdleNativeRuntime(context.Background(), "env-f", func() error { close(entered); <-release; return nil })
	}()
	<-entered
	plan, err := service.Plan(context.Background(), "env-f", domain.ActionStart, false)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := service.Execute(ctx, plan, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("lifecycle bypassed allocation gate", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	cancelled, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	if err := service.WithIdleNativeRuntime(cancelled, "env-f", func() error { called = true; return nil }); !errors.Is(err, context.Canceled) || called {
		t.Fatal("cancelled lease update executed", err)
	}
}
