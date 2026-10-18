package process

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"errors"
	gnet "github.com/shirou/gopsutil/v4/net"
	"os"
	"testing"
	"time"
)

func TestBindingsRequireCurrentSocketOwnerAndPreserveAddresses(t *testing.T) {
	t.Parallel()
	inspector := gopsutilInspector{}
	owner, err := inspector.Snapshot(context.Background(), int32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	inspector.connections = func(context.Context) ([]gnet.ConnectionStat, error) {
		return []gnet.ConnectionStat{
			{Status: "LISTEN", Pid: owner.PID, Laddr: gnet.Addr{IP: "::1", Port: 3000}},
			{Status: "LISTEN", Pid: owner.PID, Laddr: gnet.Addr{IP: "127.0.0.1", Port: 3000}},
			{Status: "LISTEN", Pid: owner.PID, Laddr: gnet.Addr{IP: "127.0.0.1", Port: 3000}},
			{Status: "ESTABLISHED", Pid: owner.PID, Laddr: gnet.Addr{IP: "127.0.0.1", Port: 3001}},
			{Status: "LISTEN", Pid: 2147483647, Laddr: gnet.Addr{IP: "0.0.0.0", Port: 3001}},
		}, nil
	}
	ports := []domain.PublishedPort{{HostPort: 3000, ContainerPort: 3000, Protocol: "tcp"}, {HostPort: 3001, ContainerPort: 3001, Protocol: "tcp"}}
	got, err := inspector.Bindings(context.Background(), ports, []domain.ProcessIdentity{owner})
	if err != nil || len(got) != 2 || got[0].HostIP != "127.0.0.1" || got[1].HostIP != "::1" {
		t.Fatalf("bindings=%v err=%v", got, err)
	}
	got, err = inspector.Bindings(context.Background(), ports, []domain.ProcessIdentity{{PID: 2147483647, Fingerprint: "unrelated"}})
	if err != nil || len(got) != 0 {
		t.Fatal("unrelated socket attributed", got, err)
	}
	failure := errors.New("socket observation unavailable")
	inspector.connections = func(context.Context) ([]gnet.ConnectionStat, error) { return nil, failure }
	if _, err := inspector.Bindings(context.Background(), ports, []domain.ProcessIdentity{owner}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := inspector.Bindings(ctx, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestExternalServiceDoesNotClaimItsOtherUnboundOrUnrelatedPorts(t *testing.T) {
	t.Parallel()
	owner := domain.ProcessIdentity{PID: 99, ProcessGroup: 99, Executable: "/usr/bin/uv", Fingerprint: "owner", StartedAt: time.Now()}
	other := domain.ProcessIdentity{PID: 100, ProcessGroup: 99, Executable: "/usr/bin/uv", Fingerprint: "other", StartedAt: time.Now()}
	inspector := inspectorFake{listeners: map[int][]domain.ProcessIdentity{3000: {owner}, 3001: {other}}}
	driver := newDriver(context.Background(), newMemoryRunStore(), inspector, &secretResolverFake{})
	project := processProject("/repo", []servicePlan{{service: domain.ServiceDeclaration{ID: "api", RuntimeName: "api", HostPorts: []int{3000, 3001, 3002}}, definition: domain.ProcessDefinition{ID: "api", Command: []string{"uv", "run", "api.py"}}}})
	observation, err := driver.InspectBindings(context.Background(), project)
	if err != nil || len(observation.Services[0].Ports) != 1 || observation.Services[0].Ports[0].HostPort != 3000 {
		t.Fatalf("observation=%#v err=%v", observation, err)
	}
}

type unavailableBindingsInspector struct{ inspectorFake }

func (unavailableBindingsInspector) Bindings(context.Context, []domain.PublishedPort, []domain.ProcessIdentity) ([]domain.PublishedPort, error) {
	return nil, errors.New("socket access denied")
}
func TestSocketAttributionFailureDoesNotDegradeOrdinaryMonitoring(t *testing.T) {
	t.Parallel()
	owner := domain.ProcessIdentity{PID: 99, ProcessGroup: 99, Executable: "/usr/bin/uv", Fingerprint: "owner", StartedAt: time.Now()}
	inspector := unavailableBindingsInspector{inspectorFake{listeners: map[int][]domain.ProcessIdentity{3000: {owner}}}}
	driver := newDriver(context.Background(), newMemoryRunStore(), inspector, &secretResolverFake{})
	project := processProject("/repo", []servicePlan{{service: domain.ServiceDeclaration{ID: "api", RuntimeName: "api", HostPorts: []int{3000}}, definition: domain.ProcessDefinition{ID: "api", Command: []string{"uv", "run", "api.py"}}}})
	observation, err := driver.Inspect(context.Background(), project)
	if err != nil || observation.State != domain.StateRunningExternal {
		t.Fatalf("ordinary monitoring depends on port access: %v %v", observation, err)
	}
	if _, err := driver.InspectBindings(context.Background(), project); err == nil {
		t.Fatal("socket observation failure hidden")
	}
}
