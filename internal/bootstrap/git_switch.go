package bootstrap

import (
	"context"
	"encoding/json"
	"errors"

	operations "envarium.dev/envarium/internal/operations/application"
	"envarium.dev/envarium/internal/operations/domain"
	sourcecontrol "envarium.dev/envarium/internal/sourcecontrol/application"
	sourceDomain "envarium.dev/envarium/internal/sourcecontrol/domain"
)

func executeBranchSwitch(ctx context.Context, service *sourcecontrol.BranchService, operation domain.Operation, progress operations.Progress) error {
	var input struct {
		Plan      sourceDomain.BranchSwitchPlan `json:"plan"`
		ActorType string                        `json:"actorType"`
	}
	if json.Unmarshal(operation.Input, &input) != nil || input.Plan.ProjectID != operation.ProjectID || (input.ActorType != "browser" && input.ActorType != "ipc") {
		return errors.New("branch switch requires a reviewed local-user operation")
	}
	if err := progress.Step(ctx, "git.switch", "running", "Revalidating clean stopped checkout and exact source/target commits"); err != nil {
		return err
	}
	if err := service.Execute(ctx, input.Plan); err != nil {
		return err
	}
	return progress.Step(ctx, "git.switch", "succeeded", "Branch switched; runtime remains stopped until a new lifecycle preview is confirmed")
}

type branchAwareExecutor struct {
	branches *sourcecontrol.BranchService
	fallback operations.Executor
}

func (e branchAwareExecutor) Execute(ctx context.Context, operation domain.Operation, progress operations.Progress) error {
	if operation.Kind == "git.switch" {
		return executeBranchSwitch(ctx, e.branches, operation, progress)
	}
	return e.fallback.Execute(ctx, operation, progress)
}
