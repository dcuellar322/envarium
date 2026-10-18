package application

import (
	"context"
	"errors"
	"hash/fnv"
	"sync"
	"time"

	runtime "envarium.dev/envarium/internal/runtime/domain"
)

// Run schedules checks with stable jitter and exits promptly on cancellation.
func (s *HealthService) Run(ctx context.Context, onError func(string, error)) {
	if onError == nil {
		onError = func(string, error) {}
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	snapshots, unsubscribe := subscribeHealthProjects(s.source)
	defer unsubscribe()
	if snapshots != nil {
		ticker.Stop()
	} else {
		s.runDue(ctx, onError)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case snapshot, ok := <-snapshots:
			if !ok {
				return
			}
			s.runDueSnapshot(ctx, snapshot.Projects, snapshot.Failures, snapshot.Err, onError)
		case <-ticker.C:
			s.runDue(ctx, onError)
		}
	}
}

func (s *HealthService) runDue(ctx context.Context, onError func(string, error)) {
	projects, failures, err := healthProjects(ctx, s.source)
	s.runDueSnapshot(ctx, projects, failures, err, onError)
}

func (s *HealthService) runDueSnapshot(ctx context.Context, projects map[string]runtime.ProjectRuntime, failures map[string]error, err error, onError func(string, error)) {
	if err != nil {
		onError("", err)
		return
	}
	now := s.now().UTC()
	semaphore := make(chan struct{}, 4)
	var wait sync.WaitGroup
	wanted := make(map[string]struct{}, len(projects)+len(failures))
	for _, id := range healthProjectIDs(projects, failures) {
		project, resolveErr := projects[id], failures[id]
		if resolveErr != nil {
			wanted[id] = struct{}{}
			onError(id, resolveErr)
			continue
		}
		if !hasHealthChecks(project) {
			continue
		}
		wanted[id] = struct{}{}
		s.mu.Lock()
		schedule, scheduled := s.nextDue[id]
		if !scheduled || schedule.manifestHash != project.ManifestHash {
			schedule = scheduledHealth{due: now.Add(healthInitialDelay(project)), manifestHash: project.ManifestHash}
			s.nextDue[id] = schedule
		}
		if schedule.due.After(now) {
			s.mu.Unlock()
			continue
		}
		s.nextDue[id] = scheduledHealth{due: now.Add(jitteredInterval(id, healthInterval(project))), manifestHash: project.ManifestHash}
		s.mu.Unlock()
		wait.Add(1)
		go s.evaluateScheduled(ctx, id, semaphore, &wait, onError)
	}
	wait.Wait()
	s.removeUnscheduled(wanted)
	s.pruneHistory(ctx, now, onError)
}

func (s *HealthService) evaluateScheduled(ctx context.Context, projectID string, semaphore chan struct{}, wait *sync.WaitGroup, onError func(string, error)) {
	defer wait.Done()
	select {
	case semaphore <- struct{}{}:
		defer func() { <-semaphore }()
	case <-ctx.Done():
		return
	}
	if _, err := s.EvaluateProject(ctx, projectID); err != nil && !errors.Is(err, context.Canceled) {
		onError(projectID, err)
	}
}

func (s *HealthService) removeUnscheduled(wanted map[string]struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.nextDue {
		if _, exists := wanted[id]; !exists {
			delete(s.nextDue, id)
		}
	}
}

func jitteredInterval(key string, interval time.Duration) time.Duration {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	percent := int(hash.Sum32()%21) - 10
	return interval + time.Duration(percent)*interval/100
}

func healthInterval(project runtime.ProjectRuntime) time.Duration {
	interval := defaultHealthInterval
	for _, service := range project.Services {
		for _, check := range service.HealthChecks {
			configured := time.Duration(check.IntervalSeconds) * time.Second
			if configured > 0 && configured < interval {
				interval = configured
			}
		}
	}
	return interval
}

func healthInitialDelay(project runtime.ProjectRuntime) time.Duration {
	var delay time.Duration
	for _, service := range project.Services {
		for _, check := range service.HealthChecks {
			delay = max(delay, time.Duration(check.InitialDelaySeconds)*time.Second)
		}
	}
	return delay
}

// History retention is housekeeping, independent of the one-second check scheduler.
func (s *HealthService) pruneHistory(ctx context.Context, now time.Time, onError func(string, error)) {
	s.mu.Lock()
	if now.Before(s.nextPrune) {
		s.mu.Unlock()
		return
	}
	s.nextPrune = now.Add(time.Minute)
	s.mu.Unlock()
	if err := s.repository.PruneHealth(ctx, now.Add(-24*time.Hour)); err != nil && ctx.Err() == nil {
		onError("", err)
	}
}

// The scheduler persists only declared probe results. With no declarations,
// EvaluateProject performs reads whose aggregate return value is discarded.
// Direct health queries still inspect Docker/process state on demand.
func hasHealthChecks(project runtime.ProjectRuntime) bool {
	for _, service := range project.Services {
		if len(service.HealthChecks) > 0 {
			return true
		}
	}
	return false
}
