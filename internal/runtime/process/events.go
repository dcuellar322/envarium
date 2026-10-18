package process

import (
	"context"
	"strconv"
	"strings"

	"envarium.dev/envarium/internal/runtime/domain"
)

func (d *Driver) watchEvents(ctx context.Context, project domain.ProjectRuntime, sink domain.EventSink) error {
	channel := make(chan domain.RuntimeEvent, 32)
	d.mu.Lock()
	if d.subscribers[project.ProjectID] == nil {
		d.subscribers[project.ProjectID] = make(map[chan domain.RuntimeEvent]struct{})
	}
	d.subscribers[project.ProjectID][channel] = struct{}{}
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		delete(d.subscribers[project.ProjectID], channel)
		if len(d.subscribers[project.ProjectID]) == 0 {
			delete(d.subscribers, project.ProjectID)
		}
		d.mu.Unlock()
	}()
	snapshots, unsubscribe := d.subscribeRunSnapshots(project.ProjectID)
	defer unsubscribe()
	lastSignature := ""
	initial := true
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-channel:
			if err := sink.WriteRuntimeEvent(ctx, event); err != nil {
				return err
			}
			if observation, err := d.inspect(ctx, project); err == nil {
				lastSignature = observationSignature(observation)
				initial = false
			}
		case snapshot, ok := <-snapshots:
			if !ok {
				return context.Canceled
			}
			if snapshot.err != nil {
				return snapshot.err
			}
			observation, err := d.inspectRuns(ctx, project, snapshot.runs)
			if err != nil {
				return err
			}
			signature := observationSignature(observation)
			if initial {
				lastSignature = signature
				initial = false
				continue
			}
			if signature == lastSignature {
				continue
			}
			lastSignature = signature
			if err := sink.WriteRuntimeEvent(ctx, domain.RuntimeEvent{
				Driver: domain.KindProcess, ProjectIdentity: project.ProjectSlug,
				Action: "reconcile", OccurredAt: d.now().UTC(),
			}); err != nil {
				return err
			}
		}
	}
}

func observationSignature(observation domain.Observation) string {
	var builder strings.Builder
	builder.WriteString(string(observation.State))
	builder.WriteByte('|')
	builder.WriteString(string(observation.Origin))
	for _, service := range observation.Services {
		builder.WriteByte('|')
		builder.WriteString(service.ID)
		builder.WriteByte(':')
		builder.WriteString(service.State)
		if service.Process != nil {
			builder.WriteByte(':')
			builder.WriteString(strconv.FormatInt(int64(service.Process.PID), 10))
			builder.WriteByte(':')
			builder.WriteString(service.Process.Fingerprint)
		}
	}
	return builder.String()
}

func (d *Driver) emit(projectID string, event domain.RuntimeEvent) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for subscriber := range d.subscribers[projectID] {
		select {
		case subscriber <- event:
		default:
		}
	}
}
