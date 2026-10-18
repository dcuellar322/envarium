package application

import (
	"context"
	"errors"
	"testing"

	catalog "envarium.dev/envarium/internal/catalog/application"
	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	manifest "envarium.dev/envarium/internal/manifest/application"
)

type batchedEnvironmentSource struct {
	items       []RuntimeEnvironment
	reads, gets int
	err         error
}

func (s *batchedEnvironmentSource) ListRuntimeEnvironments(ctx context.Context) ([]RuntimeEnvironment, error) {
	s.reads++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.items, s.err
}
func (s *batchedEnvironmentSource) ListRuntimeEnvironmentIDs(context.Context) ([]string, error) {
	panic("scheduler must consume the batch instead of listing and fetching allocations again")
}
func (s *batchedEnvironmentSource) ResolveRuntimeEnvironment(_ context.Context, id string) (RuntimeEnvironment, error) {
	s.gets++
	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}
	return RuntimeEnvironment{}, errors.New("environment missing")
}

func TestEnvironmentBatchRetainsTrustAvailabilityAndFreshExactLeases(t *testing.T) {
	accepted, err := manifest.ParseYAML([]byte(`schemaVersion: envarium.dev/v1
kind: Project
metadata: {id: fixture, name: Fixture}
repository: {root: .}
runtime:
  driver: compose
  compose: {files: [compose.yml], projectName: fixture}
services:
  - id: api
    source: {composeService: web}
ports:
  - {id: http, service: api, host: 8080, target: 80, protocol: tcp}
`))
	if err != nil {
		t.Fatal(err)
	}
	repository := &snapshotCatalogRepository{manifest: accepted, projects: []catalogDomain.Project{{
		ID: "project", Slug: "project", PrimaryLocation: t.TempDir(), TrustState: catalogDomain.TrustTrusted, ManifestRevision: 1,
	}}}
	environments := &batchedEnvironmentSource{items: []RuntimeEnvironment{
		{ID: "env-feature", ProjectID: "project", Name: "feature", Root: t.TempDir(), Available: true, ComposeProjectName: "sy-feature", PortLeases: map[string]int{"http": 19001}},
		{ID: "env-missing-parent", ProjectID: "missing", Available: true},
		{ID: "env-unavailable", ProjectID: "project", Available: false},
	}}
	source := NewCatalogSource(catalog.NewService(repository, nil), environments)
	projects, failures, err := source.ResolveRuntimeProjects(context.Background())
	if err != nil || len(projects) != 2 || len(failures) != 2 || failures["env-missing-parent"] == nil || failures["env-unavailable"] == nil {
		t.Fatalf("snapshot projects=%v failures=%v err=%v", projects, failures, err)
	}
	feature := projects["env-feature"]
	if feature.Root != environments.items[0].Root || feature.Compose.PortOverrides["http"] != 19001 || feature.Ports["http"].Host != 19001 || environments.gets != 0 || environments.reads != 1 {
		t.Fatalf("allocation changed or refetched: %#v", feature)
	}
	oldHash := feature.ManifestHash
	feature.Compose.PortOverrides["http"] = 1
	environments.items[0].PortLeases["http"] = 19002
	projects, _, err = source.ResolveRuntimeProjects(context.Background())
	if err != nil || projects["env-feature"].Compose.PortOverrides["http"] != 19002 {
		t.Fatal("batch hid current leases or retained caller mutations")
	}
	if projects["env-feature"].ManifestHash == oldHash {
		t.Fatal("changed allocation retained old review identity")
	}
	stableHash := projects["env-feature"].ManifestHash
	projects, _, err = source.ResolveRuntimeProjects(context.Background())
	if err != nil || projects["env-feature"].ManifestHash != stableHash {
		t.Fatal("unchanged allocation changed review identity")
	}
	if _, err := source.ResolveRuntime(context.Background(), "env-feature"); err != nil || environments.gets != 1 {
		t.Fatalf("direct lifecycle resolution did not reread allocation: %v", err)
	}
	repository.projects[0].TrustState = catalogDomain.TrustPending
	projects, failures, err = source.ResolveRuntimeProjects(context.Background())
	if err != nil || len(projects) != 0 || !errors.Is(failures["env-feature"], ErrProjectUntrusted) {
		t.Fatal("batch bypassed current parent trust")
	}
	environments.err = errors.New("registry unavailable")
	if _, _, err := source.ResolveRuntimeProjects(context.Background()); !errors.Is(err, environments.err) {
		t.Fatalf("batch failure hidden: %v", err)
	}
}
