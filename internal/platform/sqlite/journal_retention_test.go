package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"envarium.dev/envarium/internal/foundation/events"
)

func TestJournalPrunesReplayWindowAndSignalsOldCursor(t *testing.T) {
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "envarium.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	// Explicit positions exercise the horizon without 100,000 slow test inserts.
	for _, sequence := range []int{1, 2, 100003} {
		if _, err := database.connection.ExecContext(ctx, `INSERT INTO event_journal(sequence,id,type,occurred_at,payload_json) VALUES(?,?, 'test', ?, '{}')`, sequence, sequence, formatTime(time.Now())); err != nil {
			t.Fatal(err)
		}
	}
	journal := NewJournal(database)
	if _, err := journal.Publish(ctx, events.Envelope{ID: "fresh", Type: "test", OccurredAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	replayed, refresh, err := journal.Replay(ctx, 2, 10)
	if err != nil || !refresh || len(replayed) != 2 {
		t.Fatalf("events=%#v refresh=%v err=%v", replayed, refresh, err)
	}
	if _, refresh, err := journal.Replay(ctx, 100003, 10); err != nil || refresh {
		t.Fatalf("recent refresh=%v err=%v", refresh, err)
	}
	var count int
	if err := database.connection.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_journal WHERE sequence <= 4").Scan(&count); err != nil || count != 0 {
		t.Fatalf("old count=%d err=%v", count, err)
	}
}
