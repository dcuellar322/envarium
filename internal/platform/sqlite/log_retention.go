package sqlite

import (
	"context"
	"fmt"
	"os"
	"time"
)

type retainedSegment struct {
	id, path, created string
	size              int64
}

// ApplyRetention enforces both age and disk caps while never deleting an active segment.
func (s *LogStore) ApplyRetention(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyRetentionLocked(ctx)
}

const maximumActiveLogSegments = 128

func (s *LogStore) makeSegmentRoom(ctx context.Context) error {
	if len(s.active) < maximumActiveLogSegments {
		return nil
	}
	var oldestKey string
	var oldest *activeLogSegment
	for key, segment := range s.active {
		if oldest == nil || segment.lastWrite.Before(oldest.lastWrite) {
			oldestKey, oldest = key, segment
		}
	}
	if err := s.closeSegment(ctx, oldest); err != nil {
		return err
	}
	delete(s.active, oldestKey)
	return nil
}

func (s *LogStore) applyRetentionLocked(ctx context.Context) error {
	// Closing an append segment is safe even if a producer later resumes: the
	// next append opens a new segment without disrupting sequence-based replay.
	cutoffIdle := s.now().UTC().Add(-min(time.Minute, s.config.RetentionAge))
	for key, segment := range s.active {
		if !segment.lastWrite.After(cutoffIdle) {
			if err := s.closeSegment(ctx, segment); err != nil {
				return err
			}
			delete(s.active, key)
		}
	}
	rows, err := s.database.connection.QueryContext(ctx, `SELECT id, path, created_at, size_bytes FROM log_segments ORDER BY created_at, id`)
	if err != nil {
		return fmt.Errorf("list retained log segments: %w", err)
	}
	var segments []retainedSegment
	var total int64
	for rows.Next() {
		var item retainedSegment
		if err := rows.Scan(&item.id, &item.path, &item.created, &item.size); err != nil {
			_ = rows.Close()
			return err
		}
		segments = append(segments, item)
		total += item.size
	}
	if err := rows.Close(); err != nil {
		return err
	}
	active := map[string]struct{}{}
	for _, segment := range s.active {
		active[segment.id] = struct{}{}
	}
	cutoff := s.now().UTC().Add(-s.config.RetentionAge)
	for _, segment := range segments {
		if _, exists := active[segment.id]; exists {
			continue
		}
		created, parseErr := parseTime(segment.created)
		if parseErr != nil {
			return parseErr
		}
		if created.Before(cutoff) || total > s.config.RetentionBytes {
			if err := s.deleteSegment(ctx, segment); err != nil {
				return err
			}
			total -= segment.size
		}
	}
	return nil
}

func (s *LogStore) deleteSegment(ctx context.Context, segment retainedSegment) error {
	staged := segment.path + ".deleting"
	if err := os.Rename(segment.path, staged); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stage retained log segment deletion: %w", err)
	}
	if _, err := s.database.connection.ExecContext(ctx, `DELETE FROM log_segments WHERE id = ?`, segment.id); err != nil {
		_ = os.Rename(staged, segment.path)
		return fmt.Errorf("delete retained log metadata: %w", err)
	}
	if err := os.Remove(staged); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove retained log segment: %w", err)
	}
	return nil
}
