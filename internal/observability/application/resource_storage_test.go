package application

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"envarium.dev/envarium/internal/observability/domain"
)

func TestOverviewReadsOnlyActiveMetricsAndReconcilesActivity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	metrics := &metricRepositoryFake{latest: map[string][]domain.MetricPoint{
		"one": {{ProjectID: "one", Timestamp: now, CPUPercent: 17, CPUAvailable: true, MemoryBytes: 512, MemoryAvailable: true}},
	}}
	service := newResourceServiceForTest(t, resourceProjectSourceFake{projects: []domain.ProjectDescriptor{{ID: "one"}, {ID: "two"}, {ID: "unknown"}}}, &resourceRuntimeFake{}, metrics, &storageInspectorFake{}, ResourceConfig{})
	service.samplerRunning, service.lastAt = true, now
	service.states["one"] = domain.RuntimeSnapshot{Active: true, State: "running"}
	service.states["two"] = domain.RuntimeSnapshot{State: "stopped", Partial: true}
	overview, err := service.Overview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(metrics.latestIDs, []string{"one"}) || !slices.Equal(metrics.recentIDs, []string{"one"}) {
		t.Fatal("overview read inactive history", metrics.latestIDs, metrics.recentIDs)
	}
	if len(overview.Projects) != 3 || !overview.Projects[0].Active || overview.Projects[0].Metric.CPUPercent != 17 {
		t.Fatal("active values/catalog rows lost", overview.Projects)
	}
	for _, project := range overview.Projects {
		if project.ProjectID == "two" && (!project.Metric.Partial || project.Active || project.Metric.CPUAvailable) {
			t.Fatal("inactive partial state changed", project)
		}
	}
	service.states["one"] = domain.RuntimeSnapshot{State: "stopped"}
	service.states["two"] = domain.RuntimeSnapshot{Active: true, State: "running"}
	if _, err := service.Overview(ctx); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(metrics.latestIDs, []string{"two"}) || !slices.Equal(metrics.recentIDs, []string{"two"}) {
		t.Fatal("activity change did not refresh metric scope", metrics.latestIDs, metrics.recentIDs)
	}
	metrics.readErr = context.Canceled
	if _, err := service.Overview(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("metric cancellation was hidden", err)
	}
}

type blockingStorage struct {
	started chan struct{}
	release chan struct{}
}

func (s *blockingStorage) InspectStorage(ctx context.Context, _ []domain.ProjectDescriptor) (domain.StorageInventory, error) {
	close(s.started)
	select {
	case <-ctx.Done():
		return domain.StorageInventory{}, ctx.Err()
	case <-s.release:
		return domain.StorageInventory{Connected: true, Projects: []domain.ProjectStorage{{ProjectID: "one", Summary: domain.StorageSummary{Bytes: 4096, Classification: domain.StorageExclusive}}}}, nil
	}
}

func TestStorageWorkerCannotBlockMetricCollectionAndStopsWithRun(t *testing.T) {
	t.Parallel()
	storage := &blockingStorage{started: make(chan struct{}), release: make(chan struct{})}
	metrics := &metricRepositoryFake{}
	service, err := NewResourceService(resourceProjectSourceFake{projects: []domain.ProjectDescriptor{{ID: "one", Name: "One", Driver: "process"}}}, &resourceRuntimeFake{observe: func(context.Context, domain.ProjectDescriptor) (domain.RuntimeSnapshot, error) {
		return domain.RuntimeSnapshot{ProjectID: "one", Active: true, State: "running"}, nil
	}}, metrics, storage, ResourceConfig{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); service.Run(ctx, nil) }()
	<-storage.started
	collected := make(chan error, 1)
	go func() { collected <- service.Collect(ctx) }()
	select {
	case err := <-collected:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("metrics waited for the storage scan")
	}
	if len(metrics.writtenPoints()) == 0 {
		t.Fatal("metrics not persisted while storage blocked")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("storage worker survived sampler shutdown")
	}
}

func TestOverviewReadsPersistedStateWithoutReschedulingRunningSampler(t *testing.T) {
	t.Parallel()
	metrics := &metricRepositoryFake{}
	runtime := &resourceRuntimeFake{}
	service := newResourceServiceForTest(t, resourceProjectSourceFake{projects: []domain.ProjectDescriptor{{ID: "one"}}}, runtime, metrics, &storageInspectorFake{}, ResourceConfig{})
	service.samplerRunning = true
	service.lastAt = time.Now().Add(-time.Minute)
	overview, err := service.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls.Load() != 0 {
		t.Fatal("overview started a competing metric cycle")
	}
	found := false
	for _, warning := range overview.Warnings {
		if warning == "Resource samples are stale; observation continues independently." {
			found = true
		}
	}
	if !found {
		t.Fatalf("stale data was not disclosed: %#v", overview.Warnings)
	}
}
