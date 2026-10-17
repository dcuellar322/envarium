package adapters

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	diagnostics "envarium.dev/envarium/internal/diagnostics/application"
	observability "envarium.dev/envarium/internal/observability/domain"
	ports "envarium.dev/envarium/internal/ports/domain"
)

type scopedPortFixture struct {
	calls int
	fail  bool
}

func (s *scopedPortFixture) Registry(context.Context) (ports.Registry, error) {
	s.calls++
	if s.fail {
		return ports.Registry{}, errors.New("unavailable")
	}
	return ports.Registry{ObservedAt: time.Now()}, nil
}

type scopedResourceFixture struct {
	calls int
	fail  bool
	diagnosticResourcesFixture
}

func (s *scopedResourceFixture) Overview(context.Context) (observability.ResourceOverview, error) {
	s.calls++
	if s.fail {
		return observability.ResourceOverview{}, errors.New("unavailable")
	}
	return observability.ResourceOverview{ObservedAt: time.Now(), Projects: []observability.ProjectSnapshot{
		{ProjectID: "one", Metric: observability.MetricPoint{CPUPercent: 11}},
		{ProjectID: "two", Metric: observability.MetricPoint{CPUPercent: 22}},
	}}, nil
}
func TestScheduledScopeSharesOnlyGlobalFactsAndPreservesPartialFailures(t *testing.T) {
	t.Parallel()
	for _, fail := range []bool{false, true} {
		ports := &scopedPortFixture{fail: fail}
		resources := &scopedResourceFixture{fail: fail}
		git := &diagnosticGitFixture{}
		collector := NewCollector(diagnosticCatalogFixture{}, diagnosticRuntimeFixture{}, diagnosticHealthFixture{}, diagnosticLogsFixture{}, ports, git, diagnosticActionsFixture{}, resources, diagnosticOperationsFixture{})
		scope, err := collector.BeginScheduledCollection(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		scheduled := scope.(diagnostics.ScheduledCollector)
		for i, id := range []string{"one", "two"} {
			bundle, err := scheduled.CollectScheduled(context.Background(), id)
			if err != nil || bundle.ProjectID != id {
				t.Fatalf("id=%s err=%v", id, err)
			}
			if !fail && bundle.Snapshot.Resources.CPUPercent != float64((i+1)*11) {
				t.Fatal("project resource attribution changed")
			}
			if fail && (!slices.Contains(bundle.Warnings, "Resource observations are unavailable.") || !slices.Contains(bundle.Warnings, "Port registry is unavailable.")) {
				t.Fatal("partial failure hidden")
			}
		}
		if ports.calls != 1 || resources.calls != 1 || git.calls != 0 || resources.cleanupCalls != 0 {
			t.Fatal("duplicated globals or executed interactive enrichment")
		}
		if _, err := collector.CollectScheduled(context.Background(), "one"); err != nil {
			t.Fatal(err)
		}
		if ports.calls != 2 || resources.calls != 2 {
			t.Fatal("scope leaked into independent collection")
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := collector.BeginScheduledCollection(ctx); !errors.Is(err, context.Canceled) {
			t.Fatal("canceled scope succeeded")
		}
	}
}
