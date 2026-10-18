package process

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

type batchRunFixture struct {
	RunStore
	mu      sync.Mutex
	calls   int
	head    string
	err     error
	ids     []string
	entered chan struct{}
	blocked bool
}

func (f *batchRunFixture) ListProjectsRuns(ctx context.Context, ids []string) (map[string][]domain.RunRecord, error) {
	f.mu.Lock()
	f.calls++
	f.ids = append([]string{}, ids...)
	head, err, blocked := f.head, f.err, f.blocked
	f.mu.Unlock()
	if f.entered != nil {
		select {
		case f.entered <- struct{}{}:
		default:
		}
	}
	if blocked {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	result := map[string][]domain.RunRecord{}
	for _, id := range ids {
		result[id] = []domain.RunRecord{{ID: head, ProjectID: id, Processes: []domain.ProcessIdentity{{PID: 123}}}}
	}
	return result, err
}
func readRunFeed(t *testing.T, ch <-chan runSnapshot) runSnapshot {
	t.Helper()
	select {
	case value, ok := <-ch:
		if !ok {
			t.Fatal("feed closed")
		}
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("no run snapshot")
		return runSnapshot{}
	}
}
func TestRunSnapshotFeedSharesFreshReadsAndBoundsSlowSubscribers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := &batchRunFixture{head: "first"}
	a, b := make(chan runSnapshot, 1), make(chan runSnapshot, 1)
	hub := &runSnapshotHub{ctx: ctx, store: source, subscribers: map[chan runSnapshot]string{a: "a", b: "b"}, wake: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() { hub.run(ctx, 1); close(done) }()
	hub.wake <- struct{}{}
	first, second := readRunFeed(t, a), readRunFeed(t, b)
	source.mu.Lock()
	calls := source.calls
	source.mu.Unlock()
	if calls != 1 || first.runs[0].ID != "first" || second.runs[0].ID != "first" {
		t.Fatal("cycle not shared")
	}
	first.runs[0].Processes[0].PID = 999
	if second.runs[0].Processes[0].PID != 123 {
		t.Fatal("subscriber values alias")
	}
	source.mu.Lock()
	source.head = "updated"
	source.mu.Unlock()
	hub.wake <- struct{}{}
	if readRunFeed(t, a).runs[0].ID != "updated" {
		t.Fatal("stale run state")
	}
	source.mu.Lock()
	source.head = "latest"
	source.err = errors.New("read failed")
	source.mu.Unlock()
	hub.wake <- struct{}{}
	if readRunFeed(t, a).err == nil {
		t.Fatal("failure hidden")
	}
	latest := readRunFeed(t, b)
	if latest.err == nil || len(latest.runs) != 0 {
		t.Fatal("slow subscriber retained old cycles")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("feed did not cancel")
	}
}
func TestRunFeedLastUnsubscribeCancelsReadAndCanRestart(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := &batchRunFixture{blocked: true, entered: make(chan struct{}, 2)}
	hub := &runSnapshotHub{ctx: ctx, store: source, subscribers: map[chan runSnapshot]string{}, wake: make(chan struct{}, 1)}
	_, release := hub.subscribe("p")
	select {
	case <-source.entered:
	case <-time.After(time.Second):
		t.Fatal("read not started")
	}
	release()
	release()
	source.mu.Lock()
	source.blocked = false
	source.head = "fresh"
	source.mu.Unlock()
	ch, again := hub.subscribe("p")
	defer again()
	if readRunFeed(t, ch).runs[0].ID != "fresh" {
		t.Fatal("resubscription stale")
	}
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("canceled driver feed stayed open")
		}
	case <-time.After(time.Second):
		t.Fatal("driver feed did not close")
	}
	closed, unsub := hub.subscribe("p")
	defer unsub()
	if _, ok := <-closed; ok {
		t.Fatal("subscription to canceled driver remained live")
	}
}
