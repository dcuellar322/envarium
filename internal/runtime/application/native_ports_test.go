package application

import (
	"context"
	catalog "envarium.dev/envarium/internal/catalog/application"
	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	manifest "envarium.dev/envarium/internal/manifest/application"
	"envarium.dev/envarium/internal/runtime/domain"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestNativePortExpansionUsesEffectiveValuesOnly(t *testing.T) {
	project := domain.ProjectRuntime{Ports: map[string]domain.PortDeclaration{"http": {Host: 19001}, "other": {Host: 19002}}, Process: &domain.ProcessRuntime{
		Environment: map[string]string{"PORT": "${ports.http}", "OVERRIDE": "${ports.other}"},
		Secrets:     map[string]domain.SecretReference{"PORT": {Key: "global"}},
		Processes:   []domain.ProcessDefinition{{ID: "web", Command: []string{"node", "server.js", "--url=http://localhost:${ports.http}"}, Environment: map[string]string{"PORT": "${ports.http}", "OVERRIDE": "literal-8080"}}},
	}}
	if err := projectNativePorts(&project); err != nil {
		t.Fatal(err)
	}
	process := project.Process.Processes[0]
	if !slices.Equal(process.PortBindings, []string{"http"}) || process.Environment["PORT"] != "19001" || process.Command[2] != "--url=http://localhost:19001" || process.Environment["OVERRIDE"] != "literal-8080" {
		t.Fatalf("unexpected expansion: %#v", process)
	}
}
func TestNativePortExpansionRejectsUnknownMalformedAndExecutableReferences(t *testing.T) {
	for _, command := range [][]string{{"server", "${ports.unknown}"}, {"server", "${ports.http"}, {"${ports.http}"}} {
		project := domain.ProjectRuntime{Ports: map[string]domain.PortDeclaration{"http": {Host: 12345}}, Process: &domain.ProcessRuntime{Processes: []domain.ProcessDefinition{{ID: "web", Command: command}}}}
		if err := projectNativePorts(&project); err == nil {
			t.Fatalf("accepted invalid reference: %v", command)
		}
	}
}
func TestNativeSecretOverrideCannotClaimPortBinding(t *testing.T) {
	project := domain.ProjectRuntime{Ports: map[string]domain.PortDeclaration{"http": {Host: 19001}}, Process: &domain.ProcessRuntime{Environment: map[string]string{"PORT": "${ports.http}"}, Processes: []domain.ProcessDefinition{{ID: "web", Command: []string{"server"}, Secrets: map[string]domain.SecretReference{"PORT": {Key: "secret"}}}}}}
	if err := projectNativePorts(&project); err != nil {
		t.Fatal(err)
	}
	if len(project.Process.Processes[0].PortBindings) != 0 {
		t.Fatal("secret override falsely consumed lease")
	}
}

func TestRuntimeSourceReadsEachCheckoutManifestAndFreshLocalOverlay(t *testing.T) {
	accepted, err := manifest.ParseYAML([]byte(`schemaVersion: envarium.dev/v1
kind: Project
metadata: {id: fixture, name: Fixture}
repository: {root: .}
runtime:
  driver: process
  process:
    processes:
      - id: web
        command: [main-server, "${ports.http}"]
services:
  - id: web
    source: {process: web}
ports:
  - {id: http, service: web, host: 8080, target: 8080, protocol: tcp}
`))
	if err != nil {
		t.Fatal(err)
	}
	main, feature := t.TempDir(), t.TempDir()
	directory := filepath.Join(feature, ".envarium")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	write := func(command string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(directory, "project.local.yml"), []byte("runtime:\n  process:\n    processes:\n      - id: web\n        command: ["+command+", '${ports.http}']\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("feature-server")
	repository := &snapshotCatalogRepository{manifest: accepted, projects: []catalogDomain.Project{{ID: "p", Slug: "p", PrimaryLocation: main, TrustState: catalogDomain.TrustTrusted, ManifestRevision: 1}}}
	source := NewCatalogSource(catalog.NewService(repository, nil), &batchedEnvironmentSource{items: []RuntimeEnvironment{{ID: "env-f", ProjectID: "p", Root: feature, Available: true, PortLeases: map[string]int{"http": 19001}}}})
	for range 2 {
		base, err := source.ResolveRuntime(context.Background(), "p")
		if err != nil {
			t.Fatal(err)
		}
		selected, err := source.ResolveRuntime(context.Background(), "env-f")
		if err != nil {
			t.Fatal(err)
		}
		if base.Process.Processes[0].Command[0] != "main-server" || selected.Root != feature || selected.Process.Processes[0].Command[0] != "feature-server" || selected.Process.Processes[0].Command[1] != "19001" || !selected.Ports["http"].Allocated || selected.CatalogProjectID != "p" {
			t.Fatalf("wrong checkout projected: %#v %#v", base, selected)
		}
	}
	write("fresh-server")
	selected, err := source.ResolveRuntime(context.Background(), "env-f")
	if err != nil || selected.Process.Processes[0].Command[0] != "fresh-server" {
		t.Fatal("local checkout edit hidden", err)
	}
	repository.projects[0].TrustState = catalogDomain.TrustPending
	if _, err := source.ResolveRuntime(context.Background(), "env-f"); !errors.Is(err, ErrProjectUntrusted) {
		t.Fatal("parent trust bypassed", err)
	}
}
