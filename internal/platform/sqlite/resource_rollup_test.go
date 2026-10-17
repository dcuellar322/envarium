package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	observability "envarium.dev/envarium/internal/observability/domain"
)

func TestMetricRollupSeparatesProjectsAndRetainsTargetOnlyHistory(t *testing.T) {
	t.Parallel()
	d := newResourceMetricDatabase(t)
	ctx := context.Background()
	base := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	if _, err := d.connection.ExecContext(ctx, `INSERT INTO projects
		(id, slug, display_name, trust_state, primary_location, created_at, updated_at)
		VALUES ('project-2', 'project-2', 'Project Two', 'trusted', '/tmp/project-2', ?, ?)`, formatTime(base), formatTime(base)); err != nil {
		t.Fatal(err)
	}
	points := []observability.MetricPoint{
		{ProjectID: "project-1", ServiceID: "api", Timestamp: base.Add(10 * time.Second), CPUPercent: 10, CPUAvailable: true},
		{ProjectID: "project-2", ServiceID: "api", Timestamp: base.Add(-50 * time.Second), CPUPercent: 30, CPUAvailable: true},
		{ProjectID: "project-1", ServiceID: "retired", Timestamp: base.Add(-time.Hour), ResolutionSeconds: 60, CPUPercent: 50, CPUAvailable: true},
		// A newer watermark belonging to the other project must not suppress this source.
		{ProjectID: "project-1", ServiceID: "api", Timestamp: base, ResolutionSeconds: 60, CPUPercent: 99, CPUAvailable: true},
	}
	if err := d.WriteMetricPoints(ctx, points); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := d.rollupMetrics(ctx, 0, 60); err != nil {
			t.Fatal(err)
		}
		for _, item := range []struct {
			project, service string
			cpu              float64
		}{
			{"project-1", "api", 10}, {"project-2", "api", 30}, {"project-1", "retired", 50},
		} {
			point := oneMetricAtResolution(t, d, item.project, item.service, 60)
			if point.CPUPercent != item.cpu || point.SampleCount != 1 {
				t.Fatalf("%s/%s rollup = %#v", item.project, item.service, point)
			}
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := d.rollupMetrics(canceled, 0, 60); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled rollup = %v", err)
	}
	if point := oneMetricAtResolution(t, d, "project-1", "retired", 60); point.CPUPercent != 50 {
		t.Fatalf("cancellation changed retained history: %#v", point)
	}
}
