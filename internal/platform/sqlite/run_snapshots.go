package sqlite

import (
	"context"
	"fmt"
	"strings"

	"envarium.dev/envarium/internal/runtime/domain"
)

// ListProjectsRuns reads fresh ownership evidence for a bounded watcher cycle.
// The limit remains per project, including all its retained process identities.
func (r *RunRepository) ListProjectsRuns(ctx context.Context, ids []string) (map[string][]domain.RunRecord, error) {
	result := make(map[string][]domain.RunRecord, len(ids))
	if len(ids) > 64 {
		return nil, fmt.Errorf("run snapshot exceeds 64 projects")
	}
	selected := []string{}
	args := []any{}
	for _, id := range ids {
		if _, exists := result[id]; exists {
			continue
		}
		result[id] = []domain.RunRecord{}
		args = append(args, id)
		selected = append(selected, `SELECT * FROM (SELECT * FROM runs WHERE runtime_scope_id = ? ORDER BY (ended_at IS NULL) DESC, started_at DESC, id DESC LIMIT 200)`)
	}
	if len(args) == 0 {
		return result, ctx.Err()
	}
	rows, err := r.database.queryContext(ctx, `WITH selected AS (`+strings.Join(selected, " UNION ALL ")+`)
 SELECT r.id, r.runtime_scope_id, r.project_id, r.service_id, r.runtime_driver, r.origin, r.started_at,
 r.ended_at, r.exit_code, r.termination_reason, r.identity_fingerprint, r.restart_count, r.operation_id,
 p.run_id, p.pid, p.process_group_id, p.executable, p.started_at, p.working_directory,
 p.identity_fingerprint, p.observed_at FROM selected r LEFT JOIN run_processes p ON p.run_id=r.id
 ORDER BY r.runtime_scope_id, r.started_at DESC, r.id DESC, p.observed_at DESC, p.pid`, args...)
	if err != nil {
		return nil, fmt.Errorf("list runtime run snapshots: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		run, identity, err := scanRunWithProcess(rows)
		if err != nil {
			return nil, err
		}
		runs := result[run.ProjectID]
		if len(runs) == 0 || runs[len(runs)-1].ID != run.ID {
			run.Processes = []domain.ProcessIdentity{}
			runs = append(runs, run)
		}
		if identity.RunID != "" {
			runs[len(runs)-1].Processes = append(runs[len(runs)-1].Processes, identity)
		}
		result[run.ProjectID] = runs
	}
	return result, rows.Err()
}
