package compose

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"sync"
)

// ObserveMetrics shares only this cycle's current Engine evidence. Statistics
// run before status assembly so restart metadata is never borrowed from an older
// status cache. No results survive this call or alter lifecycle observations.
func (d *Driver) ObserveMetrics(ctx context.Context, project domain.ProjectRuntime) (domain.Observation, []domain.MetricSample, error) {
	if err := ctx.Err(); err != nil {
		return domain.Observation{}, nil, err
	}
	config, err := d.observationConfig(ctx, project)
	if err != nil {
		return disconnectedObservation(project, normalizedConfig{}, err), nil, nil
	}
	return d.observeMetricCycle(ctx, project, config)
}

func (d *Driver) observeMetricCycle(ctx context.Context, project domain.ProjectRuntime, config normalizedConfig) (domain.Observation, []domain.MetricSample, error) {
	if err := ctx.Err(); err != nil {
		return domain.Observation{}, nil, err
	}
	engine, ping, err := d.engine.Connect(ctx, config.Connection)
	if err != nil {
		return disconnectedObservation(project, config, err), nil, nil
	}
	defer func() { _ = engine.Close() }()
	observation, items, ownership, err := d.inspectionInputs(ctx, project, config, engine, ping)
	if err != nil {
		return disconnectedObservation(project, config, err), nil, nil
	}
	scope := &metricCycleEngine{engineClient: engine, inspects: make(map[string]client.ContainerInspectResult, len(items))}
	active := []container.Summary{}
	for _, item := range items {
		if item.State == container.StateRunning || item.State == container.StatePaused || item.State == container.StateRestarting {
			active = append(active, item)
		}
	}
	sink := &metricSampleBuffer{}
	statsErr := streamMetricItems(ctx, domain.MetricRequest{Project: project}, scope, active, sink)
	if ctx.Err() != nil {
		return observation, sink.samples, ctx.Err()
	}
	observation, err = d.observeContainers(ctx, scope, project, config, observation, items, ownership)
	if err != nil {
		return domain.Observation{}, sink.samples, err
	}
	return observation, sink.samples, statsErr
}

type metricCycleEngine struct {
	engineClient
	mu       sync.Mutex
	inspects map[string]client.ContainerInspectResult
}

func (e *metricCycleEngine) ContainerInspect(ctx context.Context, id string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	e.mu.Lock()
	value, ok := e.inspects[id]
	e.mu.Unlock()
	if ok {
		return value, nil
	}
	value, err := e.engineClient.ContainerInspect(ctx, id, options)
	if err == nil {
		e.mu.Lock()
		e.inspects[id] = value
		e.mu.Unlock()
	}
	return value, err
}
