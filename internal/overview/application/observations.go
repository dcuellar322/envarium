// Package application coordinates bounded project observation reads.
package application

import (
	"context"
	"errors"
	"sync"
	"time"

	actions "envarium.dev/envarium/internal/actions/domain"
	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
	sourcecontrol "envarium.dev/envarium/internal/sourcecontrol/domain"
)

// ErrInvalidRequest prevents unbounded/ambiguous observation fan-out.
var ErrInvalidRequest = errors.New("invalid project observation request")

// RuntimeReader retains the runtime service's trust and freshness checks.
type RuntimeReader interface {
	Inspect(context.Context, string) (runtime.Observation, error)
}

// GitReader retains source control's current trust and cancellation checks.
type GitReader interface {
	GetStatus(context.Context, string) (sourcecontrol.State, error)
}

// ActionReader resolves current trusted action definitions.
type ActionReader interface {
	List(context.Context, string) (actions.ProjectActions, error)
}

// ProjectObservation contains one requested fact or an explicit warning.
type ProjectObservation struct {
	ProjectID string                       `json:"projectId"`
	Runtime   *runtime.Observation         `json:"runtime,omitempty"`
	Git       *sourcecontrol.State         `json:"git,omitempty"`
	Actions   *actions.ProjectActions      `json:"actions,omitempty"`
	Health    *observability.ProjectHealth `json:"health,omitempty"`
	Warning   string                       `json:"warning,omitempty"`
}

// Batch never combines unrelated kinds into one refresh contract.
type Batch struct {
	Kind       string               `json:"kind"`
	ObservedAt time.Time            `json:"observedAt"`
	Projects   []ProjectObservation `json:"projects"`
}

// Service groups transport reads over existing application boundaries.
type Service struct {
	runtime RuntimeReader
	git     GitReader
	actions ActionReader
	health  HealthReader
}

// NewService supplies explicit observation boundaries.
func NewService(runtime RuntimeReader, git GitReader, actions ActionReader, health HealthReader) *Service {
	return &Service{runtime: runtime, git: git, actions: actions, health: health}
}

// Read returns ordered, fresh observations with at most four concurrent reads.
func (s *Service) Read(ctx context.Context, kind string, ids []string) (Batch, error) {
	if kind != "runtime" && kind != "git-status" && kind != "actions" && kind != "status" {
		return Batch{}, ErrInvalidRequest
	}
	if len(ids) == 0 || len(ids) > 64 {
		return Batch{}, ErrInvalidRequest
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if len(id) == 0 || len(id) > 256 {
			return Batch{}, ErrInvalidRequest
		}
		if _, exists := seen[id]; exists {
			return Batch{}, ErrInvalidRequest
		}
		seen[id] = struct{}{}
	}
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	result := Batch{Kind: kind, ObservedAt: time.Now().UTC(), Projects: make([]ProjectObservation, len(ids))}
	jobs := make(chan int)
	var workers sync.WaitGroup
	for range min(4, len(ids)) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				result.Projects[index] = s.read(ctx, kind, ids[index])
			}
		}()
	}
	for index := range ids {
		select {
		case jobs <- index:
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			return Batch{}, ctx.Err()
		}
	}
	close(jobs)
	workers.Wait()
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	return result, nil
}

func (s *Service) read(ctx context.Context, kind, id string) ProjectObservation {
	item := ProjectObservation{ProjectID: id}
	var err error
	switch kind {
	case "runtime":
		value, readErr := ObserveRuntime(ctx, s.runtime, s.health, id)
		err = readErr
		if err == nil {
			item.Runtime = &value
		}
	case "status":
		status := readStatus(ctx, s.runtime, s.health, id)
		if status.runtimeErr == nil {
			item.Runtime = &status.runtime
		}
		if status.healthErr == nil {
			item.Health = &status.health
		}
		err = errors.Join(status.runtimeErr, status.healthErr)
	case "git-status":
		value, readErr := s.git.GetStatus(ctx, id)
		err = readErr
		if err == nil {
			item.Git = &value
		}
	case "actions":
		value, readErr := s.actions.List(ctx, id)
		err = readErr
		if err == nil {
			item.Actions = &value
		}
	}
	if err != nil {
		item.Warning = "Project observation unavailable; check project trust and source availability."
	}
	return item
}
