package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

func TestLogArchiveBoundsActiveSegmentsAndRetainsReplay(t *testing.T) {
	database, store, _ := newTestLogStore(t, LogStoreConfig{RetentionBytes: 10 << 20})
	ctx := context.Background()
	for index := range maximumActiveLogSegments + 20 {
		entry := runtime.LogEntry{Timestamp: time.Now(), ProjectID: "project-1", ServiceID: "api", RunID: fmt.Sprint(index), Message: "line"}
		if err := store.WriteLog(ctx, entry); err != nil {
			t.Fatal(err)
		}
		if len(store.active) > maximumActiveLogSegments {
			t.Fatal("unbounded active run segments")
		}
	}
	entries, err := store.QueryLogs(ctx, observability.LogQuery{ProjectID: "project-1", Limit: 1000})
	if err != nil || len(entries) != maximumActiveLogSegments+20 {
		t.Fatalf("durable entries=%d error=%v", len(entries), err)
	}
	var closed int
	if err := database.connection.QueryRow(`SELECT COUNT(*) FROM log_segments WHERE closed_at IS NOT NULL AND length(sha256)=64`).Scan(&closed); err != nil {
		t.Fatal(err)
	}
	if closed < 20 {
		t.Fatal("evicted append segments were not finalized")
	}
}

func TestLogArchiveReleasesIdleSegmentsAndResumes(t *testing.T) {
	_, store, _ := newTestLogStore(t, LogStoreConfig{RetentionAge: time.Hour})
	ctx := context.Background()
	now := time.Now().UTC()
	store.now = func() time.Time { return now }
	entry := runtime.LogEntry{Timestamp: now, ProjectID: "project-1", ServiceID: "api", RunID: "run", Message: "first"}
	if err := store.WriteLog(ctx, entry); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if err := store.ApplyRetention(ctx); err != nil {
		t.Fatal(err)
	}
	if len(store.active) != 0 {
		t.Fatal("idle append state retained")
	}
	entry.Timestamp = now
	entry.Message = "resumed"
	if err := store.WriteLog(ctx, entry); err != nil {
		t.Fatal(err)
	}
	entries, err := store.QueryLogs(ctx, observability.LogQuery{ProjectID: "project-1", Limit: 10})
	if err != nil || len(entries) != 2 || entries[0].Sequence >= entries[1].Sequence {
		t.Fatalf("resumed replay=%+v error=%v", entries, err)
	}
	now = now.Add(2 * time.Hour)
	if err := store.ApplyRetention(ctx); err != nil {
		t.Fatal(err)
	}
	entries, err = store.QueryLogs(ctx, observability.LogQuery{ProjectID: "project-1", Limit: 10})
	if err != nil || len(entries) != 0 || len(store.active) != 0 {
		t.Fatal("expired idle runs escaped retention")
	}
}
