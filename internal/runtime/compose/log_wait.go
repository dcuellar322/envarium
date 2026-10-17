package compose

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/moby/moby/client"
)

// Idle follow streams wait for a container event instead of reopening every
// second. A bounded refresh still notices config/context changes and recovers
// from missing events. Since covers the list-to-subscribe race.
func waitForLogChange(ctx context.Context, engine engineClient, project string, since int64) error {
	eventCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	events := engine.Events(eventCtx, client.EventsListOptions{
		Filters: projectFilters(project).Add("type", "container").Add("event", "create", "start", "restart", "die", "destroy"),
		Since:   strconv.FormatInt(since, 10),
	})
	timer := time.NewTimer(observationConfigTTL)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		case err, ok := <-events.Err:
			if !ok || err == nil {
				return io.EOF
			}
			return fmt.Errorf("wait for Compose log source: %w", err)
		case event, ok := <-events.Messages:
			if !ok {
				return io.EOF
			}
			if event.Actor.Attributes[labelProject] == project {
				return nil
			}
		}
	}
}
