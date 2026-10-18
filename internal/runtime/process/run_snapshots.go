package process

import (
	"context"
	"sort"
	"sync"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

type runSnapshot struct {
	runs []domain.RunRecord
	err  error
}
type runSnapshotHub struct {
	ctx         context.Context
	store       RunStore
	mu          sync.Mutex
	subscribers map[chan runSnapshot]string
	wake        chan struct{}
	cancel      context.CancelFunc
	generation  uint64
}

func (d *Driver) subscribeRunSnapshots(id string) (<-chan runSnapshot, func()) {
	d.mu.Lock()
	if d.runFeed == nil {
		ctx := d.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		d.runFeed = &runSnapshotHub{ctx: ctx, store: d.store, subscribers: map[chan runSnapshot]string{}, wake: make(chan struct{}, 1)}
	}
	hub := d.runFeed
	d.mu.Unlock()
	return hub.subscribe(id)
}
func (h *runSnapshotHub) subscribe(id string) (<-chan runSnapshot, func()) {
	ch := make(chan runSnapshot, 1)
	h.mu.Lock()
	if h.ctx.Err() != nil {
		close(ch)
		h.mu.Unlock()
		return ch, func() {}
	}
	h.subscribers[ch] = id
	if h.cancel == nil {
		ctx, cancel := context.WithCancel(h.ctx)
		h.cancel = cancel
		h.generation++
		go h.run(ctx, h.generation)
	}
	h.mu.Unlock()
	select {
	case h.wake <- struct{}{}:
	default:
	}
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if _, exists := h.subscribers[ch]; exists {
				delete(h.subscribers, ch)
				close(ch)
			}
			if len(h.subscribers) == 0 && h.cancel != nil {
				h.cancel()
				h.cancel = nil
			}
		})
	}
}
func (h *runSnapshotHub) run(ctx context.Context, generation uint64) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	defer func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.generation == generation && h.cancel != nil {
			h.cancel()
			h.cancel = nil
			if h.ctx.Err() != nil {
				for ch := range h.subscribers {
					close(ch)
					delete(h.subscribers, ch)
				}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-h.wake:
		}
		if ctx.Err() != nil {
			return
		}
		h.mu.Lock()
		wanted := map[string]struct{}{}
		for _, id := range h.subscribers {
			wanted[id] = struct{}{}
		}
		h.mu.Unlock()
		ids := make([]string, 0, len(wanted))
		for id := range wanted {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		runs, err := runSnapshots(ctx, h.store, ids)
		h.mu.Lock()
		if ctx.Err() != nil {
			h.mu.Unlock()
			return
		}
		for ch, id := range h.subscribers {
			if _, observed := wanted[id]; !observed {
				continue
			}
			select {
			case <-ch:
			default:
			}
			ch <- runSnapshot{runs: cloneRuns(runs[id]), err: err}
		}
		h.mu.Unlock()
	}
}
func runSnapshots(ctx context.Context, store RunStore, ids []string) (map[string][]domain.RunRecord, error) {
	result := map[string][]domain.RunRecord{}
	if batch, ok := store.(interface {
		ListProjectsRuns(context.Context, []string) (map[string][]domain.RunRecord, error)
	}); ok {
		for len(ids) > 0 {
			count := min(64, len(ids))
			runs, err := batch.ListProjectsRuns(ctx, ids[:count])
			if err != nil {
				return nil, err
			}
			for id, records := range runs {
				result[id] = records
			}
			ids = ids[count:]
		}
		return result, ctx.Err()
	}
	for _, id := range ids {
		runs, err := store.ListProjectRuns(ctx, id)
		if err != nil {
			return nil, err
		}
		result[id] = runs
	}
	return result, ctx.Err()
}
func cloneRuns(runs []domain.RunRecord) []domain.RunRecord {
	result := append([]domain.RunRecord{}, runs...)
	for i := range result {
		r := &result[i]
		r.Processes = append([]domain.ProcessIdentity{}, r.Processes...)
		if r.EndedAt != nil {
			at := *r.EndedAt
			r.EndedAt = &at
		}
		if r.ExitCode != nil {
			code := *r.ExitCode
			r.ExitCode = &code
		}
	}
	return result
}
