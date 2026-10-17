package compose

import (
	"context"
	"errors"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
	apiEvents "github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
)

func TestIdleComposeLogFollowWaitsAndCancels(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine := &fakeEngine{events: client.EventsResult{Messages: make(chan apiEvents.Message), Err: make(chan error)}}
	driver := &Driver{engine: fakeConnector{engine: engine}}
	done := make(chan error, 1)
	go func() {
		done <- driver.streamLogs(ctx, domain.LogRequest{Follow: true}, normalizedConfig{ProjectName: "fixture"}, &recordingLogSink{})
	}()
	select {
	case err := <-done:
		t.Fatalf("idle follow returned immediately: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("idle follow ignored cancellation")
	}
	// Snapshot reads remain finite when a project has no containers.
	if err := driver.streamLogs(context.Background(), domain.LogRequest{}, normalizedConfig{ProjectName: "fixture"}, &recordingLogSink{}); err != nil {
		t.Fatal(err)
	}
}

func TestIdleComposeLogsWakeOnProjectEventAndReportDisconnect(t *testing.T) {
	t.Parallel()
	for _, failure := range []bool{false, true} {
		messages := make(chan apiEvents.Message, 2)
		failures := make(chan error, 1)
		engine := &fakeEngine{events: client.EventsResult{Messages: messages, Err: failures}}
		expected := errors.New("engine disconnected")
		if failure {
			failures <- expected
		} else {
			messages <- apiEvents.Message{Actor: apiEvents.Actor{Attributes: map[string]string{labelProject: "other"}}}
			messages <- apiEvents.Message{Actor: apiEvents.Actor{Attributes: map[string]string{labelProject: "fixture"}}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := waitForLogChange(ctx, engine, "fixture", time.Now().Unix())
		cancel()
		if failure && !errors.Is(err, expected) || !failure && err != nil {
			t.Fatalf("wake/disconnect: %v", err)
		}
	}
}
