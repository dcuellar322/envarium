package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

// RunRepository persists native-process ownership evidence and lifecycle outcomes.
type RunRepository struct{ database *Database }

// NewRunRepository creates a durable native-process run adapter.
func NewRunRepository(database *Database) *RunRepository {
	return &RunRepository{database: database}
}

// CreateRun writes a new run before its process is exposed as managed.
func (r *RunRepository) CreateRun(ctx context.Context, run domain.RunRecord) error {
	owner := run.CatalogProjectID
	if owner == "" {
		owner = run.ProjectID
	}
	_, err := r.database.connection.ExecContext(ctx, `INSERT INTO runs
        (id, project_id, service_id, runtime_driver, origin, started_at, ended_at, exit_code,
         termination_reason, identity_fingerprint, restart_count, operation_id, runtime_scope_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, owner, run.ServiceID, run.RuntimeDriver, run.Origin, formatTime(run.StartedAt),
		nullTime(run.EndedAt), nullInt(run.ExitCode), run.TerminationReason, run.IdentityFingerprint, run.RestartCount, nullString(run.OperationID), run.ProjectID)
	if err != nil {
		return fmt.Errorf("create runtime run: %w", err)
	}
	return nil
}

// RecordProcess upserts one verified member of a managed process group.
func (r *RunRepository) RecordProcess(ctx context.Context, identity domain.ProcessIdentity) error {
	_, err := r.database.connection.ExecContext(ctx, `INSERT INTO run_processes
        (run_id, pid, process_group_id, executable, started_at, working_directory, identity_fingerprint, observed_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(run_id, pid, started_at) DO UPDATE SET
          process_group_id = excluded.process_group_id,
          executable = excluded.executable,
          working_directory = excluded.working_directory,
          identity_fingerprint = excluded.identity_fingerprint,
          observed_at = excluded.observed_at`,
		identity.RunID, identity.PID, identity.ProcessGroup, identity.Executable, formatTime(identity.StartedAt),
		identity.WorkingDirectory, identity.Fingerprint, formatTime(identity.ObservedAt))
	if err != nil {
		return fmt.Errorf("record run process: %w", err)
	}
	return nil
}

// FinishRun records an immutable terminal outcome for an active run.
func (r *RunRepository) FinishRun(ctx context.Context, runID string, endedAt time.Time, exitCode *int, reason string) error {
	_, err := r.database.connection.ExecContext(ctx, `UPDATE runs
        SET ended_at = ?, exit_code = ?, termination_reason = ?
        WHERE id = ? AND ended_at IS NULL`, formatTime(endedAt), nullInt(exitCode), reason, runID)
	if err != nil {
		return fmt.Errorf("finish runtime run: %w", err)
	}
	return nil
}

// SetRestartCount records each opt-in supervisor restart attempt.
func (r *RunRepository) SetRestartCount(ctx context.Context, runID string, count int) error {
	_, err := r.database.connection.ExecContext(ctx, `UPDATE runs SET restart_count = ? WHERE id = ?`, count, runID)
	if err != nil {
		return fmt.Errorf("update runtime restart count: %w", err)
	}
	return nil
}

// ListProjectRuns returns newest runs first with every recorded process fingerprint.
func (r *RunRepository) ListProjectRuns(ctx context.Context, projectID string) ([]domain.RunRecord, error) {
	result, err := r.ListProjectsRuns(ctx, []string{projectID})
	if err != nil {
		return nil, err
	}
	return result[projectID], nil
}

func scanRunWithProcess(row rowScanner) (domain.RunRecord, domain.ProcessIdentity, error) {
	var id, executable, started, directory, fingerprint, observed sql.NullString
	var pid, group sql.NullInt64
	run, err := scanRun(row, &id, &pid, &group, &executable, &started, &directory, &fingerprint, &observed)
	if err != nil || !id.Valid {
		return run, domain.ProcessIdentity{}, err
	}
	if pid.Int64 < math.MinInt32 || pid.Int64 > math.MaxInt32 || group.Int64 < math.MinInt32 || group.Int64 > math.MaxInt32 {
		return domain.RunRecord{}, domain.ProcessIdentity{}, errors.New("persisted process identity is outside PID range")
	}
	began, err := parseTime(started.String)
	if err != nil {
		return domain.RunRecord{}, domain.ProcessIdentity{}, err
	}
	at, err := parseTime(observed.String)
	if err != nil {
		return domain.RunRecord{}, domain.ProcessIdentity{}, err
	}
	return run, domain.ProcessIdentity{RunID: id.String, PID: int32(pid.Int64), ProcessGroup: int32(group.Int64),
		Executable: executable.String, StartedAt: began, WorkingDirectory: directory.String,
		Fingerprint: fingerprint.String, ObservedAt: at}, nil
}

func scanRun(row rowScanner, extra ...any) (domain.RunRecord, error) {
	var run domain.RunRecord
	var started string
	var ended sql.NullString
	var exitCode sql.NullInt64
	var operationID sql.NullString
	destinations := []any{&run.ID, &run.ProjectID, &run.CatalogProjectID, &run.ServiceID, &run.RuntimeDriver, &run.Origin,
		&started, &ended, &exitCode, &run.TerminationReason, &run.IdentityFingerprint, &run.RestartCount, &operationID}
	if err := row.Scan(append(destinations, extra...)...); err != nil {
		return domain.RunRecord{}, err
	}
	run.StartedAt, _ = parseTime(started)
	if ended.Valid {
		parsed, err := parseTime(ended.String)
		if err != nil {
			return domain.RunRecord{}, err
		}
		run.EndedAt = &parsed
	}
	if exitCode.Valid {
		value := int(exitCode.Int64)
		run.ExitCode = &value
	}
	run.OperationID = operationID.String
	return run, nil
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func nullInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
