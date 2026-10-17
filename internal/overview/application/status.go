package application

import (
	"context"
	"errors"

	observability "envarium.dev/envarium/internal/observability/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
)

// HealthReader retains the health service's live trust and probe observations.
type HealthReader interface {
	Get(context.Context, string) (observability.ProjectHealth, error)
}

type projectStatus struct {
	runtime               runtime.Observation
	health                observability.ProjectHealth
	runtimeErr, healthErr error
}

// ObserveRuntime preserves health-derived runtime states across adapters.
// Unavailable health does not invent a runtime failure or hide runtime evidence.
func ObserveRuntime(ctx context.Context, source RuntimeReader, health HealthReader, id string) (runtime.Observation, error) {
	status := readStatus(ctx, source, health, id)
	return status.runtime, status.runtimeErr
}

func readStatus(ctx context.Context, source RuntimeReader, health HealthReader, id string) projectStatus {
	var result projectStatus
	if err := ctx.Err(); err != nil {
		result.runtimeErr, result.healthErr = err, err
		return result
	}
	result.runtime, result.runtimeErr = source.Inspect(ctx, id)
	if err := ctx.Err(); err != nil {
		result.runtimeErr, result.healthErr = err, err
		return result
	}
	if health != nil {
		result.health, result.healthErr = health.Get(ctx, id)
	} else {
		result.healthErr = errors.New("health observer unavailable")
	}
	if err := ctx.Err(); err != nil {
		result.runtimeErr, result.healthErr = err, err
		return result
	}
	if result.runtimeErr == nil {
		services := make([]runtime.ServiceObservation, len(result.runtime.Services))
		copy(services, result.runtime.Services)
		for i := range services {
			if services[i].Ports == nil {
				services[i].Ports = []runtime.PublishedPort{}
			}
		}
		result.runtime.Services = services
	}
	if result.healthErr == nil && result.health.Results == nil {
		result.health.Results = []observability.HealthResult{}
	}
	if result.runtimeErr != nil || result.healthErr != nil {
		return result
	}
	result.runtime.State = runtimeStateWithHealth(result.runtime.State, result.health)
	return result
}

func runtimeStateWithHealth(state runtime.ProjectState, health observability.ProjectHealth) runtime.ProjectState {
	if health.Status == "unhealthy" && (state == runtime.StateRunning || state == runtime.StateRunningExternal || state == runtime.StatePartiallyRunning) {
		return runtime.StateDegraded
	}
	if health.Status == "unknown" && health.ObserverState == "connected" && len(health.Results) > 0 && (state == runtime.StateRunning || state == runtime.StateRunningExternal) {
		return runtime.StateStarting
	}
	return state
}
