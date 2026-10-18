package application

import (
	"context"
	"time"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

// coalescedObservation shares only an in-flight fresh read. Every caller still
// resolves trust before entering; canceling one waiter cannot poison another.
func (s *Service) coalescedObservation(ctx context.Context, projectID, root string, statusOnly bool) (domain.State, error) {
	scope := "full"
	if statusOnly {
		scope = "status"
	}
	return s.coalescedRead(ctx, projectID, root, scope)
}

func (s *Service) coalescedRead(ctx context.Context, projectID, root, scope string) (domain.State, error) {
	if err := ctx.Err(); err != nil {
		return domain.State{}, err
	}
	statusOnly := scope == "status"
	result := s.observations.DoChan(scope+"\x00"+projectID+"\x00"+root, func() (any, error) {
		readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		select {
		case s.gitReads <- struct{}{}:
			defer func() { <-s.gitReads }()
		case <-readCtx.Done():
			return domain.State{}, readCtx.Err()
		}
		if scope == "inventory" {
			observer, ok := s.observer.(interface {
				ObserveInventory(context.Context, string, string) (domain.State, error)
			})
			if !ok {
				return domain.State{}, ErrInventoryUnsupported
			}
			return observer.ObserveInventory(readCtx, projectID, root)
		}
		if observer, ok := s.observer.(StatusObserver); statusOnly && ok {
			return observer.ObserveStatus(readCtx, projectID, root)
		}
		return s.observer.Observe(readCtx, projectID, root)
	})
	select {
	case <-ctx.Done():
		return domain.State{}, ctx.Err()
	case value := <-result:
		if value.Err != nil {
			return domain.State{}, value.Err
		}
		state := value.Val.(domain.State)
		if statusOnly {
			state.Remotes = nil
			state.Worktrees = nil
			state.LastCommit = nil
			state.Details = nil
		}
		state.Remotes = append([]domain.Remote{}, state.Remotes...)
		state.Worktrees = append([]domain.Worktree{}, state.Worktrees...)
		if state.LastCommit != nil {
			commitCopy := *state.LastCommit
			state.LastCommit = &commitCopy
		}
		if state.Details != nil {
			details := *state.Details
			details.Files = append([]domain.FileChange{}, details.Files...)
			details.Branches = append([]domain.Branch{}, details.Branches...)
			state.Details = &details
		}
		return state, nil
	}
}
