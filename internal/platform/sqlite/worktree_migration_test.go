package sqlite

import (
	"context"
	"database/sql"
	"envarium.dev/envarium/internal/runtime/domain"
	"path/filepath"
	"testing"
	"time"
)

func TestNativeScopeUpgradePreservesLegacyOwnershipAndRejectsLossyDowngrade(t *testing.T) {
	ctx := context.Background()
	connection, err := sql.Open("sqlite", sqliteFileDSN(filepath.Join(t.TempDir(), "upgrade.db"), ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = connection.Close() }()
	provider, err := migrationProvider(connection)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 18); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, err = connection.ExecContext(ctx, `INSERT INTO projects(id,slug,display_name,trust_state,primary_location,created_at,updated_at) VALUES('p','p','P','trusted','/tmp',?,?)`, formatTime(now), formatTime(now))
	if err != nil {
		t.Fatal(err)
	}
	_, err = connection.ExecContext(ctx, `INSERT INTO runs(id,project_id,service_id,runtime_driver,origin,started_at,termination_reason,identity_fingerprint,restart_count) VALUES('legacy','p','web','process','envarium',?,'','identity',0)`, formatTime(now))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	var scope, owner, fingerprint string
	if err := connection.QueryRowContext(ctx, `SELECT runtime_scope_id,project_id,identity_fingerprint FROM runs WHERE id='legacy'`).Scan(&scope, &owner, &fingerprint); err != nil || scope != "p" || owner != "p" || fingerprint != "identity" {
		t.Fatal("legacy evidence changed", scope, owner, fingerprint, err)
	}
	if _, err := provider.Down(ctx); err != nil {
		t.Fatal("lossless downgrade failed", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	repository := NewRunRepository(&Database{connection: connection})
	if err := repository.CreateRun(ctx, domain.RunRecord{ID: "isolated", ProjectID: "env-f", CatalogProjectID: "p", ServiceID: "web", RuntimeDriver: domain.KindProcess, Origin: domain.OriginEnvarium, StartedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Down(ctx); err == nil {
		t.Fatal("lossy downgrade accepted")
	}
	if err := connection.QueryRowContext(ctx, `SELECT runtime_scope_id FROM runs WHERE id='isolated'`).Scan(&scope); err != nil || scope != "env-f" {
		t.Fatal("failed downgrade changed ownership", scope, err)
	}
}
