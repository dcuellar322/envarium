package application

import (
	"context"
	"testing"
	"time"

	"envarium.dev/envarium/internal/observability/domain"
)

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
