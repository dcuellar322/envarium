package adapters

import (
	"context"
	catalog "envarium.dev/envarium/internal/catalog/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
	"testing"
)

type portCatalog struct{}

func (portCatalog) ListProjects(context.Context) ([]catalog.Project, error) {
	return []catalog.Project{{ID: "p", DisplayName: "App", TrustState: catalog.TrustTrusted}}, nil
}

type portRuntime struct{}

func (portRuntime) Inspect(context.Context, string) (runtime.Observation, error) {
	services := []runtime.ServiceObservation{}
	for _, state := range []string{"running", "paused", "restarting", "stopped", "created", "failed", "unknown"} {
		services = append(services, runtime.ServiceObservation{ID: state, State: state, Ports: []runtime.PublishedPort{{HostPort: 1234, ContainerPort: 80, Protocol: "tcp"}}})
	}
	return runtime.Observation{Driver: runtime.KindProcess, Services: services}, nil
}
func TestRuntimeBindingsExcludeStoppedDeclarations(t *testing.T) {
	t.Parallel()
	facts, err := NewRuntimeBindings(portCatalog{}, portRuntime{}).Facts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 3 {
		t.Fatalf("facts=%#v", facts)
	}
	for _, fact := range facts {
		if fact.ServiceID != "running" && fact.ServiceID != "paused" && fact.ServiceID != "restarting" {
			t.Fatalf("inactive service presented as bound: %#v", fact)
		}
	}
}
