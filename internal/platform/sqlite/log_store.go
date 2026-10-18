package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"envarium.dev/envarium/internal/foundation/identifier"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

// LogRedactor sanitizes a canonical entry before memory, disk, or subscribers receive it.
type LogRedactor interface {
	RedactLog(runtime.LogEntry) runtime.LogEntry
}

// LogStoreConfig bounds memory, segment size, retention age, and disk usage.
type LogStoreConfig struct {
	Directory      string
	SegmentBytes   int64
	RetentionAge   time.Duration
	RetentionBytes int64
}

type activeLogSegment struct {
	id, path  string
	lines     int64
	size      int64
	lastWrite time.Time
}

// LogStore is a redaction-first rotating NDJSON store with SQLite segment metadata.
type LogStore struct {
	database *Database
	config   LogStoreConfig
	redactor LogRedactor
	now      func() time.Time

	mu          sync.Mutex
	active      map[string]*activeLogSegment
	subscribers map[chan runtime.LogEntry]struct{}
	writes      int
}

// NewLogStore creates private log storage and applies retention before collection begins.
func NewLogStore(database *Database, config LogStoreConfig, redactor LogRedactor) (*LogStore, error) {
	if config.SegmentBytes <= 0 {
		config.SegmentBytes = 1 << 20
	}
	if config.RetentionAge <= 0 {
		config.RetentionAge = 7 * 24 * time.Hour
	}
	if config.RetentionBytes <= 0 {
		config.RetentionBytes = 256 << 20
	}
	if err := os.MkdirAll(config.Directory, 0o700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	store := &LogStore{database: database, config: config, redactor: redactor, now: time.Now,
		active: map[string]*activeLogSegment{}, subscribers: map[chan runtime.LogEntry]struct{}{}}
	if err := store.recoverStagedDeletions(context.Background()); err != nil {
		return nil, err
	}
	if err := store.ApplyRetention(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *LogStore) recoverStagedDeletions(ctx context.Context) error {
	rows, err := s.database.connection.QueryContext(ctx, `SELECT path FROM log_segments`)
	if err != nil {
		return fmt.Errorf("list log segment paths: %w", err)
	}
	referenced := map[string]struct{}{}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			_ = rows.Close()
			return err
		}
		referenced[path] = struct{}{}
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			if _, stagedErr := os.Stat(path + ".deleting"); stagedErr == nil {
				if renameErr := os.Rename(path+".deleting", path); renameErr != nil {
					_ = rows.Close()
					return renameErr
				}
			}
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	root, err := os.OpenRoot(s.config.Directory)
	if err != nil {
		return fmt.Errorf("open confined log root: %w", err)
	}
	defer func() { _ = root.Close() }()
	return filepath.Walk(s.config.Directory, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() || filepath.Ext(path) != ".deleting" {
			return walkErr
		}
		original := path[:len(path)-len(".deleting")]
		if _, exists := referenced[original]; !exists {
			relative, err := filepath.Rel(s.config.Directory, path)
			if err != nil {
				return err
			}
			return root.Remove(relative)
		}
		return nil
	})
}

// WriteLog redacts and deduplicates an entry before every downstream sink.
func (s *LogStore) WriteLog(ctx context.Context, entry runtime.LogEntry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = s.now().UTC()
	}
	entry = s.redactor.RedactLog(entry)
	entry.Sequence = 0
	digest, encodedSize, err := logDigest(entry)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	segment, err := s.segmentForAppend(ctx, entry, digest, encodedSize)
	if err != nil || segment == nil {
		return err
	}
	lineNumber := segment.lines + 1
	tx, err := s.database.connection.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin log append: %w", err)
	}
	rollback := func() { _ = tx.Rollback() }
	result, err := tx.ExecContext(ctx, `INSERT INTO log_entries
        (digest, segment_id, line_number, project_id, service_id, run_id, operation_id, occurred_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(digest) DO NOTHING`, digest, segment.id, lineNumber,
		entry.ProjectID, entry.ServiceID, entry.RunID, nullString(entry.OperationID), formatTime(entry.Timestamp))
	if err != nil {
		rollback()
		return fmt.Errorf("index log entry: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		rollback()
		return nil
	}
	entry.Sequence, err = result.LastInsertId()
	if err != nil {
		rollback()
		return err
	}
	encoded, err := json.Marshal(entry)
	if err != nil {
		rollback()
		return err
	}
	if err := appendPrivateLine(segment.path, encoded); err != nil {
		rollback()
		_ = os.Truncate(segment.path, segment.size)
		return err
	}
	nextLines := segment.lines + 1
	nextSize := segment.size + int64(len(encoded)+1)
	_, err = tx.ExecContext(ctx, `UPDATE log_segments SET
        first_timestamp = COALESCE(first_timestamp, ?), last_timestamp = ?,
        first_sequence = COALESCE(first_sequence, ?), last_sequence = ?, entry_count = ?, size_bytes = ? WHERE id = ?`,
		formatTime(entry.Timestamp), formatTime(entry.Timestamp), entry.Sequence, entry.Sequence, nextLines, nextSize, segment.id)
	if err != nil {
		rollback()
		_ = os.Truncate(segment.path, segment.size)
		return fmt.Errorf("update log segment: %w", err)
	}
	if err := tx.Commit(); err != nil {
		_ = os.Truncate(segment.path, segment.size)
		return fmt.Errorf("commit log append: %w", err)
	}
	segment.lines = nextLines
	segment.size = nextSize
	segment.lastWrite = s.now().UTC()
	s.publish(entry)
	s.writes++
	if s.writes%100 == 0 {
		return s.applyRetentionLocked(ctx)
	}
	return nil
}

// segmentForAppend returns nil for a replay that needs no segment allocation.
// The caller holds s.mu throughout selection and the subsequent append.
func (s *LogStore) segmentForAppend(ctx context.Context, entry runtime.LogEntry, digest string, encodedSize int) (*activeLogSegment, error) {
	key := logSegmentKey(entry)
	segment := s.active[key]
	rotate := segment != nil && segment.lines > 0 && segment.size+int64(encodedSize+1) > s.config.SegmentBytes
	// Normal appends already deduplicate atomically in the INSERT. A separate
	// lookup is needed only before allocating or rotating a segment for a replay.
	if segment == nil || rotate {
		var existing int64
		err := s.database.queryRowContext(ctx, `SELECT sequence FROM log_entries WHERE digest = ?`, digest).Scan(&existing)
		if err == nil {
			return nil, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("check log replay: %w", err)
		}
	}
	if rotate {
		if err := s.closeSegment(ctx, segment); err != nil {
			return nil, err
		}
		delete(s.active, key)
		segment = nil
	}
	if segment == nil {
		if err := s.makeSegmentRoom(ctx); err != nil {
			return nil, err
		}
		var err error
		segment, err = s.createSegment(ctx, entry)
		if err != nil {
			return nil, err
		}
		s.active[key] = segment
	}
	return segment, nil
}

func (s *LogStore) createSegment(ctx context.Context, entry runtime.LogEntry) (*activeLogSegment, error) {
	id, err := identifier.New("log")
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(s.config.Directory, safeLogPart(entry.ProjectID), safeLogPart(entry.ServiceID), safeLogPart(entry.RunID))
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, id+".ndjson")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create log segment: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	owner := entry.CatalogProjectID
	if owner == "" {
		owner = entry.ProjectID
	}
	_, err = s.database.connection.ExecContext(ctx, `INSERT INTO log_segments
        (id, project_id, service_id, run_id, operation_id, path, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, owner, entry.ServiceID, entry.RunID, nullString(entry.OperationID), path, formatTime(s.now().UTC()))
	if err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("create log segment metadata: %w", err)
	}
	return &activeLogSegment{id: id, path: path}, nil
}

func (s *LogStore) closeSegment(ctx context.Context, segment *activeLogSegment) error {
	contents, err := os.ReadFile(segment.path)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(contents)
	_, err = s.database.connection.ExecContext(ctx, `UPDATE log_segments SET closed_at = ?, sha256 = ? WHERE id = ?`,
		formatTime(s.now().UTC()), hex.EncodeToString(sum[:]), segment.id)
	return err
}

func (s *LogStore) publish(entry runtime.LogEntry) {
	for subscriber := range s.subscribers {
		select {
		case subscriber <- entry:
		default:
			delete(s.subscribers, subscriber)
			close(subscriber)
		}
	}
}

// SubscribeLogs follows redacted entries; overflow closes the stream so clients can replay by sequence.
func (s *LogStore) SubscribeLogs(buffer int) (<-chan runtime.LogEntry, func()) {
	if buffer <= 0 {
		buffer = 128
	}
	stream := make(chan runtime.LogEntry, buffer)
	s.mu.Lock()
	s.subscribers[stream] = struct{}{}
	s.mu.Unlock()
	return stream, func() {
		s.mu.Lock()
		if _, exists := s.subscribers[stream]; exists {
			delete(s.subscribers, stream)
			close(stream)
		}
		s.mu.Unlock()
	}
}

// Close finalizes checksums and releases live subscribers.
func (s *LogStore) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result error
	for _, segment := range s.active {
		if err := s.closeSegment(ctx, segment); err != nil {
			result = fmt.Errorf("close log segment: %w", err)
		}
	}
	for subscriber := range s.subscribers {
		close(subscriber)
	}
	s.active = map[string]*activeLogSegment{}
	s.subscribers = map[chan runtime.LogEntry]struct{}{}
	return result
}

func appendPrivateLine(path string, encoded []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log segment: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return fmt.Errorf("append log segment: %w", err)
	}
	return nil
}

func logDigest(entry runtime.LogEntry) (string, int, error) {
	encoded, err := json.Marshal(entry)
	if err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), len(encoded), nil
}

func logSegmentKey(entry runtime.LogEntry) string {
	return entry.ProjectID + "\x00" + entry.ServiceID + "\x00" + entry.RunID + "\x00" + entry.OperationID
}

var unsafeLogPart = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func safeLogPart(value string) string {
	value = unsafeLogPart.ReplaceAllString(value, "_")
	if value == "" || value == "." || value == ".." {
		return "unknown"
	}
	return value
}
