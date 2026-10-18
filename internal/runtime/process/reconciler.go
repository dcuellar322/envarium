package process

import (
	"context"
	"os/exec"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

func (d *Driver) monitor(managed *managedRun, command *exec.Cmd, environment []string) {
	trackingDone := make(chan struct{})
	go d.trackGroup(managed, trackingDone)
	err := command.Wait()
	close(trackingDone)
	exitCode := command.ProcessState.ExitCode()
	d.recordGroupMembers(context.Background(), managed)
	managed.mu.Lock()
	stopping := managed.stopping
	current := managed.command == command
	restartCount := managed.run.RestartCount
	managed.mu.Unlock()
	if stopping || !current {
		return
	}
	if d.managedGroupStillRunning(context.Background(), managed, managed.group) {
		d.waitForOrphanedGroup(managed, exitCode)
		return
	}
	if exitCode != 0 && managed.service.definition.Restart.Mode == "on-failure" &&
		restartCount < managed.service.definition.Restart.MaxRetries && d.ctx.Err() == nil {
		d.restartAfterCrash(managed, environment, exitCode)
		return
	}
	reason := "exited"
	if err != nil || exitCode != 0 {
		reason = "crashed"
	}
	d.finishManaged(managed, &exitCode, reason)
}

func (d *Driver) restartAfterCrash(managed *managedRun, environment []string, priorExit int) {
	managed.mu.Lock()
	if managed.stopping || managed.run.EndedAt != nil {
		managed.mu.Unlock()
		return
	}
	//nolint:gosec // G118: the supervised restart owns cancel until completion; stop cancels and joins it.
	restartCtx, cancel := context.WithCancel(d.ctx)
	done := make(chan struct{})
	managed.restartCancel, managed.restartDone = cancel, done
	managed.mu.Unlock()
	defer func() {
		cancel()
		managed.mu.Lock()
		if managed.restartDone == done {
			managed.restartCancel = nil
			managed.restartDone = nil
		}
		managed.mu.Unlock()
		close(done)
	}()
	backoff := time.NewTimer(time.Duration(managed.service.definition.Restart.BackoffSeconds) * time.Second)
	defer backoff.Stop()
	select {
	case <-restartCtx.Done():
		d.finishRestartFailure(managed, priorExit, "crashed")
		return
	case <-backoff.C:
	}
	managed.mu.Lock()
	if managed.stopping || restartCtx.Err() != nil {
		managed.mu.Unlock()
		return
	}
	managed.run.RestartCount++
	restartCount := managed.run.RestartCount
	managed.mu.Unlock()
	if err := d.store.SetRestartCount(restartCtx, managed.run.ID, restartCount); err != nil {
		d.finishRestartFailure(managed, priorExit, "restart_persistence_failed")
		return
	}
	managed.mu.Lock()
	priorOwnership := managed.ownership
	managed.mu.Unlock()
	closeOwnership(priorOwnership)
	command, identity, ownership, err := d.launch(restartCtx, managed, environment)
	if err != nil {
		d.finishRestartFailure(managed, priorExit, "restart_failed")
		return
	}
	if err := d.store.RecordProcess(restartCtx, identity); err != nil {
		abortOwnedProcess(command, ownership, identity.ProcessGroup)
		d.finishRestartFailure(managed, priorExit, "restart_persistence_failed")
		return
	}
	managed.mu.Lock()
	if managed.stopping || restartCtx.Err() != nil {
		managed.mu.Unlock()
		abortOwnedProcess(command, ownership, identity.ProcessGroup)
		return
	}
	managed.command = command
	managed.group = identity.ProcessGroup
	managed.ownership = ownership
	managed.run.Processes = append(managed.run.Processes, identity)
	managed.mu.Unlock()
	d.emit(managed.project.ProjectID, domain.RuntimeEvent{
		Driver: domain.KindProcess, ProjectIdentity: managed.project.ProjectSlug,
		ServiceIdentity: managed.service.service.ID, RunID: managed.run.ID,
		Action: "restart_after_crash", OccurredAt: d.now().UTC(),
	})
	go d.monitor(managed, command, environment)
}

func (d *Driver) waitForOrphanedGroup(managed *managedRun, exitCode int) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		managed.mu.Lock()
		stopping := managed.stopping
		managed.mu.Unlock()
		if stopping {
			return
		}
		if !d.managedGroupStillRunning(context.Background(), managed, managed.group) {
			d.finishManaged(managed, &exitCode, "process_tree_exited")
			return
		}
		d.recordGroupMembers(context.Background(), managed)
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *Driver) trackGroup(managed *managedRun, done <-chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	started := time.Now()
	handoff := true
	for {
		select {
		case <-done:
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.recordGroupMembers(context.Background(), managed)
			if handoff && time.Since(started) >= identityHandoffGrace {
				handoff = false
				ticker.Reset(time.Second)
			}
		}
	}
}

func (d *Driver) recordGroupMembers(ctx context.Context, managed *managedRun) {
	managed.mu.Lock()
	group := managed.group
	runID := managed.run.ID
	ownership := managed.ownership
	managed.mu.Unlock()
	members, err := d.groupMembers(ctx, ownership, group)
	if err != nil {
		return
	}
	for _, member := range members {
		member.RunID = runID
		managed.mu.Lock()
		known := containsIdentity(managed.run.Processes, member)
		managed.mu.Unlock()
		if known {
			continue
		}
		if d.store.RecordProcess(ctx, member) == nil {
			managed.mu.Lock()
			if !containsIdentity(managed.run.Processes, member) {
				managed.run.Processes = append(managed.run.Processes, member)
			}
			managed.mu.Unlock()
		}
	}
}

func (d *Driver) groupMembers(
	ctx context.Context,
	ownership processOwnership,
	group int32,
) ([]domain.ProcessIdentity, error) {
	source, ok := ownership.(processMemberSource)
	if !ok {
		return d.inspector.GroupMembers(ctx, group)
	}
	pids, err := source.MemberPIDs()
	if err != nil {
		return nil, err
	}
	members := make([]domain.ProcessIdentity, 0, len(pids))
	for _, pid := range pids {
		identity, snapshotErr := d.inspector.Snapshot(ctx, pid)
		if snapshotErr != nil {
			continue
		}
		identity.ProcessGroup = group
		members = append(members, identity)
	}
	return members, nil
}

func containsIdentity(values []domain.ProcessIdentity, candidate domain.ProcessIdentity) bool {
	for _, value := range values {
		if value.PID == candidate.PID && value.StartedAt.Equal(candidate.StartedAt) && value.Fingerprint == candidate.Fingerprint {
			return true
		}
	}
	return false
}

func (d *Driver) groupStillRunning(ctx context.Context, group int32) bool {
	members, err := d.inspector.GroupMembers(ctx, group)
	return err == nil && len(members) > 0
}

func (d *Driver) managedGroupStillRunning(ctx context.Context, managed *managedRun, group int32) bool {
	if managed != nil {
		managed.mu.Lock()
		ownership := managed.ownership
		managedGroup := managed.group
		managed.mu.Unlock()
		if ownership != nil && managedGroup == group {
			return ownership.Running()
		}
	}
	return d.groupStillRunning(ctx, group)
}

func (d *Driver) finishManaged(managed *managedRun, exitCode *int, reason string) {
	managed.mu.Lock()
	if managed.run.EndedAt != nil {
		managed.mu.Unlock()
		return
	}
	endedAt := d.now().UTC()
	managed.run.EndedAt = &endedAt
	managed.run.ExitCode = exitCode
	managed.run.TerminationReason = reason
	ownership := managed.ownership
	managed.ownership = nil
	managed.mu.Unlock()
	closeOwnership(ownership)
	_ = d.store.FinishRun(context.Background(), managed.run.ID, endedAt, exitCode, reason)
	d.mu.Lock()
	delete(d.managed, serviceKey(managed.project.ProjectID, managed.service.service.ID))
	d.mu.Unlock()
	d.emit(managed.project.ProjectID, domain.RuntimeEvent{
		Driver: domain.KindProcess, ProjectIdentity: managed.project.ProjectSlug,
		ServiceIdentity: managed.service.service.ID, RunID: managed.run.ID,
		Action: reason, OccurredAt: endedAt,
	})
}
