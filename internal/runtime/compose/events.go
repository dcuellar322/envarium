package compose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
	"github.com/moby/moby/client"
)

func (d *Driver) watchEvents(ctx context.Context, config normalizedConfig, sink domain.EventSink) error {
	engine, _, err := d.engine.Connect(ctx, config.Connection)
	if err != nil {
		return err
	}
	defer func() { _ = engine.Close() }()
	filters := projectFilters(config.ProjectName).Add("type", "container")
	events := engine.Events(ctx, client.EventsListOptions{Filters: filters})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-events.Err:
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("watch Docker events: %w", err)
		case event, ok := <-events.Messages:
			if !ok {
				return nil
			}
			if event.Actor.Attributes[labelProject] != config.ProjectName || !runtimeRelevantAction(string(event.Action)) {
				continue
			}
			occurred := time.Unix(0, event.TimeNano).UTC()
			if event.TimeNano == 0 {
				occurred = time.Unix(event.Time, 0).UTC()
			}
			if err := sink.WriteRuntimeEvent(ctx, domain.RuntimeEvent{
				Driver: domain.KindCompose, ProjectIdentity: config.ProjectName,
				ServiceIdentity: event.Actor.Attributes[labelService], ContainerID: event.Actor.ID,
				Action: string(event.Action), OccurredAt: occurred,
			}); err != nil {
				return err
			}
		}
	}
}

// Exec lifecycle events include periodic health-check command execution. They
// change no container lifecycle/identity fact; health_status still reconciles.
// Unknown future events remain observable rather than being silently dropped.
func runtimeRelevantAction(action string) bool {
	kind, _, _ := strings.Cut(action, ":")
	switch strings.TrimSpace(kind) {
	case "exec_create", "exec_start", "exec_die":
		return false
	default:
		return true
	}
}
