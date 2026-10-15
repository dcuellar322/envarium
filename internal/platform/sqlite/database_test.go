package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOpenMigratesEmptyDatabase(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "envarium.db")
	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	version, err := database.SchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("SchemaVersion() error = %v", err)
	}
	if version != 18 {
		t.Fatalf("SchemaVersion() = %d, want 18", version)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); runtime.GOOS != "windows" && got != 0o600 {
		t.Fatalf("database permissions = %o, want 600", got)
	}
}

func TestOpenBacksUpAndPreservesAlphaDataBeforeUpgrade(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "envarium.db")
	connection, err := sql.Open("sqlite", sqliteFileDSN(path, ""))
	if err != nil {
		t.Fatal(err)
	}
	provider, err := migrationProvider(connection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 3); err != nil {
		t.Fatal(err)
	}
	_, err = connection.Exec(`
		INSERT INTO projects (id, slug, display_name, trust_state, primary_location, created_at, updated_at)
		VALUES ('project-1','fixture','Fixture','trusted','/tmp/fixture','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z');
		INSERT INTO manifest_proposals (id, project_id, scanner_version, schema_version, candidate_json, confidence_json, unresolved_json, validation_json, status, created_at)
		VALUES ('proposal-1','project-1','deterministic/v1','envarium.dev/v1alpha1','{}','{}','[]','{}','accepted','2026-01-01T00:00:00Z');
		INSERT INTO manifest_snapshots (project_id, revision, proposal_id, manifest_json, created_at)
		VALUES ('project-1',1,'proposal-1','{"schemaVersion":"envarium.dev/v1alpha1"}','2026-01-01T00:00:00Z');
		INSERT INTO audit_events (event_type, actor_type, actor_id, project_id, detail_json, occurred_at)
		VALUES ('project.accepted','user','fixture','project-1','{}','2026-01-01T00:00:00Z');
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := connection.Close(); err != nil {
		t.Fatal(err)
	}
	database, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	backupPath := preMigrationBackupPath(path, 3, 18)
	for _, candidate := range []string{path, backupPath} {
		check, err := sql.Open("sqlite", sqliteFileDSN(candidate, ""))
		if err != nil {
			t.Fatal(err)
		}
		for _, table := range []string{"projects", "manifest_snapshots", "audit_events"} {
			var count int
			if err := check.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 1 {
				_ = check.Close()
				t.Fatalf("%s %s count=%d error=%v", candidate, table, count, err)
			}
		}
		_ = check.Close()
	}
	status, err := InspectMigrations(ctx, path)
	if err != nil || status.CurrentVersion != 18 || status.MigrationRequired {
		t.Fatalf("status=%#v error=%v", status, err)
	}
}

func TestOpenMigratesSwitchyardIdentityToEnvarium(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "envarium.db")
	connection, err := sql.Open("sqlite", sqliteFileDSN(path, "_pragma=foreign_keys(1)"))
	if err != nil {
		t.Fatal(err)
	}
	provider, err := migrationProvider(connection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 17); err != nil {
		t.Fatal(err)
	}
	_, err = connection.Exec(`
		INSERT INTO projects (id, slug, display_name, trust_state, primary_location, created_at, updated_at)
		VALUES ('project-1','switchyard','Switchyard','trusted','/Users/example/dev/switchyard','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z');
		INSERT INTO manifest_proposals (id, project_id, scanner_version, schema_version, candidate_json, confidence_json, unresolved_json, validation_json, status, created_at)
		VALUES ('proposal-1','project-1','deterministic/v1','switchyard.dev/v1alpha1','{"schemaVersion":"switchyard.dev/v1alpha1","id":"switchyard","name":"Switchyard","command":["./bin/switchyard"]}','{}','[]','{}','accepted','2026-01-01T00:00:00Z');
		INSERT INTO manifest_snapshots (project_id, revision, proposal_id, manifest_json, created_at)
		VALUES ('project-1',1,'proposal-1','{"schemaVersion":"switchyard.dev/v1alpha1","id":"switchyard","name":"Switchyard","command":["./bin/switchyard"]}','2026-01-01T00:00:00Z');
		INSERT INTO runs (id, project_id, service_id, runtime_driver, origin, started_at, identity_fingerprint)
		VALUES ('run-1','project-1','api','process','switchyard','2026-01-01T00:00:00Z','fingerprint');
		INSERT INTO run_processes (run_id, pid, process_group_id, executable, started_at, working_directory, identity_fingerprint, observed_at)
		VALUES ('run-1',123,123,'go','2026-01-01T00:00:00Z','/Users/example/dev/switchyard','fingerprint','2026-01-01T00:00:00Z');
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := connection.Close(); err != nil {
		t.Fatal(err)
	}

	database, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	check, err := sql.Open("sqlite", sqliteFileDSN(path, "_pragma=foreign_keys(1)"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = check.Close() }()
	var slug, name, location, origin, workingDirectory, manifest string
	if err := check.QueryRow("SELECT slug, display_name, primary_location FROM projects WHERE id = 'project-1'").Scan(&slug, &name, &location); err != nil {
		t.Fatal(err)
	}
	if err := check.QueryRow("SELECT origin FROM runs WHERE id = 'run-1'").Scan(&origin); err != nil {
		t.Fatal(err)
	}
	if err := check.QueryRow("SELECT working_directory FROM run_processes WHERE run_id = 'run-1'").Scan(&workingDirectory); err != nil {
		t.Fatal(err)
	}
	if err := check.QueryRow("SELECT manifest_json FROM manifest_snapshots WHERE project_id = 'project-1'").Scan(&manifest); err != nil {
		t.Fatal(err)
	}
	if slug != "envarium" || name != "Envarium" || location != "/Users/example/dev/envarium" {
		t.Fatalf("project identity = %q %q %q", slug, name, location)
	}
	if origin != "envarium" || workingDirectory != "/Users/example/dev/envarium" {
		t.Fatalf("runtime identity = %q %q", origin, workingDirectory)
	}
	if strings.Contains(manifest, "switchyard") || !strings.Contains(manifest, "envarium.dev") {
		t.Fatalf("manifest identity = %s", manifest)
	}
}

func TestOpenRejectsDatabaseFromNewerBinary(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "envarium.db")
	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	connection, err := sql.Open("sqlite", sqliteFileDSN(path, ""))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := connection.Exec(`INSERT INTO goose_db_version (version_id, is_applied, tstamp) VALUES (999, 1, CURRENT_TIMESTAMP)`); err != nil {
		_ = connection.Close()
		t.Fatalf("insert future migration version: %v", err)
	}
	if err := connection.Close(); err != nil {
		t.Fatalf("close future database: %v", err)
	}

	_, err = Open(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "newer than this binary supports") {
		t.Fatalf("Open() error = %v, want newer-schema rejection", err)
	}
}
