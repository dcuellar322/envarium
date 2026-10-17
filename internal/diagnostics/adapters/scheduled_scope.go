package adapters

import (
	"context"

	diagnostics "envarium.dev/envarium/internal/diagnostics/application"
	observability "envarium.dev/envarium/internal/observability/domain"
	ports "envarium.dev/envarium/internal/ports/domain"
)

// BeginScheduledCollection captures shared facts once per scheduler cycle.
// Captured source failures remain explicit in every affected project bundle.
func (c *Collector) BeginScheduledCollection(ctx context.Context) (diagnostics.Collector, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	registry, portErr := c.ports.Registry(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	overview, resourceErr := c.resources.Overview(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := *c
	result.ports = scheduledPorts{registry: registry, err: portErr}
	result.resources = scheduledResources{overview: overview, err: resourceErr, source: c.resources}
	return &result, nil
}

type scheduledPorts struct {
	registry ports.Registry
	err      error
}

func (s scheduledPorts) Registry(ctx context.Context) (ports.Registry, error) {
	if err := ctx.Err(); err != nil {
		return ports.Registry{}, err
	}
	return s.registry, s.err
}

type scheduledResources struct {
	overview observability.ResourceOverview
	err      error
	source   resourceSource
}

func (s scheduledResources) Overview(ctx context.Context) (observability.ResourceOverview, error) {
	if err := ctx.Err(); err != nil {
		return observability.ResourceOverview{}, err
	}
	return s.overview, s.err
}
func (s scheduledResources) CleanupPreview(ctx context.Context, id string) (observability.CleanupPreview, error) {
	return s.source.CleanupPreview(ctx, id)
}
