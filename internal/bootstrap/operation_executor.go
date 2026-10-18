package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	actionsApplication "envarium.dev/envarium/internal/actions/application"
	agentsApplication "envarium.dev/envarium/internal/agents/application"
	operations "envarium.dev/envarium/internal/operations/application"
	"envarium.dev/envarium/internal/operations/domain"
	pluginsApplication "envarium.dev/envarium/internal/plugins/application"
	runtimeApplication "envarium.dev/envarium/internal/runtime/application"
	runtimeDomain "envarium.dev/envarium/internal/runtime/domain"
	workspaceApplication "envarium.dev/envarium/internal/workspace/application"
	workspaceDomain "envarium.dev/envarium/internal/workspace/domain"
)

func executeOperation(
	ctx context.Context,
	runtimeService *runtimeApplication.Service,
	health requiredHealthWaiter,
	actions *actionsApplication.Service,
	ai *agentsApplication.EnhancementService,
	plugins *pluginsApplication.Service,
	workspaces *workspaceApplication.Service,
	recipes workspaceApplication.RecipeRunner,
	environments environmentLifecycle,
	operation domain.Operation,
	progress operations.Progress,
) error {
	switch operation.Kind {
	case "manifest.enhance":
		return executeManifestEnhancement(ctx, ai, operation, progress)
	case "action.run":
		return executeActionOperation(ctx, actions, operation, progress)
	case "plugin.operate":
		return executePluginOperation(ctx, plugins, operation, progress)
	case "workspace.start", "workspace.stop":
		return executeWorkspaceOperation(ctx, workspaces, recipes, operation, progress)
	default:
		return executeRuntimeOperation(ctx, runtimeService, health, environments, operation, progress)
	}
}

func executeManifestEnhancement(ctx context.Context, ai *agentsApplication.EnhancementService, operation domain.Operation, progress operations.Progress) error {
	var input struct {
		ProposalID    string                   `json:"proposalId"`
		Provider      string                   `json:"provider"`
		ConsentSHA256 string                   `json:"consentSha256"`
		Limits        agentsApplication.Limits `json:"limits"`
	}
	if err := json.Unmarshal(operation.Input, &input); err != nil || input.ProposalID == "" || input.Provider == "" {
		return errors.New("decode manifest enhancement operation")
	}
	_ = progress.Step(ctx, "manifest.evidence", "succeeded", "redacted evidence bundle prepared")
	_ = progress.Step(ctx, "manifest.provider", "running", "provider proposal requested")
	err := ai.Execute(ctx, operation.ID, input.ProposalID, input.Provider, input.Limits, input.ConsentSHA256)
	if err == nil {
		_ = progress.Step(ctx, "manifest.provider", "succeeded", "provider proposal validated and merged")
	}
	return err
}

func executeActionOperation(ctx context.Context, actions *actionsApplication.Service, operation domain.Operation, progress operations.Progress) error {
	var input struct {
		ActionID         string `json:"actionId"`
		ConfirmRisk      bool   `json:"confirmRisk"`
		AllowOutsideRoot bool   `json:"allowOutsideRoot"`
		ActorType        string `json:"actorType"`
		ActorID          string `json:"actorId"`
	}
	if err := json.Unmarshal(operation.Input, &input); err != nil || input.ActionID == "" {
		return errors.New("decode action operation")
	}
	_ = progress.Step(ctx, "action.execute", "running", "action authorized")
	err := actions.Execute(ctx, operation.ID, operation.ProjectID, input.ActionID, input.ActorType, input.ActorID, input.ConfirmRisk, input.AllowOutsideRoot)
	if err == nil {
		_ = progress.Step(ctx, "action.execute", "succeeded", "action completed")
	}
	return err
}

func executePluginOperation(ctx context.Context, plugins *pluginsApplication.Service, operation domain.Operation, progress operations.Progress) error {
	if plugins == nil {
		return errors.New("plugin service is unavailable")
	}
	var input struct {
		PluginID string          `json:"pluginId"`
		Action   string          `json:"action"`
		Input    json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(operation.Input, &input); err != nil || input.PluginID == "" || input.Action == "" {
		return errors.New("decode plugin operation")
	}
	_ = progress.Step(ctx, "plugin.execute", "running", "reviewed plugin action started")
	result, err := plugins.Operate(ctx, input.PluginID, operation.ProjectID, input.Action, input.Input)
	if err != nil {
		return err
	}
	_ = progress.Step(ctx, "plugin.execute", "succeeded", result.Summary)
	if result.Status == "partially_succeeded" {
		return operations.PartialSuccess(result.Summary)
	}
	if result.Status == "failed" {
		return errors.New("plugin reported action failure")
	}
	return nil
}

func executeWorkspaceOperation(ctx context.Context, workspaces *workspaceApplication.Service, recipes workspaceApplication.RecipeRunner, operation domain.Operation, progress operations.Progress) error {
	if workspaces == nil {
		return errors.New("workspace service is unavailable")
	}
	var input struct {
		Action            string                        `json:"action"`
		WorkspaceRevision int64                         `json:"workspaceRevision"`
		Policy            workspaceDomain.FailurePolicy `json:"policy"`
		ProfileID         string                        `json:"profileId"`
		RemoveData        bool                          `json:"removeData"`
		RunRecipes        bool                          `json:"runRecipes"`
	}
	if err := json.Unmarshal(operation.Input, &input); err != nil || "workspace."+input.Action != operation.Kind {
		return errors.New("decode workspace operation")
	}
	workspaceID := strings.TrimPrefix(operation.ProjectID, "workspace:")
	kind := workspaceDomain.ExecutionKind(input.Action)
	_, err := workspaces.Execute(ctx, workspaceID, workspaceApplication.ExecuteRequest{
		Kind: kind, Policy: input.Policy, ProfileID: input.ProfileID, RemoveData: input.RemoveData, ExpectedRevision: input.WorkspaceRevision,
	}, workspaceProgress{progress: progress})
	var executionErr *workspaceApplication.ExecutionError
	if errors.As(err, &executionErr) && executionErr.Partial() {
		return operations.PartialSuccess(executionErr.Error())
	}
	if err != nil {
		return err
	}
	if kind != workspaceDomain.ExecutionStart || !input.RunRecipes {
		return nil
	}
	_ = progress.Step(ctx, "workspace.recipes", "running", "running opt-in workspace recipes")
	if err := workspaces.ExecuteRecipes(ctx, workspaceID, recipes, input.WorkspaceRevision); err != nil {
		return err
	}
	_ = progress.Step(ctx, "workspace.recipes", "succeeded", "workspace recipes completed")
	return nil
}

func executeRuntimeOperation(
	ctx context.Context,
	service *runtimeApplication.Service,
	health requiredHealthWaiter,
	environments environmentLifecycle,
	operation domain.Operation,
	progress operations.Progress,
) error {
	var input struct {
		Action        string   `json:"action"`
		RemoveVolumes bool     `json:"removeVolumes"`
		Services      []string `json:"services"`
		Profiles      []string `json:"profiles"`
	}
	if err := json.Unmarshal(operation.Input, &input); err != nil {
		return fmt.Errorf("decode runtime operation: %w", err)
	}
	action, err := runtimeDomain.ParseAction(input.Action)
	if err != nil || operation.Kind != "runtime."+input.Action {
		return fmt.Errorf("invalid runtime operation kind %q", operation.Kind)
	}
	plan, err := service.PlanSelection(ctx, operation.ProjectID, action, input.RemoveVolumes, input.Services, input.Profiles)
	if err != nil {
		return err
	}
	plan.OperationID = operation.ID
	if err := service.Execute(ctx, plan, progress); err != nil {
		return err
	}
	if action == runtimeDomain.ActionStart || action == runtimeDomain.ActionRestart || action == runtimeDomain.ActionDeploy || action == runtimeDomain.ActionRebuild || action == runtimeDomain.ActionUnpause {
		if err := health.WaitRequired(ctx, operation.ProjectID); err != nil {
			return err
		}
		if environments != nil {
			return environments.Started(ctx, operation.ProjectID)
		}
	}
	if action == runtimeDomain.ActionStop || action == runtimeDomain.ActionKill || action == runtimeDomain.ActionTeardown || action == runtimeDomain.ActionPause {
		if environments != nil {
			return environments.Stopped(ctx, operation.ProjectID)
		}
	}
	return nil
}

type requiredHealthWaiter interface {
	WaitRequired(context.Context, string) error
}
