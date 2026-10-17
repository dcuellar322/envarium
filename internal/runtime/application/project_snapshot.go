package application

import (
	"context"
	"sort"

	catalogApp "envarium.dev/envarium/internal/catalog/application"
	catalog "envarium.dev/envarium/internal/catalog/domain"
	"envarium.dev/envarium/internal/runtime/domain"
)

// ProjectSnapshotSource batches read-only scheduler resolution. Direct lifecycle
// requests still call ResolveRuntime and read their current catalog entry.
type ProjectSnapshotSource interface {
	ResolveRuntimeProjects(context.Context) (map[string]domain.ProjectRuntime, map[string]error, error)
}

// ResolveRuntimeProjects reads the current catalog once, preserving individual
// resolution failures and freshly reading local manifest/worktree inputs.
func (s *CatalogSource) ResolveRuntimeProjects(ctx context.Context) (map[string]domain.ProjectRuntime, map[string]error, error) {
	projects, err := s.catalog.ListProjects(ctx)
	if err != nil {
		return nil, nil, err
	}
	parents := make(map[string]catalog.Project, len(projects))
	ids := []string{}
	for _, project := range projects {
		parents[project.ID] = project
		if project.TrustState == catalog.TrustTrusted {
			ids = append(ids, project.ID)
		}
	}
	if s.environments != nil {
		environments, err := s.environments.ListRuntimeEnvironmentIDs(ctx)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, environments...)
	}
	resolved := make(map[string]domain.ProjectRuntime, len(ids))
	failures := make(map[string]error)
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		parentID, environment, err := s.resolveEnvironment(ctx, id)
		if err != nil {
			failures[id] = err
			continue
		}
		project, exists := parents[parentID]
		if !exists {
			failures[id] = catalogApp.ErrNotFound
			continue
		}
		runtime, err := s.runtimeProject(ctx, project, id, environment)
		if err != nil {
			failures[id] = err
			continue
		}
		resolved[id] = runtime
	}
	return resolved, failures, nil
}

// snapshotProjects preserves the legacy source boundary for alternate adapters.
func snapshotProjects(ctx context.Context, source ProjectSource) (map[string]domain.ProjectRuntime, map[string]error, error) {
	if source, ok := source.(ProjectSnapshotSource); ok {
		return source.ResolveRuntimeProjects(ctx)
	}
	ids, err := source.ListRuntimeProjectIDs(ctx)
	if err != nil {
		return nil, nil, err
	}
	resolved := make(map[string]domain.ProjectRuntime, len(ids))
	failures := make(map[string]error)
	for _, id := range ids {
		project, err := source.ResolveRuntime(ctx, id)
		if err != nil {
			failures[id] = err
		} else {
			resolved[id] = project
		}
	}
	return resolved, failures, nil
}

func snapshotIDs(projects map[string]domain.ProjectRuntime, failures map[string]error) []string {
	ids := make([]string, 0, len(projects)+len(failures))
	for id := range projects {
		ids = append(ids, id)
	}
	for id := range failures {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
