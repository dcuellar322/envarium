package process

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestNativeWorktreePlansRequireLeasesAndExplicitBindings(t *testing.T) {
	root := t.TempDir()
	definition := domain.ProcessDefinition{ID: "web", Command: []string{"server"}, WorkingDirectory: "."}
	project := processProject(root, []servicePlan{{service: domain.ServiceDeclaration{ID: "web", RuntimeName: "web"}, definition: definition}})
	project.EnvironmentID = "env-feature"
	project.Ports = map[string]domain.PortDeclaration{"http": {Service: "web", Host: 19001, Protocol: "tcp"}}
	plan := func(action domain.Action) error {
		_, err := buildPlan(domain.PlanRequest{Project: project, Action: action})
		return err
	}
	if plan(domain.ActionStart) == nil {
		t.Fatal("unallocated start accepted")
	}
	p := project.Ports["http"]
	p.Allocated = true
	project.Ports["http"] = p
	if plan(domain.ActionStart) == nil {
		t.Fatal("hardcoded start accepted")
	}
	project.Process.Processes[0].PortBindings = []string{"http"}
	if err := plan(domain.ActionStart); err != nil {
		t.Fatal(err)
	}
	project.Process.Processes[0].WorkingDirectory = filepath.Dir(root)
	if plan(domain.ActionStart) == nil {
		t.Fatal("escaped working directory accepted")
	}
	if err := plan(domain.ActionStop); err != nil {
		t.Fatalf("cleanup blocked: %v", err)
	}
}
func TestWorktreePortCollisionDoesNotTouchListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	port := listener.Addr().(*net.TCPAddr).Port
	project := domain.ProjectRuntime{EnvironmentID: "env-feature", Ports: map[string]domain.PortDeclaration{"http": {Service: "web", Host: port, Protocol: "tcp"}}}
	driver := &Driver{}
	if driver.checkWorktreePorts(context.Background(), project, servicePlan{service: domain.ServiceDeclaration{ID: "web"}}) == nil {
		t.Fatal("occupied lease accepted")
	}
	connection, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal("listener disturbed", err)
	}
	_ = connection.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if driver.checkWorktreePorts(ctx, project, servicePlan{service: domain.ServiceDeclaration{ID: "web"}}) == nil {
		t.Fatal("cancellation ignored")
	}
}
func TestWholeRuntimeStopIncludesRetiredServicesOnlyInItsScope(t *testing.T) {
	plan := executionPlan{allServices: true, services: []servicePlan{{service: domain.ServiceDeclaration{ID: "web"}}}}
	active := map[string]domain.RunRecord{"worker": {ServiceID: "worker"}}
	if got := stoppingServices(plan, active); len(got) != 2 || got[1].service.ID != "worker" {
		t.Fatalf("retired service omitted: %v", got)
	}
	plan.allServices = false
	if len(stoppingServices(plan, active)) != 1 {
		t.Fatal("selected stop broadened")
	}
}

func TestNativeWorktreesRunInParallelAndRestartIndependently(t *testing.T) {
	portA, portB := freePort(t), freePort(t)
	driver, store, a, cancel := nativeTestDriver(t, "child", portA, domain.RestartPolicy{})
	defer cancel()
	a.ProjectID = "env-native-a"
	a.EnvironmentID = a.ProjectID
	a.CatalogProjectID = "catalog"
	a.Ports = map[string]domain.PortDeclaration{"http": {ID: "http", Service: "web", Host: portA, Protocol: "tcp", Allocated: true}}
	a.Services[0].HostPorts = []int{portA}
	a.Process.Processes[0].PortBindings = []string{"http"}
	b := a
	b.ProjectID = "env-native-b"
	b.EnvironmentID = b.ProjectID
	b.Root = t.TempDir()
	definition := a.Process.Processes[0]
	definition.WorkingDirectory = b.Root
	definition.Environment = map[string]string{"ENVARIUM_PROCESS_HELPER": "1", "PORT": fmt.Sprint(portB)}
	b.Process = &domain.ProcessRuntime{Processes: []domain.ProcessDefinition{definition}}
	b.Ports = map[string]domain.PortDeclaration{"http": {ID: "http", Service: "web", Host: portB, Protocol: "tcp", Allocated: true}}
	b.Services = []domain.ServiceDeclaration{{ID: "web", RuntimeName: "web", HostPorts: []int{portB}}}
	t.Cleanup(func() {
		executeProcessAction(t, driver, a, domain.ActionStop)
		executeProcessAction(t, driver, b, domain.ActionStop)
	})
	executeProcessAction(t, driver, a, domain.ActionStart)
	waitForTCP(t, portA)
	executeProcessAction(t, driver, b, domain.ActionStart)
	waitForTCP(t, portB)
	before := waitForProcessState(t, driver, b, domain.StateRunning).Services[0].Process.RunID
	executeProcessAction(t, driver, a, domain.ActionRestart)
	waitForTCP(t, portA)
	if got := waitForProcessState(t, driver, b, domain.StateRunning).Services[0].Process.RunID; got != before {
		t.Fatal("restarted the other worktree")
	}
	for _, project := range []domain.ProjectRuntime{a, b} {
		observation, err := driver.InspectBindings(context.Background(), project)
		if err != nil || len(observation.Services[0].Ports) != 1 {
			t.Fatalf("verified worktree binding missing: %#v %v", observation, err)
		}
		runs, err := store.ListProjectRuns(context.Background(), project.ProjectID)
		if err != nil || runs[0].CatalogProjectID != "catalog" {
			t.Fatalf("catalog owner lost: %#v %v", runs, err)
		}
	}
	// A new daemon instance recovers ownership from durable fingerprints.
	recovered := newDriver(context.Background(), store, gopsutilInspector{}, &secretResolverFake{})
	executeProcessAction(t, recovered, a, domain.ActionStop)
	if got := waitForProcessState(t, driver, b, domain.StateRunning).Services[0].Process.RunID; got != before {
		t.Fatal("recovered stop crossed scope")
	}
	executeProcessAction(t, driver, b, domain.ActionStop)
}

func TestRemovedNativeServiceRemainsVisibleAndBlocksStoppedCheckout(t *testing.T) {
	ctx := context.Background()
	store := newMemoryRunStore()
	now := time.Now().UTC()
	identity := domain.ProcessIdentity{PID: 123, ProcessGroup: 123, Fingerprint: "owned", StartedAt: now, ObservedAt: now}
	run := domain.RunRecord{ID: "retired", ProjectID: "env-f", ServiceID: "old", StartedAt: now, Processes: []domain.ProcessIdentity{identity}}
	if err := store.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	driver := newDriver(ctx, store, inspectorFake{snapshots: map[int32]domain.ProcessIdentity{123: identity}, groups: map[int32][]domain.ProcessIdentity{123: {identity}}}, &secretResolverFake{})
	project := domain.ProjectRuntime{ProjectID: "env-f", Kind: domain.KindProcess, Process: &domain.ProcessRuntime{}}
	observation, err := driver.Inspect(ctx, project)
	if err != nil || observation.State == domain.StateStopped || len(observation.Services) != 1 || observation.Services[0].Process.RunID != "retired" {
		t.Fatalf("retired ownership hidden: %#v %v", observation, err)
	}
}
