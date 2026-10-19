package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestDesktopStreamReadsFreshBoundedSnapshots(t *testing.T) {
	var output bytes.Buffer
	reads := 0
	err := streamDesktopSnapshots(t.Context(), strings.NewReader("snapshot\nsnapshot\n"), &output, func(ctx context.Context) (desktopPresentation, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 15*time.Second {
			t.Fatal("missing per-snapshot deadline")
		}
		reads++
		return desktopPresentation{PortConflictCount: reads}, nil
	})
	if err != nil || reads != 2 {
		t.Fatalf("reads=%d err=%v", reads, err)
	}
	decoder := json.NewDecoder(&output)
	for count := 1; count <= 2; count++ {
		var frame struct {
			SchemaVersion string              `json:"schemaVersion"`
			Command       string              `json:"command"`
			Data          desktopPresentation `json:"data"`
		}
		if err := decoder.Decode(&frame); err != nil || frame.Command != "desktop.snapshot" || frame.SchemaVersion != cliSchemaVersion || frame.Data.PortConflictCount != count {
			t.Fatalf("frame=%+v err=%v", frame, err)
		}
	}
}

func TestDesktopStreamRejectsUnknownAndUnboundedRequests(t *testing.T) {
	for _, input := range []string{"start project\n", "snapshot extra\n", strings.Repeat("x", 33)} {
		t.Run(input[:5], func(t *testing.T) {
			err := streamDesktopSnapshots(t.Context(), strings.NewReader(input), io.Discard, func(context.Context) (desktopPresentation, error) {
				t.Fatal("invalid request reached application reads")
				return desktopPresentation{}, nil
			})
			if err == nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
}

func TestDesktopStreamStopsOnReadOrOutputFailure(t *testing.T) {
	failed := errors.New("offline")
	reads := 0
	var output bytes.Buffer
	err := streamDesktopSnapshots(t.Context(), strings.NewReader("snapshot\nsnapshot\n"), &output, func(context.Context) (desktopPresentation, error) {
		reads++
		return desktopPresentation{}, failed
	})
	if !errors.Is(err, failed) || reads != 1 || output.Len() != 0 {
		t.Fatalf("read failure produced output: %d reads, %v", reads, err)
	}
	err = streamDesktopSnapshots(t.Context(), strings.NewReader("snapshot\n"), &output, func(context.Context) (desktopPresentation, error) {
		return desktopPresentation{Warnings: []string{strings.Repeat("x", desktopResponseLimit)}}, nil
	})
	if err == nil || output.Len() != 0 {
		t.Fatal("oversized frame was written")
	}
	reader, writer := io.Pipe()
	_ = reader.Close()
	err = streamDesktopSnapshots(t.Context(), strings.NewReader("snapshot\n"), writer, func(context.Context) (desktopPresentation, error) {
		return desktopPresentation{}, nil
	})
	_ = writer.Close()
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("output failure = %v", err)
	}
}

func TestDesktopStreamCancellationClosesIdleInputAndCancelsRead(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	reader, writer := io.Pipe()
	t.Cleanup(func() { _ = writer.Close() })
	done := make(chan error, 1)
	go func() {
		done <- streamDesktopSnapshots(ctx, reader, io.Discard, func(context.Context) (desktopPresentation, error) {
			t.Error("idle bridge performed a read")
			return desktopPresentation{}, nil
		})
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("idle cancellation = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("idle bridge did not cancel")
	}
	ctx, cancel = context.WithCancel(t.Context())
	err := streamDesktopSnapshots(ctx, strings.NewReader("snapshot\n"), io.Discard, func(ctx context.Context) (desktopPresentation, error) {
		cancel()
		<-ctx.Done()
		return desktopPresentation{}, ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("active cancellation = %v", err)
	}
}
