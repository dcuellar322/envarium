package application

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"errors"
)

// WithIdleNativeRuntime serializes native allocation changes with lifecycle
// execution at the checkout root. Only fresh owned process evidence blocks an
// allocation; an unrelated external listener is never adopted as owned work.
func (s *Service) WithIdleNativeRuntime(ctx context.Context, id string, change func() error) error {
	project, _, err := s.resolve(ctx, id)
	if err != nil {
		return err
	}
	release, err := s.checkouts.acquire(ctx, project.Root)
	if err != nil {
		return err
	}
	defer release()
	current, driver, err := s.resolve(ctx, id)
	if err != nil {
		return err
	}
	if current.Root != project.Root {
		return ErrPlanStale
	}
	if current.Kind == domain.KindProcess {
		observation, err := driver.Inspect(ctx, current)
		if err != nil {
			return err
		}
		for _, service := range observation.Services {
			if service.Process != nil && service.Process.RunID != "" && service.Process.FinishedAt == nil {
				return errors.New("stop this native worktree before changing its allocated ports")
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return change()
}
