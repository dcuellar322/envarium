package cli

import (
	"context"
	"envarium.dev/envarium/internal/transport/contract/generated"
	"errors"
	"strings"
)

type runtimeScopeClient interface {
	Projects(context.Context) ([]generated.Project, error)
	Environment(context.Context, string) (generated.ProjectEnvironment, error)
}
type runtimeScope struct{ ID, Name string }

// Runtime commands accept exact environment IDs without inventing catalog rows.
// The shared runtime application service still checks current parent trust.
func resolveRuntimeScope(ctx context.Context, client runtimeScopeClient, value string) (runtimeScope, error) {
	projects, err := client.Projects(ctx)
	if err != nil {
		return runtimeScope{}, err
	}
	project, err := selectProject(projects, value)
	if err == nil {
		return runtimeScope{ID: project.Id, Name: project.Slug}, nil
	}
	var selectionError *Error
	if !strings.HasPrefix(value, "env-") || !errors.As(err, &selectionError) || selectionError.ExitCode != exitNotFound {
		return runtimeScope{}, err
	}
	environment, err := client.Environment(ctx, value)
	if err != nil {
		return runtimeScope{}, err
	}
	return runtimeScope{ID: environment.Id, Name: environment.Name}, nil
}
