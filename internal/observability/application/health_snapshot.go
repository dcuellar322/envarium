package application

import (
	"context"
	"sort"

	runtime "envarium.dev/envarium/internal/runtime/domain"
)

type healthSnapshotSource interface {
	ResolveRuntimeProjects(context.Context) (map[string]runtime.ProjectRuntime, map[string]error, error)
}

func healthProjects(ctx context.Context, source RuntimeSource) (map[string]runtime.ProjectRuntime, map[string]error, error) {
	if source, ok := source.(healthSnapshotSource); ok {
		return source.ResolveRuntimeProjects(ctx)
	}
	ids, err := source.ListRuntimeProjectIDs(ctx)
	if err != nil {
		return nil, nil, err
	}
	projects := make(map[string]runtime.ProjectRuntime, len(ids))
	failures := make(map[string]error)
	for _, id := range ids {
		project, err := source.ResolveRuntime(ctx, id)
		if err != nil {
			failures[id] = err
		} else {
			projects[id] = project
		}
	}
	return projects, failures, nil
}

func healthProjectIDs(projects map[string]runtime.ProjectRuntime, failures map[string]error) []string {
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
