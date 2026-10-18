package adapters

import (
	"context"
	"envarium.dev/envarium/internal/environments/application"
	portsApplication "envarium.dev/envarium/internal/ports/application"
	"errors"
)

// PortRuntimeScopes supplies trusted worktree names to live port attribution.
type PortRuntimeScopes struct {
	environments *RuntimeSource
	projects     *CatalogSource
}

// NewPortRuntimeScopes connects the environment and port application contracts.
func NewPortRuntimeScopes(environments *application.Service, projects *CatalogSource) *PortRuntimeScopes {
	return &PortRuntimeScopes{environments: NewRuntimeSource(environments), projects: projects}
}

// PortRuntimeScopes lists available checkouts whose parent remains trusted.
func (s *PortRuntimeScopes) PortRuntimeScopes(ctx context.Context) ([]portsApplication.RuntimeScope, error) {
	environments, err := s.environments.ListRuntimeEnvironments(ctx)
	if err != nil {
		return nil, err
	}
	result := []portsApplication.RuntimeScope{}
	for _, environment := range environments {
		if _, err := s.projects.Project(ctx, environment.ProjectID); errors.Is(err, application.ErrProjectUntrusted) {
			continue
		} else if err != nil {
			return nil, err
		}
		result = append(result, portsApplication.RuntimeScope{ID: environment.ID, Name: environment.Name})
	}
	return result, nil
}
