package process

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

type managedRun struct {
	mu            sync.Mutex
	run           domain.RunRecord
	project       domain.ProjectRuntime
	service       servicePlan
	command       *exec.Cmd
	group         int32
	ownership     processOwnership
	restartCancel context.CancelFunc
	restartDone   chan struct{}
	stopping      bool
	logs          *logBuffer
}

func (d *Driver) startAll(ctx context.Context, plan executionPlan, sink domain.ProgressSink) error {
	started := []*managedRun{}
	for _, service := range plan.services {
		if err := ctx.Err(); err != nil {
			d.rollbackStarts(started)
			return err
		}
		managed, created, err := d.startService(ctx, plan.project, service)
		if err != nil {
			d.rollbackStarts(started)
			return err
		}
		if created {
			started = append(started, managed)
		}
		message := "already owned by Envarium; running or awaiting supervised restart"
		if created {
			managed.mu.Lock()
			pid := managed.run.Processes[0].PID
			managed.mu.Unlock()
			message = fmt.Sprintf("started %s as PID %d", service.service.ID, pid)
		}
		if err := sink.Step(ctx, "process.start."+service.service.ID, "succeeded", message); err != nil {
			d.rollbackStarts(started)
			return err
		}
	}
	return nil
}

func (d *Driver) rollbackStarts(runs []*managedRun) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for index := len(runs) - 1; index >= 0; index-- {
		runs[index].mu.Lock()
		run := runs[index].run
		runs[index].mu.Unlock()
		_ = d.stopRun(ctx, run, time.Second, "start_cancelled")
	}
}
