package sqlite

import (
	"context"
	observability "envarium.dev/envarium/internal/observability/domain"
	"envarium.dev/envarium/internal/runtime/domain"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestNativeRunsIsolateWorktreeScopesUnderCatalogOwner(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "runs.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	now := time.Now().UTC().Truncate(time.Millisecond)
	if _, err := db.connection.ExecContext(ctx, `INSERT INTO projects(id,slug,display_name,trust_state,primary_location,created_at,updated_at) VALUES('p','p','P','trusted','/tmp',?,?)`, formatTime(now), formatTime(now)); err != nil {
		t.Fatal(err)
	}
	store := NewRunRepository(db)
	for i, scope := range []string{"p", "env-one", "env-two"} {
		run := domain.RunRecord{ID: fmt.Sprint("run", i), ProjectID: scope, CatalogProjectID: "p", ServiceID: "api", RuntimeDriver: domain.KindProcess, Origin: domain.OriginEnvarium, StartedAt: now, IdentityFingerprint: scope}
		if err := store.CreateRun(ctx, run); err != nil {
			t.Fatal(err)
		}
		if err := store.RecordProcess(ctx, domain.ProcessIdentity{RunID: run.ID, PID: int32(i + 100), StartedAt: now, ObservedAt: now, Fingerprint: scope}); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.FinishRun(ctx, "run1", now, nil, "stopped"); err != nil {
		t.Fatal(err)
	}
	all, err := store.ListProjectsRuns(ctx, []string{"p", "env-one", "env-two"})
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"p", "env-one", "env-two"} {
		runs := all[scope]
		if len(runs) != 1 || runs[0].CatalogProjectID != "p" || runs[0].ProjectID != scope || runs[0].Processes[0].Fingerprint != scope {
			t.Fatalf("scope lost: %#v", runs)
		}
		if (runs[0].EndedAt != nil) != (scope == "env-one") {
			t.Fatal("cross-scope stop")
		}
	}
	// A long-lived run remains visible after more than the retained history limit.
	for i := range 205 {
		end := now.Add(time.Duration(i+1) * time.Second)
		if err := store.CreateRun(ctx, domain.RunRecord{ID: fmt.Sprint("old", i), ProjectID: "env-two", CatalogProjectID: "p", ServiceID: "old", RuntimeDriver: domain.KindProcess, Origin: domain.OriginEnvarium, StartedAt: end, EndedAt: &end}); err != nil {
			t.Fatal(err)
		}
	}
	runs, err := store.ListProjectRuns(ctx, "env-two")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, run := range runs {
		found = found || run.ID == "run2"
	}
	if !found {
		t.Fatal("active ownership evicted by ended history")
	}
}

func TestNativeWorktreeHealthAndLogArchiveKeepScopeAndCatalogOwner(t *testing.T) {
	database, logs, _ := newTestLogStore(t, LogStoreConfig{SegmentBytes: 1 << 20, RetentionBytes: 1 << 20})
	ctx := context.Background()
	now := time.Now().UTC()
	for _, scope := range []string{"env-a", "env-b"} {
		entry := domain.LogEntry{CatalogProjectID: "project-1", ProjectID: scope, ServiceID: "web", RunID: scope, Timestamp: now, Message: scope}
		if err := logs.WriteLog(ctx, entry); err != nil {
			t.Fatal(err)
		}
		health := observability.HealthResult{CatalogProjectID: "project-1", ProjectID: scope, ServiceID: "web", CheckID: "ready", Type: "http", Status: observability.StatusHealthy, Severity: "warning", ObservedAt: now}
		if err := NewHealthRepository(database).AppendHealth(ctx, health); err != nil {
			t.Fatal(err)
		}
	}
	for _, scope := range []string{"env-a", "env-b"} {
		entries, err := logs.QueryLogs(ctx, observability.LogQuery{ProjectID: scope, Limit: 10})
		if err != nil || len(entries) != 1 || entries[0].Message != scope {
			t.Fatalf("log scope lost: %#v %v", entries, err)
		}
		health, err := NewHealthRepository(database).LatestHealth(ctx, scope)
		if err != nil || len(health) != 1 || health[0].ProjectID != scope {
			t.Fatalf("health scope lost: %#v %v", health, err)
		}
	}
}
