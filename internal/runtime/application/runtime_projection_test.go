package application

import (
	"strings"
	"testing"

	catalog "envarium.dev/envarium/internal/catalog/domain"
	"envarium.dev/envarium/internal/runtime/domain"
)

func TestRuntimeProjectionCopiesInputsAndChecksIdentity(t *testing.T) {
	t.Parallel()
	value := domain.ProjectRuntime{ProjectID: "one", ProjectSlug: "one", Root: "/one", ManifestHash: "first",
		Compose:  &domain.ComposeRuntime{Files: []string{"compose.yml"}, Profiles: []string{"web"}, PortOverrides: map[string]int{"web": 80}},
		Process:  &domain.ProcessRuntime{Environment: map[string]string{"A": "B"}, Secrets: map[string]domain.SecretReference{"s": {Key: "key"}}, Processes: []domain.ProcessDefinition{{Command: []string{"server"}, Environment: map[string]string{"A": "B"}, Secrets: map[string]domain.SecretReference{"s": {Key: "key"}}}}},
		Ports:    map[string]domain.PortDeclaration{"web": {Host: 80}},
		Services: []domain.ServiceDeclaration{{Dependencies: []string{"db"}, HostPorts: []int{80}, HealthChecks: []domain.HealthCheckDefinition{{Command: []string{"check"}, Members: []string{"web"}}}}},
	}
	var cache runtimeProjections
	cache.store(value)
	project := catalog.Project{ID: "one", Slug: "one", PrimaryLocation: "/one"}
	first, ok := cache.load(project, "first")
	if !ok {
		t.Fatal("missing projection")
	}
	first.Compose.Files[0] = "changed"
	first.Compose.Profiles[0] = "changed"
	first.Compose.PortOverrides["web"] = 90
	first.Process.Environment["A"] = "changed"
	first.Process.Secrets["s"] = domain.SecretReference{}
	first.Process.Processes[0].Command[0] = "changed"
	first.Process.Processes[0].Environment["A"] = "changed"
	first.Process.Processes[0].Secrets["s"] = domain.SecretReference{}
	first.Ports["web"] = domain.PortDeclaration{}
	first.Services[0].Dependencies[0] = "changed"
	first.Services[0].HostPorts[0] = 90
	first.Services[0].HealthChecks[0].Command[0] = "changed"
	first.Services[0].HealthChecks[0].Members[0] = "changed"
	next, ok := cache.load(project, "first")
	if !ok || next.Compose.Files[0] != "compose.yml" || next.Compose.Profiles[0] != "web" || next.Compose.PortOverrides["web"] != 80 || next.Process.Environment["A"] != "B" || next.Process.Secrets["s"].Key != "key" || next.Process.Processes[0].Command[0] != "server" || next.Process.Processes[0].Environment["A"] != "B" || next.Process.Processes[0].Secrets["s"].Key != "key" || next.Ports["web"].Host != 80 || next.Services[0].Dependencies[0] != "db" || next.Services[0].HostPorts[0] != 80 || next.Services[0].HealthChecks[0].Command[0] != "check" || next.Services[0].HealthChecks[0].Members[0] != "web" {
		t.Fatal("caller mutation leaked")
	}
	if _, ok := cache.load(project, "other"); ok {
		t.Fatal("stale declaration")
	}
	project.PrimaryLocation = "/moved"
	if _, ok := cache.load(project, "first"); ok {
		t.Fatal("stale location")
	}
	value.Compose.Files[0] = "changed"
	if next, _ := cache.load(catalog.Project{ID: "one", Slug: "one", PrimaryLocation: "/one"}, "first"); next.Compose.Files[0] != "compose.yml" {
		t.Fatal("input mutation leaked")
	}
}

func TestRuntimeProjectionRetentionIsBounded(t *testing.T) {
	t.Parallel()
	var cache runtimeProjections
	for i := range 200 {
		cache.store(domain.ProjectRuntime{ProjectID: strings.Repeat("a", i+1)})
	}
	if len(cache.entries) > 128 || cache.bytes > 4<<20 {
		t.Fatal("unbounded cache")
	}
	before := len(cache.entries)
	cache.store(domain.ProjectRuntime{ProjectID: "oversize", Root: strings.Repeat("x", 1<<20)})
	if len(cache.entries) != before {
		t.Fatal("retained oversized declaration")
	}
}
