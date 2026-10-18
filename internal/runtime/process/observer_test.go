package process

import (
	"context"
	"errors"
	"testing"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestInspectRejectsReusedPIDFingerprint(t *testing.T) {
	t.Parallel()
	store := newMemoryRunStore()
	started := time.Now().Add(-time.Minute).UTC()
	stored := domain.ProcessIdentity{
		RunID: "run-stale", PID: 42, ProcessGroup: 42, Executable: "/usr/bin/uv",
		WorkingDirectory: "/repo", StartedAt: started, Fingerprint: "original", ObservedAt: started,
	}
	if err := store.CreateRun(context.Background(), domain.RunRecord{
		ID: "run-stale", ProjectID: "project-process", ServiceID: "api", RuntimeDriver: domain.KindProcess,
		Origin: domain.OriginEnvarium, StartedAt: started, IdentityFingerprint: stored.Fingerprint,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordProcess(context.Background(), stored); err != nil {
		t.Fatal(err)
	}
	inspector := inspectorFake{snapshots: map[int32]domain.ProcessIdentity{42: {
		PID: 42, ProcessGroup: 42, Executable: "/usr/bin/uv", WorkingDirectory: "/repo",
		StartedAt: started.Add(30 * time.Second), Fingerprint: "reused",
	}}}
	driver := newDriver(context.Background(), store, inspector, &secretResolverFake{})
	project := processProject("/repo", []servicePlan{{
		service:    domain.ServiceDeclaration{ID: "api", RuntimeName: "api"},
		definition: domain.ProcessDefinition{ID: "api", Command: []string{"uv", "run", "api.py"}},
	}})
	observation, err := driver.Inspect(context.Background(), project)
	if err != nil {
		t.Fatal(err)
	}
	if observation.State != domain.StateUnknown || observation.Services[0].State != "stale" {
		t.Fatalf("observation = %#v", observation)
	}
	runs, _ := store.ListProjectRuns(context.Background(), project.ProjectID)
	if runs[0].EndedAt == nil || runs[0].TerminationReason != "identity_lost" {
		t.Fatalf("run = %#v", runs[0])
	}
}

func TestInspectLabelsMatchingExternalListenerHonestly(t *testing.T) {
	t.Parallel()
	identity := domain.ProcessIdentity{
		PID: 99, ProcessGroup: 99, Executable: "/opt/homebrew/bin/uv", WorkingDirectory: "/repo",
		StartedAt: time.Now().UTC(), Fingerprint: "external",
	}
	inspector := inspectorFake{listeners: map[int][]domain.ProcessIdentity{18082: {identity}}}
	driver := newDriver(context.Background(), newMemoryRunStore(), inspector, &secretResolverFake{})
	project := processProject("/repo", []servicePlan{{
		service:    domain.ServiceDeclaration{ID: "api", RuntimeName: "api", HostPorts: []int{18082}},
		definition: domain.ProcessDefinition{ID: "api", Command: []string{"uv", "run", "api.py"}},
	}})
	observation, err := driver.Inspect(context.Background(), project)
	if err != nil {
		t.Fatal(err)
	}
	if observation.State != domain.StateRunningExternal || observation.Origin != domain.OriginExternal || observation.Services[0].Process.RunID != "" {
		t.Fatalf("observation = %#v", observation)
	}
}

func TestStartRechecksListenersAfterCachedStoppedObservation(t *testing.T) {
	t.Parallel()
	service := servicePlan{
		service:    domain.ServiceDeclaration{ID: "api", RuntimeName: "api", HostPorts: []int{18082}},
		definition: domain.ProcessDefinition{ID: "api", Command: []string{"uv", "run", "api.py"}},
	}
	project := processProject("/repo", []servicePlan{service})
	fresh := inspectorFake{listeners: map[int][]domain.ProcessIdentity{18082: {{PID: 99, Executable: "/usr/bin/uv", Fingerprint: "external"}}}}
	driver := newDriver(context.Background(), newMemoryRunStore(), fresh, &secretResolverFake{})
	driver.observedInspector = inspectorFake{}
	observation, err := driver.Inspect(context.Background(), project)
	if err != nil || observation.State != domain.StateStopped {
		t.Fatalf("observation: %v %v", observation, err)
	}
	if _, _, err := driver.startService(context.Background(), project, service); !errors.Is(err, ErrExternalProcess) {
		t.Fatalf("start used stale empty inventory: %v", err)
	}
	runs, _ := driver.store.ListProjectRuns(context.Background(), project.ProjectID)
	if len(runs) != 0 {
		t.Fatal("external process was adopted")
	}
}

func TestCachedListenerPIDsStillRequireFreshIdentity(t *testing.T) {
	t.Parallel()
	inspector := gopsutilInspector{connections: func(context.Context) ([]gnet.ConnectionStat, error) {
		return []gnet.ConnectionStat{{Status: "LISTEN", Pid: 2147483647, Laddr: gnet.Addr{Port: 3000}}}, nil
	}}
	values, err := inspector.Listeners(context.Background(), 3000)
	if err != nil || len(values) != 0 {
		t.Fatalf("stale PID treated as live identity: %v %v", values, err)
	}
}
