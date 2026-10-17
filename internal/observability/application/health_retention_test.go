package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

type pruningHealthRepository struct {
	healthRepositoryStub
	calls  int
	cutoff time.Time
	err    error
}

func (r *pruningHealthRepository) PruneHealth(_ context.Context, cutoff time.Time) error {
	r.calls++
	r.cutoff = cutoff
	return r.err
}

func TestHealthHistoryPruningDoesNotScanOnEverySchedulerTick(t *testing.T) {
	t.Parallel()
	repository := &pruningHealthRepository{err: errors.New("unavailable")}
	service := NewHealthService(nil, nil, repository, nil)
	now := time.Now().UTC()
	reported := 0
	report := func(string, error) { reported++ }
	for range 60 {
		service.pruneHistory(context.Background(), now, report)
		now = now.Add(time.Second)
	}
	if repository.calls != 1 || reported != 1 {
		t.Fatalf("pruning/reports: %d %d", repository.calls, reported)
	}
	repository.err = nil
	service.pruneHistory(context.Background(), now, report)
	if repository.calls != 2 || !repository.cutoff.Equal(now.Add(-24*time.Hour)) {
		t.Fatal("retention cutoff or retry was lost")
	}
}
