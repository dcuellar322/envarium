package process

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
)

// Supervisors, rather than an observation racing command.Wait, finish owned runs.
func (d *Driver) supervisingRun(run domain.RunRecord) bool {
	d.mu.RLock()
	managed := d.managed[serviceKey(run.ProjectID, run.ServiceID)]
	d.mu.RUnlock()
	if managed == nil {
		return false
	}
	managed.mu.Lock()
	defer managed.mu.Unlock()
	return managed.run.ID == run.ID && managed.run.EndedAt == nil
}

// Cancel and join a pending restart before checking/signal-fencing its processes.
func (d *Driver) prepareManagedStop(ctx context.Context, run domain.RunRecord) (*managedRun, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	d.mu.RLock()
	managed := d.managed[serviceKey(run.ProjectID, run.ServiceID)]
	d.mu.RUnlock()
	if managed == nil {
		return nil, nil
	}
	managed.mu.Lock()
	if managed.run.ID != run.ID {
		managed.mu.Unlock()
		return nil, nil
	}
	managed.stopping = true
	cancel, done := managed.restartCancel, managed.restartDone
	managed.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-ctx.Done():
			return managed, ctx.Err()
		case <-done:
		}
	}
	return managed, nil
}

func (d *Driver) finishRestartFailure(managed *managedRun, exit int, reason string) {
	managed.mu.Lock()
	stopping := managed.stopping
	managed.mu.Unlock()
	if !stopping {
		d.finishManaged(managed, &exit, reason)
	}
}
