package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	observability "envarium.dev/envarium/internal/observability/adapters"
	"envarium.dev/envarium/internal/plugins/domain"
	pluginsdk "envarium.dev/envarium/sdk/plugin"
)

type leakingRunner struct{ fail bool }

func (r leakingRunner) Call(_ context.Context, invocation Invocation) ([]domain.LogEntry, error) {
	if r.fail {
		return []domain.LogEntry{{Message: "secret-canary"}}, errors.New("secret-canary")
	}
	switch result := invocation.Result.(type) {
	case *pluginsdk.HealthResult:
		result.Message = "secret-canary"
	case *pluginsdk.InspectResult:
		result.Summary = "secret-canary"
		result.Warnings = []string{"secret-canary"}
		result.Facts = []pluginsdk.Fact{{Value: "secret-canary"}}
	case *pluginsdk.OperateResult:
		result.Summary = "secret-canary"
		result.Output = []byte(`{"nested":["secret-canary"]}`)
	}
	return []domain.LogEntry{{Message: "secret-canary"}}, nil
}

func TestStructuredPluginResultsLogsAndErrorsAreRedacted(t *testing.T) {
	redactor, _ := observability.NewRedactor(nil)
	redactor.AddSecret("secret-canary")
	service := &Service{runner: leakingRunner{}, redactor: redactor}
	health := &pluginsdk.HealthResult{}
	inspect := &pluginsdk.InspectResult{}
	operate := &pluginsdk.OperateResult{}
	for _, result := range []any{health, inspect, operate} {
		logs, err := service.call(context.Background(), Invocation{Result: result})
		if err != nil || strings.Contains(logs[0].Message, "secret-canary") {
			t.Fatalf("logs=%#v err=%v", logs, err)
		}
	}
	if strings.Contains(health.Message+inspect.Summary+inspect.Warnings[0]+inspect.Facts[0].Value+operate.Summary+string(operate.Output), "secret-canary") {
		t.Fatal("structured diagnostic leaked")
	}
	service.runner = leakingRunner{fail: true}
	logs, err := service.call(context.Background(), Invocation{Result: health})
	if err == nil || strings.Contains(err.Error()+logs[0].Message, "secret-canary") {
		t.Fatalf("logs=%#v err=%v", logs, err)
	}
}

func TestPluginSecretIdentifiersAreRejectedBeforeDisplayOrInvocation(t *testing.T) {
	redactor, _ := observability.NewRedactor(nil)
	redactor.AddSecret("secret-canary")
	for _, field := range []string{"factID", "actionID", "risk"} {
		service := &Service{runner: identifierRunner{field: field}, redactor: redactor}
		if _, err := service.call(context.Background(), Invocation{Result: &pluginsdk.InspectResult{}}); err == nil || strings.Contains(err.Error(), "secret-canary") {
			t.Fatalf("field %s err=%v", field, err)
		}
	}
}

type identifierRunner struct{ field string }

func (r identifierRunner) Call(_ context.Context, invocation Invocation) ([]domain.LogEntry, error) {
	result := invocation.Result.(*pluginsdk.InspectResult)
	switch r.field {
	case "factID":
		result.Facts = []pluginsdk.Fact{{ID: "secret-canary"}}
	case "actionID":
		result.Actions = []pluginsdk.Action{{ID: "secret-canary"}}
	case "risk":
		result.Actions = []pluginsdk.Action{{ID: "action", Risk: "secret-canary"}}
	}
	return nil, nil
}
