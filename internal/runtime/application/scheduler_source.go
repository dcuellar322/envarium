package application

import (
	"context"
	"sync"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

// SchedulerSource shares fresh periodic inputs only among background workers.
// Embedded direct reads always retain the underlying source's fresh checks.
type SchedulerSource struct {
	ProjectSource
	mu          sync.Mutex
	subscribers map[chan domain.ProjectSnapshot]struct{}
	wake        chan struct{}
	closed      bool
}

// NewSchedulerSource creates an explicitly owned background observation feed.
func NewSchedulerSource(source ProjectSource) *SchedulerSource {
	return &SchedulerSource{ProjectSource: source, subscribers: make(map[chan domain.ProjectSnapshot]struct{}), wake: make(chan struct{}, 1)}
}

// ResolveRuntimeProjects preserves independent fresh reads for explicit callers.
func (s *SchedulerSource) ResolveRuntimeProjects(ctx context.Context) (map[string]domain.ProjectRuntime, map[string]error, error) {
	return snapshotProjects(ctx, s.ProjectSource)
}

// SubscribeRuntimeProjects delivers latest cycles with bounded backpressure.
// The consumer must release its subscription when its worker exits.
func (s *SchedulerSource) SubscribeRuntimeProjects() (<-chan domain.ProjectSnapshot, func()) {
	ch := make(chan domain.ProjectSnapshot, 1)
	s.mu.Lock()
	if s.closed {
		close(ch)
	} else {
		s.subscribers[ch] = struct{}{}
	}
	s.mu.Unlock()
	select {
	case s.wake <- struct{}{}:
	default:
	}
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if _, ok := s.subscribers[ch]; ok {
				delete(s.subscribers, ch)
				close(ch)
			}
		})
	}
}

// Run resolves current trust and manifest bytes once each second. Background
// workers may share that exact cycle, but cannot mutate each other's inputs.
func (s *SchedulerSource) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.closed = true
		for ch := range s.subscribers {
			close(ch)
			delete(s.subscribers, ch)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.wake:
		}
		if ctx.Err() != nil {
			return
		}
		s.mu.Lock()
		needed := len(s.subscribers) > 0
		s.mu.Unlock()
		if !needed {
			continue
		}
		projects, failures, err := s.ResolveRuntimeProjects(ctx)
		snapshot := domain.ProjectSnapshot{Projects: projects, Failures: failures, Err: err}
		s.mu.Lock()
		for ch := range s.subscribers {
			select {
			case <-ch:
			default:
			}
			ch <- cloneProjectSnapshot(snapshot)
		}
		s.mu.Unlock()
	}
}

func cloneProjectSnapshot(snapshot domain.ProjectSnapshot) domain.ProjectSnapshot {
	result := domain.ProjectSnapshot{Projects: make(map[string]domain.ProjectRuntime, len(snapshot.Projects)), Failures: make(map[string]error, len(snapshot.Failures)), Err: snapshot.Err}
	for id, project := range snapshot.Projects {
		result.Projects[id] = cloneRuntimeProjection(project)
	}
	for id, err := range snapshot.Failures {
		result.Failures[id] = err
	}
	return result
}

// SubscribeRuntimeProjects exposes the configured background feed to collectors.
// Alternate sources retain the original independent scheduler behavior.
func (s *Service) SubscribeRuntimeProjects() (<-chan domain.ProjectSnapshot, func()) {
	if feed, ok := s.projects.(interface {
		SubscribeRuntimeProjects() (<-chan domain.ProjectSnapshot, func())
	}); ok {
		return feed.SubscribeRuntimeProjects()
	}
	return nil, func() {}
}
