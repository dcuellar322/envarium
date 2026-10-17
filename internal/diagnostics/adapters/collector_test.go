package adapters

import (
	"context"
	"errors"
	"testing"
	"time"

	actions "envarium.dev/envarium/internal/actions/domain"
	catalog "envarium.dev/envarium/internal/catalog/domain"
	manifest "envarium.dev/envarium/internal/manifest/application"
	observability "envarium.dev/envarium/internal/observability/domain"
	operations "envarium.dev/envarium/internal/operations/domain"
	ports "envarium.dev/envarium/internal/ports/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
	git "envarium.dev/envarium/internal/sourcecontrol/domain"
)

type diagnosticCatalogFixture struct{ catalogSource }

func (diagnosticCatalogFixture) GetProject(ctx context.Context, id string) (catalog.Project, error) {
	return catalog.Project{ID: id, DisplayName: "Fixture", TrustState: catalog.TrustTrusted, UpdatedAt: time.Now()}, ctx.Err()
}
func (diagnosticCatalogFixture) EffectiveManifest(ctx context.Context, _ string, _ []byte) (manifest.EffectiveManifest, error) {
	return manifest.EffectiveManifest{Provenance: map[string]string{"/runtime": "accepted"}}, ctx.Err()
}

type diagnosticRuntimeFixture struct{}

func (diagnosticRuntimeFixture) Inspect(ctx context.Context, id string) (runtime.Observation, error) {
	return runtime.Observation{ProjectID: id, Driver: runtime.KindProcess, State: runtime.StateRunning}, ctx.Err()
}

type diagnosticHealthFixture struct{}

func (diagnosticHealthFixture) Get(ctx context.Context, id string) (observability.ProjectHealth, error) {
	return observability.ProjectHealth{ProjectID: id}, ctx.Err()
}

type diagnosticLogsFixture struct{}

func (diagnosticLogsFixture) Logs(ctx context.Context, _ string, _ string, _ string, _ string, _ string, _ int) ([]runtime.LogEntry, error) {
	return []runtime.LogEntry{{Message: "connection refused", Timestamp: time.Now(), Redacted: true}}, ctx.Err()
}

type diagnosticPortsFixture struct{}

func (diagnosticPortsFixture) Registry(ctx context.Context) (ports.Registry, error) {
	return ports.Registry{ObservedAt: time.Now()}, ctx.Err()
}

type diagnosticGitFixture struct{ calls int }

func (g *diagnosticGitFixture) Get(ctx context.Context, _ string) (git.State, error) {
	g.calls++
	return git.State{}, ctx.Err()
}

type diagnosticActionsFixture struct{}

func (diagnosticActionsFixture) List(ctx context.Context, id string) (actions.ProjectActions, error) {
	return actions.ProjectActions{ProjectID: id, Actions: []actions.Definition{{ID: "check", Name: "Check", Type: "health.check", Risk: actions.RiskReadOnly}}}, ctx.Err()
}

type diagnosticResourcesFixture struct{ cleanupCalls int }

func (r *diagnosticResourcesFixture) Overview(ctx context.Context) (observability.ResourceOverview, error) {
	return observability.ResourceOverview{ObservedAt: time.Now()}, ctx.Err()
}
func (r *diagnosticResourcesFixture) CleanupPreview(ctx context.Context, _ string) (observability.CleanupPreview, error) {
	r.cleanupCalls++
	return observability.CleanupPreview{}, ctx.Err()
}

type diagnosticOperationsFixture struct{}

func (diagnosticOperationsFixture) List(ctx context.Context, _ string, _ int64) ([]operations.Operation, error) {
	return []operations.Operation{}, ctx.Err()
}

func TestScheduledCollectorOmitsOnlyInteractiveEnrichment(t *testing.T) {
	t.Parallel()
	gitSource := &diagnosticGitFixture{}
	resources := &diagnosticResourcesFixture{}
	collector := NewCollector(diagnosticCatalogFixture{}, diagnosticRuntimeFixture{}, diagnosticHealthFixture{}, diagnosticLogsFixture{}, diagnosticPortsFixture{}, gitSource, diagnosticActionsFixture{}, resources, diagnosticOperationsFixture{})
	bundle, err := collector.CollectScheduled(t.Context(), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if gitSource.calls != 0 || resources.cleanupCalls != 0 {
		t.Fatal("scheduled collection ran Git or cleanup enrichment")
	}
	kinds := map[string]bool{}
	for _, evidence := range bundle.Evidence {
		kinds[evidence.Kind] = true
	}
	for _, kind := range []string{"project", "runtime", "health", "config", "log", "actions", "operations"} {
		if !kinds[kind] {
			t.Fatalf("notification evidence omitted: %s", kind)
		}
	}
	if len(bundle.Warnings) != 1 || len(bundle.Actions) != 1 || len(bundle.Snapshot.RecentLogs) != 1 {
		t.Fatal("omission scope or current evidence missing")
	}
	if _, err := collector.Collect(t.Context(), "fixture"); err != nil {
		t.Fatal(err)
	}
	if gitSource.calls != 1 || resources.cleanupCalls != 1 {
		t.Fatal("full manual enrichment changed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := collector.CollectScheduled(ctx, "fixture"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation became a successful bundle")
	}
}
