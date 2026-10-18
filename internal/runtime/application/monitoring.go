package application

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
)

// InspectProject resolves current trusted inputs once for a health observation.
// This is not a retained configuration snapshot or a lifecycle execution plan.
func (s *Service) InspectProject(ctx context.Context, id string) (domain.ProjectRuntime, domain.Observation, error) {
	project, driver, err := s.resolve(ctx, id)
	if err != nil {
		return domain.ProjectRuntime{}, domain.Observation{}, err
	}
	observation, err := s.observations.load(ctx, id, func() (domain.Observation, error) { return driver.Inspect(ctx, project) })
	return project, observation, err
}

// ObserveMetrics shares only this call's current trusted runtime definition
// between status observation and driver statistics. Inactive projects do no stats.
func (s *Service) ObserveMetrics(ctx context.Context, id string) (domain.Observation, []domain.MetricSample, error) {
	project, driver, err := s.resolve(ctx, id)
	if err != nil {
		return domain.Observation{}, nil, err
	}
	observation, err := s.observations.load(ctx, id, func() (domain.Observation, error) { return driver.Inspect(ctx, project) })
	if err != nil || !observation.State.IsActive() {
		return observation, []domain.MetricSample{}, err
	}
	sink := &metricCollector{samples: []domain.MetricSample{}}
	err = driver.StreamMetrics(ctx, domain.MetricRequest{Project: project}, sink)
	return observation, sink.samples, err
}
