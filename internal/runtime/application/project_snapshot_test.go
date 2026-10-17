package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	catalog "envarium.dev/envarium/internal/catalog/application"
	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	manifest "envarium.dev/envarium/internal/manifest/application"
	manifestDomain "envarium.dev/envarium/internal/manifest/domain"
)

type snapshotCatalogRepository struct {
	catalog.Repository
	projects    []catalogDomain.Project
	manifest    manifestDomain.Manifest
	gets, lists int
}

func (r *snapshotCatalogRepository) ListProjects(context.Context) ([]catalogDomain.Project, error) {
	r.lists++
	return append([]catalogDomain.Project(nil), r.projects...), nil
}
func (r *snapshotCatalogRepository) GetProject(_ context.Context, id string) (catalogDomain.Project, error) {
	r.gets++
	for _, p := range r.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return catalogDomain.Project{}, catalog.ErrNotFound
}
func (r *snapshotCatalogRepository) AcceptedManifest(context.Context, string) (manifestDomain.Manifest, error) {
	return r.manifest, nil
}

func TestRuntimeSnapshotBatchesCatalogWithoutCachingTrustOrLocalEdits(t *testing.T) {
	t.Parallel()
	good, bad := t.TempDir(), t.TempDir()
	accepted, err := manifest.ParseYAML([]byte(`schemaVersion: envarium.dev/v1
kind: Project
metadata: {id: fixture, name: Fixture}
repository: {root: .}
runtime:
  driver: process
  process:
    processes:
      - id: web
        command: [fixture-server]
services:
  - id: web
    source: {process: web}
`))
	if err != nil {
		t.Fatal(err)
	}
	badDir := filepath.Join(bad, ".envarium")
	if err := os.MkdirAll(badDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badDir, "project.local.yml"), []byte("invalid: ["), 0600); err != nil {
		t.Fatal(err)
	}
	repo := &snapshotCatalogRepository{manifest: accepted, projects: []catalogDomain.Project{
		{ID: "good", Slug: "good", PrimaryLocation: good, TrustState: catalogDomain.TrustTrusted, ManifestRevision: 1},
		{ID: "bad", Slug: "bad", PrimaryLocation: bad, TrustState: catalogDomain.TrustTrusted, ManifestRevision: 1},
		{ID: "untrusted", Slug: "untrusted", PrimaryLocation: good, TrustState: catalogDomain.TrustPending, ManifestRevision: 1},
	}}
	source := NewCatalogSource(catalog.NewService(repo, nil))
	projects, failures, err := source.ResolveRuntimeProjects(context.Background())
	if err != nil || len(projects) != 1 || projects["good"].Root != good || failures["bad"] == nil || len(failures) != 1 {
		t.Fatalf("snapshot: projects=%v failures=%v err=%v", projects, failures, err)
	}
	if repo.gets != 0 || repo.lists != 1 {
		t.Fatalf("catalog fan-out: gets=%d lists=%d", repo.gets, repo.lists)
	}
	directory := filepath.Join(good, ".envarium")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "project.local.yml"), []byte("runtime:\n  process:\n    processes:\n      - id: web\n        command: [changed-server]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	projects, _, err = source.ResolveRuntimeProjects(context.Background())
	if err != nil || projects["good"].Process.Processes[0].Command[0] != "changed-server" {
		t.Fatal("local edit hidden by batching")
	}
	repo.projects[0].TrustState = catalogDomain.TrustPending
	if _, err := source.ResolveRuntime(context.Background(), "good"); !errors.Is(err, ErrProjectUntrusted) {
		t.Fatalf("direct trust check: %v", err)
	}
	projects, _, err = source.ResolveRuntimeProjects(context.Background())
	if err != nil || len(projects) != 0 || repo.gets != 1 {
		t.Fatal("snapshot retained revoked trust or direct resolution used cached metadata")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := source.ResolveRuntimeProjects(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}
