package adapters

import (
	"context"
	"fmt"
	"strings"

	"envarium.dev/envarium/internal/environments/application"
	"envarium.dev/envarium/internal/environments/domain"
)

// GitRoots keeps checkout inspection behind current parent-project trust.
type GitRoots struct {
	projects     *CatalogSource
	environments *application.Service
}

// NewGitRoots resolves base and registered checkout roots behind catalog trust.
func NewGitRoots(projects *CatalogSource, environments *application.Service) *GitRoots {
	return &GitRoots{projects: projects, environments: environments}
}

// Root returns the currently available trusted checkout path.
func (s *GitRoots) Root(ctx context.Context, id string) (string, error) {
	if !strings.HasPrefix(id, "env-") {
		project, err := s.projects.Project(ctx, id)
		return project.PrimaryLocation, err
	}
	environment, err := s.environments.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if _, err := s.projects.Project(ctx, environment.ProjectID); err != nil {
		return "", err
	}
	if environment.Availability != domain.AvailabilityAvailable || environment.Bare {
		return "", fmt.Errorf("selected worktree is unavailable or bare; reconcile its registration")
	}
	return environment.Path, nil
}
