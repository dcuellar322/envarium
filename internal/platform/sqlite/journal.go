package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"envarium.dev/envarium/internal/foundation/events"
	"envarium.dev/envarium/internal/platform/sqlite/generated"
)

// Journal is a durable SQLite event journal with in-process fan-out.
type Journal struct {
	queries    *generated.Queries
	connection *sql.DB
	pruned     bool
	mu         sync.Mutex
	nextID     uint64
	live       map[uint64]chan events.Envelope
}

// NewJournal creates a journal over database.
func NewJournal(database *Database) *Journal {
	return &Journal{queries: database.queries, connection: database.connection, live: make(map[uint64]chan events.Envelope)}
}

// Publish persists an event before broadcasting it.
func (j *Journal) Publish(ctx context.Context, event events.Envelope) (events.Envelope, error) {
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	tx, err := j.connection.BeginTx(ctx, nil)
	if err != nil {
		return events.Envelope{}, err
	}
	defer func() { _ = tx.Rollback() }()
	sequence, err := j.queries.WithTx(tx).CreateJournalEvent(ctx, generated.CreateJournalEventParams{
		ID: event.ID, Type: event.Type, OccurredAt: formatTime(event.OccurredAt),
		ProjectID: nullable(event.ProjectID), OperationID: nullable(event.OperationID),
		PayloadJson: string(event.Payload),
	})
	if err != nil {
		return events.Envelope{}, fmt.Errorf("insert journal event: %w", err)
	}
	event.Sequence = sequence
	if !j.pruned || sequence%1024 == 0 {
		// Sequence is monotonic even after deletion; this indexed range avoids a
		// table count or a full history scan on the monitoring hot path.
		_, pruneErr := tx.ExecContext(ctx, "DELETE FROM event_journal WHERE sequence <= ?", sequence-100000)
		if pruneErr != nil {
			return events.Envelope{}, fmt.Errorf("prune replay journal: %w", pruneErr)
		}
	}
	if err := tx.Commit(); err != nil {
		return events.Envelope{}, err
	}
	j.pruned = true
	for id, subscriber := range j.live {
		select {
		case subscriber <- event:
		default:
			close(subscriber)
			delete(j.live, id)
		}
	}
	return event, nil
}

// Replay returns at most limit events and reports when the caller must refresh.
func (j *Journal) Replay(ctx context.Context, after int64, limit int) ([]events.Envelope, bool, error) {
	if limit < 1 {
		limit = 1
	}
	records, err := j.queries.ListJournalEventsAfter(ctx, generated.ListJournalEventsAfterParams{
		Sequence: after, Limit: int64(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("replay journal events: %w", err)
	}
	var oldest sql.NullInt64
	if err := j.connection.QueryRowContext(ctx, "SELECT MIN(sequence) FROM event_journal").Scan(&oldest); err != nil {
		return nil, false, err
	}
	stale := oldest.Valid && after < oldest.Int64-1
	truncated := len(records) > limit
	if truncated {
		records = records[:limit]
	}
	result := make([]events.Envelope, 0, len(records))
	for _, record := range records {
		occurredAt, err := parseTime(record.OccurredAt)
		if err != nil {
			return nil, false, err
		}
		result = append(result, events.Envelope{
			ID: record.ID, Type: record.Type, OccurredAt: occurredAt,
			Sequence: record.Sequence, ProjectID: record.ProjectID.String,
			OperationID: record.OperationID.String, Payload: json.RawMessage(record.PayloadJson),
		})
	}
	return result, truncated || stale, nil
}

// Subscribe returns a bounded live event stream and idempotent cancellation.
func (j *Journal) Subscribe(buffer int) (<-chan events.Envelope, func()) {
	if buffer < 1 {
		buffer = 1
	}
	j.mu.Lock()
	id := j.nextID
	j.nextID++
	stream := make(chan events.Envelope, buffer)
	j.live[id] = stream
	j.mu.Unlock()
	return stream, func() {
		j.mu.Lock()
		if existing, ok := j.live[id]; ok {
			close(existing)
			delete(j.live, id)
		}
		j.mu.Unlock()
	}
}
