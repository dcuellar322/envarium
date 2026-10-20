package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envarium.dev/envarium/internal/bootstrap"
)

func TestDataInspectAndBackupCommandsDoNotStartDaemon(t *testing.T) {
	t.Parallel()
	dataDir := t.TempDir()
	if _, err := bootstrap.MigrateData(context.Background(), dataDir); err != nil {
		t.Fatal(err)
	}
	databasePath := filepath.Join(dataDir, "envarium.db")
	database, err := sql.Open("sqlite", sqliteTestDSN(databasePath))
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(`
		INSERT INTO settings (singleton, revision, document_json, updated_at)
		VALUES (1, 4, '{"fixture":"settings-survive-backup"}', '2026-07-16T00:00:00Z');
		INSERT INTO settings_audit_events (revision, actor_type, actor_id, sections_json, occurred_at)
		VALUES (4, 'user', 'fixture', '["appearance"]', '2026-07-16T00:00:00Z');
	`)
	if err != nil {
		_ = database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := Execute(context.Background(), []string{"--data-dir", dataDir, "data", "inspect"}, &output, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "schema=19 target=19 migration=false") {
		t.Fatalf("inspect output = %q", output.String())
	}
	output.Reset()
	backup := filepath.Join(dataDir, "manual.bak")
	if err := Execute(context.Background(), []string{"--data-dir", dataDir, "data", "backup", "--output", backup}, &output, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Verified backup") {
		t.Fatalf("backup output = %q", output.String())
	}
	if info, err := os.Stat(backup); err != nil || info.Size() == 0 {
		t.Fatalf("backup info=%v error=%v", info, err)
	}
	backupDatabase, err := sql.Open("sqlite", sqliteTestDSN(backup))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = backupDatabase.Close() }()
	var revision, auditCount int
	var document string
	if err := backupDatabase.QueryRow(`SELECT revision, document_json FROM settings WHERE singleton=1`).Scan(&revision, &document); err != nil {
		t.Fatal(err)
	}
	if err := backupDatabase.QueryRow(`SELECT COUNT(*) FROM settings_audit_events WHERE revision=4`).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if revision != 4 || document != `{"fixture":"settings-survive-backup"}` || auditCount != 1 {
		t.Fatalf("backup settings revision=%d document=%q audits=%d", revision, document, auditCount)
	}
}

func sqliteTestDSN(path string) string {
	normalized := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" && !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return (&url.URL{Scheme: "file", Path: normalized}).String()
}

func TestDataMigrateRejectsNewerSchemaWithoutMutation(t *testing.T) {
	t.Parallel()
	for _, machine := range []bool{false, true} {
		for _, write := range []bool{false, true} {
			t.Run(fmt.Sprintf("json=%t/write=%t", machine, write), func(t *testing.T) {
				dataDir := t.TempDir()
				if _, err := bootstrap.MigrateData(t.Context(), dataDir); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dataDir, "envarium.db")
				database, err := sql.Open("sqlite", sqliteTestDSN(path))
				if err != nil {
					t.Fatal(err)
				}
				_, err = database.Exec(`INSERT INTO goose_db_version (version_id, is_applied, tstamp) VALUES (999, 1, CURRENT_TIMESTAMP)`)
				closeErr := database.Close()
				if err != nil || closeErr != nil {
					t.Fatalf("prepare future schema: %v, %v", err, closeErr)
				}
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				args := []string{"--data-dir", dataDir, "data", "migrate"}
				if machine {
					args = append(args, "--json")
				}
				if write {
					args = append(args, "--write")
				}
				var output bytes.Buffer
				err = Execute(t.Context(), args, &output, &output)
				if err == nil || output.Len() != 0 {
					t.Fatalf("unsupported schema reported success: error=%v output=%q", err, output.String())
				}
				failure := classifyError(err)
				if failure.Code != "DATABASE_SCHEMA_NEWER" || failure.ExitCode != exitConflict || !strings.Contains(failure.Message, "newer than this binary supports") {
					t.Fatalf("failure = %#v", failure)
				}
				if machine {
					if err := writeMachineError(&output, failure); err != nil {
						t.Fatal(err)
					}
					var envelope struct {
						Error Error `json:"error"`
					}
					if err := json.Unmarshal(output.Bytes(), &envelope); err != nil || envelope.Error.Code != failure.Code || envelope.Error.ExitCode != exitConflict {
						t.Fatalf("machine failure = %s, %v", output.Bytes(), err)
					}
				}
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatalf("newer database changed: %v", err)
				}
				entries, err := os.ReadDir(dataDir)
				if err != nil {
					t.Fatal(err)
				}
				for _, entry := range entries {
					switch entry.Name() {
					case "envarium.db", "envarium.db-wal", "envarium.db-shm":
					default:
						t.Fatalf("unexpected backup/daemon side effect: %s", entry.Name())
					}
				}
			})
		}
	}
}
