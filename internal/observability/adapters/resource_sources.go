package adapters

import (
	"context"
	"errors"
	"fmt"
	"math"

	catalogApplication "envarium.dev/envarium/internal/catalog/application"
	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	manifestDomain "envarium.dev/envarium/internal/manifest/domain"
	observabilityApplication "envarium.dev/envarium/internal/observability/application"
	observabilityDomain "envarium.dev/envarium/internal/observability/domain"
	runtimeApplication "envarium.dev/envarium/internal/runtime/application"
	runtimeDomain "envarium.dev/envarium/internal/runtime/domain"
)

// ResourceCatalogSource adapts trusted manifests to resource identity and policy.
type ResourceCatalogSource struct{ catalog *catalogApplication.Service }

// NewResourceCatalogSource creates a catalog-backed resource project source.
func NewResourceCatalogSource(catalog *catalogApplication.Service) *ResourceCatalogSource {
	return &ResourceCatalogSource{catalog: catalog}
}

// ListResourceProjects returns only trusted, non-secret resource inputs.
func (s *ResourceCatalogSource) ListResourceProjects(ctx context.Context) ([]observabilityDomain.ProjectDescriptor, error) {
	projects, err := s.catalog.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]observabilityDomain.ProjectDescriptor, 0, len(projects))
	for _, project := range projects {
		if project.TrustState != catalogDomain.TrustTrusted {
			continue
		}
		effective, effectiveErr := s.catalog.ResolveManifest(ctx, project, nil)
		if effectiveErr != nil {
			return nil, effectiveErr
		}
		manifest := effective.Manifest
		budget, budgetErr := resourceBudget(manifest.ResourcePolicy.Warnings)
		if budgetErr != nil {
			return nil, fmt.Errorf("project %s resource policy: %w", project.ID, budgetErr)
		}
		descriptor := observabilityDomain.ProjectDescriptor{
			ID: project.ID, Name: project.DisplayName, Driver: manifest.Runtime.Driver,
			Budget: budget,
		}
		if manifest.Runtime.Compose != nil {
			descriptor.ComposeProjectName = manifest.Runtime.Compose.ProjectName
		}
		result = append(result, descriptor)
	}
	return result, nil
}

func resourceBudget(warnings manifestDomain.ResourceWarnings) (observabilityDomain.ResourceBudget, error) {
	if warnings.MemoryMiB < 0 || uint64(warnings.MemoryMiB) > math.MaxUint64>>20 {
		return observabilityDomain.ResourceBudget{}, errors.New("memory warning is outside the supported range")
	}
	if warnings.StorageGiB < 0 || uint64(warnings.StorageGiB) > uint64(math.MaxInt64>>30) {
		return observabilityDomain.ResourceBudget{}, errors.New("storage warning is outside the supported range")
	}
	return observabilityDomain.ResourceBudget{
		CPUPercent:   float64(warnings.CPUPercent),
		MemoryBytes:  uint64(warnings.MemoryMiB) << 20,
		StorageBytes: int64(warnings.StorageGiB) << 30,
	}, nil
}

type runtimeMetrics interface {
	Inspect(context.Context, string) (runtimeDomain.Observation, error)
	Metrics(context.Context, string, string) ([]runtimeDomain.MetricSample, error)
}

type latestHealth interface {
	LatestResults(context.Context, string) ([]observabilityDomain.HealthResult, error)
}

// ResourceRuntimeSource adapts raw runtime and health observations to resource samples.
type ResourceRuntimeSource struct {
	runtime runtimeMetrics
	health  latestHealth
}

// NewResourceRuntimeSource creates a combined runtime/health sampling adapter.
func NewResourceRuntimeSource(runtime *runtimeApplication.Service, health *observabilityApplication.HealthService) *ResourceRuntimeSource {
	return &ResourceRuntimeSource{runtime: runtime, health: health}
}

// ObserveResources skips driver statistics unless current runtime evidence is active.
func (s *ResourceRuntimeSource) ObserveResources(ctx context.Context, project observabilityDomain.ProjectDescriptor) (observabilityDomain.RuntimeSnapshot, error) {
	observation, samples, err := observeMetricCycle(ctx, s.runtime, project.ID)
	if err != nil && !observation.State.IsActive() {
		return observabilityDomain.RuntimeSnapshot{}, err
	}
	result := observabilityDomain.RuntimeSnapshot{ProjectID: project.ID, Driver: string(observation.Driver), State: string(observation.State), Active: activeRuntime(observation.State), Samples: []observabilityDomain.MetricPoint{}, Warnings: []string{}}
	if err != nil {
		return result, err
	}
	if !result.Active {
		return result, nil
	}
	latencies := map[string]int64{}
	if s.health != nil {
		health, healthErr := s.health.LatestResults(ctx, project.ID)
		if healthErr != nil {
			result.Partial = true
			result.Warnings = append(result.Warnings, "Health latency is unavailable.")
		} else {
			for _, item := range health {
				latencies[item.ServiceID] = max(latencies[item.ServiceID], item.LatencyMS)
			}
		}
	}
	for _, sample := range samples {
		latency, healthAvailable := latencies[sample.ServiceID]
		result.Samples = append(result.Samples, observabilityDomain.MetricPoint{
			Timestamp: sample.Timestamp, ProjectID: sample.ProjectID, ServiceID: sample.ServiceID,
			CPUPercent: sample.CPUPercent, CPUMaxPercent: sample.CPUPercent, CPUAvailable: sample.CPUAvailable,
			MemoryBytes: sample.MemoryBytes, MemoryMaxBytes: sample.MemoryBytes, MemoryLimit: sample.MemoryLimit, MemoryAvailable: sample.MemoryAvailable,
			NetworkRxBytes: sample.NetworkRxBytes, NetworkTxBytes: sample.NetworkTxBytes, NetworkAvailable: sample.NetworkAvailable,
			DiskReadBytes: sample.DiskReadBytes, DiskWriteBytes: sample.DiskWriteBytes, DiskAvailable: sample.DiskAvailable,
			ProcessCount: sample.ProcessCount, RestartCount: sample.RestartCount,
			HealthLatencyMS: latency, HealthAvailable: healthAvailable,
			StorageClassification: observabilityDomain.StorageUnknown, Partial: sample.Partial,
		})
		result.Partial = result.Partial || sample.Partial
	}
	return result, nil
}

func activeRuntime(state runtimeDomain.ProjectState) bool { return state.IsActive() }

func observeMetricCycle(ctx context.Context, source runtimeMetrics, id string) (runtimeDomain.Observation, []runtimeDomain.MetricSample, error) {
	if combined, ok := source.(interface {
		ObserveMetrics(context.Context, string) (runtimeDomain.Observation, []runtimeDomain.MetricSample, error)
	}); ok {
		return combined.ObserveMetrics(ctx, id)
	}
	observation, err := source.Inspect(ctx, id)
	if err != nil || !observation.State.IsActive() {
		return observation, nil, err
	}
	samples, err := source.Metrics(ctx, id, "")
	return observation, samples, err
}
