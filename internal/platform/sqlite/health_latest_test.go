package sqlite

import (
	"context"
	health "envarium.dev/envarium/internal/observability/domain"
	"testing"
	"time"
)

func TestLatestHealthPreservesScopeAndTimestampIDTieBreaking(t *testing.T) {
	t.Parallel()
	database := newResourceMetricDatabase(t)
	repo := NewHealthRepository(database)
	at := time.Now().UTC()
	appendResult := func(service, check string, status health.HealthStatus, when time.Time) {
		t.Helper()
		err := repo.AppendHealth(context.Background(), health.HealthResult{ProjectID: "project-1", ServiceID: service, CheckID: check, Type: "http", Status: status, Severity: "critical", ObservedAt: when})
		if err != nil {
			t.Fatal(err)
		}
	}
	appendResult("api", "ready", health.StatusHealthy, at)
	appendResult("api", "ready", health.StatusUnhealthy, at)
	appendResult("db", "ready", health.StatusHealthy, at.Add(-time.Minute))
	for i := range 8 {
		appendResult("api", "ready", health.StatusHealthy, at.Add(-time.Duration(i+1)*time.Hour))
	}
	latest, err := repo.LatestHealth(context.Background(), "project-1")
	if err != nil || len(latest) != 2 || latest[0].ServiceID != "api" || latest[0].Status != health.StatusUnhealthy || !latest[0].ObservedAt.Equal(at) || latest[1].ServiceID != "db" || latest[1].Status != health.StatusHealthy {
		t.Fatalf("latest=%#v err=%v", latest, err)
	}
	missing, err := repo.LatestHealth(context.Background(), "missing")
	if err != nil || len(missing) != 0 {
		t.Fatalf("cross-project results=%#v err=%v", missing, err)
	}
}
