package application

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

type checkoutDriver struct {
	Driver
	states     map[string]domain.ProjectState
	executions atomic.Int32
}

func (*checkoutDriver) Kind() domain.Kind { return domain.KindCompose }
func (d *checkoutDriver) Inspect(_ context.Context, project domain.ProjectRuntime) (domain.Observation, error) {
	return domain.Observation{State: d.states[project.ProjectID]}, nil
}
func (d *checkoutDriver) Plan(_ context.Context, request domain.PlanRequest) (domain.Plan, error) {
	return domain.Plan{ProjectID: request.Project.ProjectID, Driver: domain.KindCompose, Action: request.Action}, nil
}
func (d *checkoutDriver) Execute(context.Context, domain.Plan, domain.ProgressSink) error {
	d.executions.Add(1)
	return nil
}

func TestCheckoutGateExcludesSharedRootLifecycleAndRejectsStalePlan(t *testing.T) {
	t.Parallel()
	projects := map[string]domain.ProjectRuntime{"p": {ProjectID: "p", Root: "/repo", ManifestHash: "one", Kind: domain.KindCompose}, "env-primary": {ProjectID: "env-primary", Root: "/repo", ManifestHash: "one", Kind: domain.KindCompose}}
	driver := &checkoutDriver{states: map[string]domain.ProjectState{"p": domain.StateStopped, "env-primary": domain.StateStopped}}
	service := NewService(projectSourceStub{projects: projects}, driver)
	plan, err := service.Plan(context.Background(), "env-primary", domain.ActionStart, false)
	if err != nil {
		t.Fatal(err)
	}
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- service.WithStoppedProject(context.Background(), "p", func() error { close(entered); <-release; return nil })
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("checkout gate did not enter")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := service.Execute(ctx, plan, nil); !errors.Is(err, context.DeadlineExceeded) || driver.executions.Load() != 0 {
		t.Fatalf("shared-root execution bypassed gate: %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := service.Execute(context.Background(), plan, nil); err != nil || driver.executions.Load() != 1 {
		t.Fatalf("gate not released: %v", err)
	}
	changed := projects["env-primary"]
	changed.ManifestHash = "new-checkout"
	projects["env-primary"] = changed
	if err := service.Execute(context.Background(), plan, nil); !errors.Is(err, ErrPlanStale) || driver.executions.Load() != 1 {
		t.Fatalf("stale plan executed: %v", err)
	}
}
func TestCheckoutGateChecksPrimaryEnvironmentWithFreshDriverObservation(t *testing.T) {
	t.Parallel()
	projects := map[string]domain.ProjectRuntime{"p": {ProjectID: "p", Root: "/repo", Kind: domain.KindCompose}, "env-primary": {ProjectID: "env-primary", Root: "/repo", Kind: domain.KindCompose}}
	driver := &checkoutDriver{states: map[string]domain.ProjectState{"p": domain.StateStopped, "env-primary": domain.StateStopped}}
	service := NewService(projectSourceStub{projects: projects}, driver)
	if _, err := service.Inspect(context.Background(), "env-primary"); err != nil {
		t.Fatal(err)
	}
	driver.states["env-primary"] = domain.StateRunning
	called := false
	err := service.WithStoppedProject(context.Background(), "p", func() error { called = true; return nil })
	if !errors.Is(err, ErrCheckoutActive) || called {
		t.Fatalf("cached stopped state authorized checkout: %v", err)
	}
}

func TestManagedCheckoutMutationInvalidatesUnchangedManifestReview(t *testing.T) {
	t.Parallel()
	projects := map[string]domain.ProjectRuntime{"p": {ProjectID: "p", Root: "/repo", ManifestHash: "same", Kind: domain.KindCompose}}
	driver := &checkoutDriver{states: map[string]domain.ProjectState{"p": domain.StateStopped}}
	service := NewService(projectSourceStub{projects: projects}, driver)
	old, err := service.Plan(context.Background(), "p", domain.ActionStart, false)
	if err != nil {
		t.Fatal(err)
	}
	partial := errors.New("checkout failed after modifying files")
	if err := service.WithCheckoutMutation(context.Background(), "p", func() error { return partial }); !errors.Is(err, partial) {
		t.Fatal(err)
	}
	current, err := service.Plan(context.Background(), "p", domain.ActionStart, false)
	if err != nil {
		t.Fatal(err)
	}
	if current.SourceManifestHash != old.SourceManifestHash || current.SourceCheckoutRevision == old.SourceCheckoutRevision {
		t.Fatal("checkout review not independently invalidated")
	}
	if err := CheckReview(current, old.SourceCheckoutRevision, old.SourceManifestHash); !errors.Is(err, ErrPlanStale) {
		t.Fatal("old confirmation accepted")
	}
	if err := service.Execute(context.Background(), old, nil); !errors.Is(err, ErrPlanStale) || driver.executions.Load() != 0 {
		t.Fatal("old queued plan executed")
	}
	if err := CheckReview(current, current.SourceCheckoutRevision, current.SourceManifestHash); err != nil {
		t.Fatal(err)
	}
}
func TestCheckoutRevisionsBoundMemoryAndFailClosedOnIdentityFailure(t *testing.T) {
	t.Parallel()
	gate := &checkoutGate{}
	old, err := gate.revision("active")
	if err != nil {
		t.Fatal(err)
	}
	release, err := gate.acquire(context.Background(), "active")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 256; i++ {
		if _, err := gate.revision(string(rune(i))); err != nil {
			t.Fatal(err)
		}
	}
	current, err := gate.revision("active")
	if err != nil || current != old || len(gate.entries) > 128 {
		t.Fatal("live root evicted or unbounded")
	}
	release()
	failure := errors.New("entropy unavailable")
	gate.newRevision = func() (string, error) { return "", failure }
	if err := gate.changed("active"); !errors.Is(err, failure) {
		t.Fatal("identity failure hidden")
	}
	if _, err := gate.revision("active"); !errors.Is(err, failure) {
		t.Fatal("old revision survived failed invalidation")
	}
}
func TestNativeEnvironmentCanPlanThroughSharedLifecycleGate(t *testing.T) {
	t.Parallel()
	project := domain.ProjectRuntime{ProjectID: "env-feature", EnvironmentID: "env-feature", Root: "/worktree", Kind: domain.KindProcess, Ports: map[string]domain.PortDeclaration{}}
	driver := &driverStub{kind: domain.KindProcess}
	service := NewService(projectSourceStub{projects: map[string]domain.ProjectRuntime{project.ProjectID: project}}, driver)
	if _, err := service.Plan(context.Background(), project.ProjectID, domain.ActionStart, false); err != nil {
		t.Fatalf("native environment planning failed: %v", err)
	}
	if _, err := service.Plan(context.Background(), project.ProjectID, domain.ActionStop, false); err != nil {
		t.Fatal("existing environment cannot be stopped", err)
	}
}
