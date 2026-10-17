package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"envarium.dev/envarium/internal/diagnostics/domain"
)

type scheduledCollectorFake struct {
	bundle          domain.Bundle
	full, scheduled int
	err             error
}

func (c *scheduledCollectorFake) Collect(context.Context, string) (domain.Bundle, error) {
	c.full++
	return c.bundle, c.err
}
func (c *scheduledCollectorFake) CollectScheduled(context.Context, string) (domain.Bundle, error) {
	c.scheduled++
	result := c.bundle
	result.Snapshot.Git = domain.GitSnapshot{}
	result.Snapshot.Cleanup = domain.CleanupSnapshot{}
	result.Warnings = []string{"Git and cleanup omitted"}
	return result, c.err
}

func TestScheduledDiagnosisPreservesEveryNotificationAndRecipeTrigger(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	bundle := domain.Bundle{ProjectID: "project_fixture", ProjectAgeDay: 60, ProjectState: "degraded", CollectedAt: now,
		Actions:  []domain.Action{{ID: "check", Name: "Check", Type: "health.check", Risk: "read_only"}},
		Evidence: []domain.Evidence{{ID: "runtime", Kind: "runtime", ObservedAt: now}, {ID: "operations", Kind: "operations", ObservedAt: now}, {ID: "port:conflict", Kind: "port", ObservedAt: now}, {ID: "health:probe", Kind: "health", ObservedAt: now}, {ID: "resources", Kind: "resources", ObservedAt: now}, {ID: "log", Kind: "log", ObservedAt: now}, {ID: "git", Kind: "git", ObservedAt: now}, {ID: "cleanup", Kind: "cleanup", ObservedAt: now}},
		Snapshot: domain.ProjectSnapshot{
			Runtime: domain.RuntimeSnapshot{Driver: "compose", EngineConnected: false}, FailedRuns: 3,
			Health:        domain.HealthSnapshot{Failures: []domain.HealthCheck{{ID: "probe", ServiceID: "api", Required: true, Status: "unhealthy", Severity: "error", Message: "failed"}}},
			PortConflicts: []domain.PortConflict{{ID: "conflict", Summary: "port conflict"}},
			Resources:     domain.ResourceSnapshot{Warnings: []string{"CPU pressure"}},
			RecentLogs:    []domain.LogLine{{EvidenceID: "log", Message: "address already in use; connection refused; out of memory; permission denied; missing environment"}},
			Git:           domain.GitSnapshot{Repository: true, Conflicted: 1, Operation: "merge"},
			Cleanup:       domain.CleanupSnapshot{Candidates: 1, EstimatedBytes: 1024},
		}}
	collector := &scheduledCollectorFake{bundle: bundle}
	repo := newRepositoryFake()
	service, _ := NewService(collector, nil, repo)
	full, err := service.Diagnose(t.Context(), bundle.ProjectID, "")
	if err != nil {
		t.Fatal(err)
	}
	scheduled, err := service.DiagnoseScheduled(t.Context(), bundle.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	notifying := func(values []domain.Hypothesis) map[string]domain.Hypothesis {
		result := map[string]domain.Hypothesis{}
		for _, value := range values {
			if value.Notifies {
				result[value.Code] = value
			}
		}
		return result
	}
	if !reflect.DeepEqual(notifying(full.Hypotheses), notifying(scheduled.Hypotheses)) {
		t.Fatal("scheduled notification codes, summaries, evidence or actions changed")
	}
	if len(scheduled.Hypotheses) != len(notifying(full.Hypotheses)) || len(scheduled.Hypotheses) == 0 {
		t.Fatal("scheduled enrichment generated recommendations from missing evidence")
	}
	for _, code := range []string{"REPEATED_CRASH", "PORT_CONFLICT", "PORT_BIND_FAILED", "RESOURCE_PRESSURE", "RESOURCE_EXHAUSTED", "UNHEALTHY_DEPENDENCY", "DEPENDENCY_UNREACHABLE"} {
		if !supportedTrigger(code) || !notifying(scheduled.Hypotheses)[code].Notifies {
			t.Fatalf("supported trigger lost: %s", code)
		}
	}
	if collector.full != 1 || collector.scheduled != 1 || len(repo.notifications) == 0 {
		t.Fatal("collector selection or automatic notifications lost")
	}
	collector.err = context.Canceled
	if _, err := service.DiagnoseScheduled(t.Context(), bundle.ProjectID); !errors.Is(err, context.Canceled) {
		t.Fatal("collector cancellation hidden")
	}
}

func TestScheduledAutomationStillDispatchesDueRecipeAndHonorsCooldown(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	collector := &scheduledCollectorFake{bundle: domain.Bundle{ProjectID: "project_fixture", ProjectState: "running", CollectedAt: now, Snapshot: domain.ProjectSnapshot{FailedRuns: 3}}}
	diagnostics, _ := NewService(collector, nil, newRepositoryFake())
	repo := newRecipeRepositoryFake()
	actions := &automationSourceFake{actions: []domain.Action{{ID: "check", Name: "Check", Type: "health.check", Risk: "read_only"}}}
	submitter := &submitterFake{}
	service, _ := NewAutomationService(repo, actions, diagnostics, actions, submitter)
	service.now = func() time.Time { return now }
	if ids, err := service.evaluate(t.Context(), "project_fixture", true); err != nil || len(ids) != 0 || collector.scheduled != 1 {
		t.Fatal("empty recipe set lost notification diagnosis")
	}
	recipe, err := service.Save(t.Context(), "project_fixture", "Check crashes", "REPEATED_CRASH", "check", 3600, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetEnabled(t.Context(), recipe.ID, true); err != nil {
		t.Fatal(err)
	}
	if ids, err := service.evaluate(t.Context(), "project_fixture", true); err != nil || len(ids) != 1 || submitter.calls != 1 {
		t.Fatal("due recipe did not dispatch")
	}
	if ids, err := service.evaluate(t.Context(), "project_fixture", true); err != nil || len(ids) != 0 || submitter.calls != 1 {
		t.Fatal("cooldown or daily limit bypassed")
	}
	if collector.full != 0 {
		t.Fatal("scheduled evaluator invoked full collection")
	}
}
