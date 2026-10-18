package sqlite

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

func TestLogReplayAndCancellationDoNotChangeSegmentsOrPublish(t *testing.T) {
	for _, segmentBytes := range []int64{1, 1 << 20} {
		t.Run(strconv.FormatInt(segmentBytes, 10), func(t *testing.T) {
			database, store, root := newTestLogStore(t, LogStoreConfig{SegmentBytes: segmentBytes})
			stream, unsubscribe := store.SubscribeLogs(4)
			defer unsubscribe()
			entry := runtime.LogEntry{Timestamp: time.Now().UTC(), ProjectID: "project-1", ServiceID: "api", RunID: "run", Message: "one"}
			if err := store.WriteLog(context.Background(), entry); err != nil {
				t.Fatal(err)
			}
			first := <-stream
			before, err := readAllLogFiles(root)
			if err != nil {
				t.Fatal(err)
			}
			if err := store.WriteLog(context.Background(), entry); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			entry.Message = "canceled"
			if err := store.WriteLog(ctx, entry); !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled append=%v", err)
			}
			after, err := readAllLogFiles(root)
			if err != nil || after != before {
				t.Fatalf("segment changed after replay/cancellation: %v", err)
			}
			var segments int
			if err := database.connection.QueryRow(`SELECT COUNT(*) FROM log_segments`).Scan(&segments); err != nil || segments != 1 {
				t.Fatalf("segments=%d err=%v", segments, err)
			}
			entries, err := store.QueryLogs(context.Background(), observability.LogQuery{ProjectID: "project-1", Limit: 10})
			if err != nil || len(entries) != 1 || entries[0].Sequence != first.Sequence {
				t.Fatalf("persisted entries=%#v err=%v", entries, err)
			}
			select {
			case extra := <-stream:
				t.Fatalf("unexpected live publication: %#v", extra)
			default:
			}
		})
	}
}
