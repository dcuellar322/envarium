package process

import (
	"context"
	"fmt"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

func (d *Driver) killAll(ctx context.Context, plan executionPlan, sink domain.ProgressSink) error {
	runs, err := d.store.ListProjectRuns(ctx, plan.project.ProjectID)
	if err != nil {
		return err
	}
	active := make(map[string]domain.RunRecord)
	for _, run := range runs {
		if run.EndedAt == nil {
			active[run.ServiceID] = run
		}
	}
	for _, service := range stoppingServices(plan, active) {
		run, ok := active[service.service.ID]
		if !ok {
			if err := sink.Step(ctx, "process.kill."+service.service.ID, "succeeded", "already stopped"); err != nil {
				return err
			}
			continue
		}
		if err := d.killRun(ctx, run); err != nil {
			return err
		}
		if err := sink.Step(ctx, "process.kill."+service.service.ID, "succeeded", "process group force-stopped"); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) killRun(ctx context.Context, run domain.RunRecord) error {
	verified, err := d.verifiedMembersWithGrace(ctx, run, identityHandoffGrace)
	if err != nil {
		return err
	}
	if len(verified) == 0 {
		return d.store.FinishRun(ctx, run.ID, d.now().UTC(), nil, "identity_lost")
	}
	d.mu.RLock()
	managed := d.managed[serviceKey(run.ProjectID, run.ServiceID)]
	d.mu.RUnlock()
	if managed != nil {
		managed.mu.Lock()
		managed.stopping = true
		managed.mu.Unlock()
	}
	groups := uniqueGroups(verified)
	for _, group := range groups {
		if err := signalManagedGroup(managed, group, true); err != nil && d.managedGroupStillRunning(ctx, managed, group) {
			return fmt.Errorf("force-stop process group %d: %w", group, err)
		}
	}
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for anyGroupRunning(groups, func(group int32) bool {
		return d.managedGroupStillRunning(context.Background(), managed, group)
	}) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return errorsForKillTimeout(groups)
		case <-ticker.C:
		}
	}
	return d.finishStoppedRun(ctx, run, managed, "killed")
}

func errorsForKillTimeout(groups []int32) error {
	return fmt.Errorf("process groups did not terminate after force-stop: %v", groups)
}
