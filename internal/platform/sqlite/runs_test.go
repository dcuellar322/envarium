package sqlite

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestRunRepositoryPersistsFingerprintsAndTerminalOutcome(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "envarium.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, err = database.connection.ExecContext(ctx, `INSERT INTO projects
        (id, slug, display_name, trust_state, primary_location, created_at, updated_at)
        VALUES ('project-run', 'project-run', 'Project Run', 'trusted', '/tmp/project-run', ?, ?)`, formatTime(now), formatTime(now))
	if err != nil {
		t.Fatal(err)
	}
	repository := NewRunRepository(database)
	run := domain.RunRecord{
		ID: "run-1", ProjectID: "project-run", ServiceID: "api", RuntimeDriver: domain.KindProcess,
		Origin: domain.OriginEnvarium, StartedAt: now, IdentityFingerprint: "fingerprint-1", OperationID: "op-1",
	}
	if err := repository.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	identity := domain.ProcessIdentity{
		RunID: run.ID, PID: 123, ProcessGroup: 123, Executable: "/usr/bin/uv", StartedAt: now,
		WorkingDirectory: "/tmp/project-run", Fingerprint: "fingerprint-1", ObservedAt: now,
	}
	if err := repository.RecordProcess(ctx, identity); err != nil {
		t.Fatal(err)
	}
	if err := repository.SetRestartCount(ctx, run.ID, 1); err != nil {
		t.Fatal(err)
	}
	exitCode := 17
	endedAt := now.Add(time.Second)
	if err := repository.FinishRun(ctx, run.ID, endedAt, &exitCode, "crashed"); err != nil {
		t.Fatal(err)
	}
	runs, err := repository.ListProjectRuns(ctx, run.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].EndedAt == nil || runs[0].ExitCode == nil || *runs[0].ExitCode != 17 ||
		runs[0].RestartCount != 1 || runs[0].OperationID != "op-1" || len(runs[0].Processes) != 1 || runs[0].Processes[0].Fingerprint != "fingerprint-1" {
		t.Fatalf("runs = %#v", runs)
	}
}

func TestRunHistoryJoinPreservesRunLimitAndEveryIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "runs.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	now := time.Now().UTC().Truncate(time.Millisecond)
	if _, err := database.connection.ExecContext(ctx, `INSERT INTO projects (id,slug,display_name,trust_state,primary_location,created_at,updated_at) VALUES ('p','p','P','trusted','/tmp',?,?)`, formatTime(now), formatTime(now)); err != nil {
		t.Fatal(err)
	}
	repository := NewRunRepository(database)
	for index := 0; index < 201; index++ {
		run := domain.RunRecord{ID: fmt.Sprintf("run-%03d", index), ProjectID: "p", ServiceID: "web", RuntimeDriver: domain.KindProcess, Origin: domain.OriginEnvarium, StartedAt: now.Add(time.Duration(index) * time.Second)}
		if err := repository.CreateRun(ctx, run); err != nil {
			t.Fatal(err)
		}
		if index == 200 {
			for _, pid := range []int32{9, 7, 8} {
				identity := domain.ProcessIdentity{RunID: run.ID, PID: pid, ProcessGroup: 7, Executable: "server", StartedAt: now, ObservedAt: now, Fingerprint: fmt.Sprint(pid)}
				if err := repository.RecordProcess(ctx, identity); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	runs, err := repository.ListProjectRuns(ctx, "p")
	if err != nil || len(runs) != 200 || runs[0].ID != "run-200" || runs[199].ID != "run-001" {
		t.Fatalf("limit/order: count=%d err=%v", len(runs), err)
	}
	if len(runs[0].Processes) != 3 || runs[0].Processes[0].PID != 7 || runs[0].Processes[1].PID != 8 || runs[0].Processes[2].PID != 9 {
		t.Fatalf("identity order: %v", runs[0].Processes)
	}
	if runs[1].Processes == nil || len(runs[1].Processes) != 0 {
		t.Fatal("run without identities was lost or changed")
	}
	empty, err := repository.ListProjectRuns(ctx, "other")
	if err != nil || len(empty) != 0 {
		t.Fatal("cross-project run leak")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repository.ListProjectRuns(canceled, "p"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}
