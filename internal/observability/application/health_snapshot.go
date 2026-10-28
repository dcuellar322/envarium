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
	ids := make([]string, 0, len(projects))
	for id := range projects {
		ids = append(ids, id)
	}
	for id := range failures {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// runtimeProjectFeed is restricted to background configuration observation.
type runtimeProjectFeed interface {
	SubscribeRuntimeProjects() (<-chan runtime.ProjectSnapshot, func())
}

func subscribeHealthProjects(source any) (<-chan runtime.ProjectSnapshot, func()) {
	if feed, ok := source.(runtimeProjectFeed); ok {
		return feed.SubscribeRuntimeProjects()
	}
	return nil, func() {}
}

// A combined runtime observer owns its current trust/configuration check. This
// avoids immediately resolving the same files twice in one health read.
func observeHealthProject(ctx context.Context, source RuntimeSource, observer RuntimeObserver, id string) (runtime.ProjectRuntime, runtime.Observation, error) {
	if combined, ok := observer.(interface {
		InspectProject(context.Context, string) (runtime.ProjectRuntime, runtime.Observation, error)
	}); ok {
		return combined.InspectProject(ctx, id)
	}
	project, err := source.ResolveRuntime(ctx, id)
	if err != nil {
		return runtime.ProjectRuntime{}, runtime.Observation{}, err
	}
	observation, err := observer.Inspect(ctx, id)
	return project, observation, err
}
