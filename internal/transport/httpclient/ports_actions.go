package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

// PortRegistry reconciles declarations, leases, and live listeners.
func (c *Client) PortRegistry(ctx context.Context) (generated.PortRegistry, error) {
	response, err := c.generated.GetPortRegistryWithResponse(ctx)
	if err != nil {
		return generated.PortRegistry{}, fmt.Errorf("read port registry: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.PortRegistry{}, apiError("read port registry", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// SuggestPort returns the first available port in a bounded range.
func (c *Client) SuggestPort(ctx context.Context, start, end int, protocol, projectID string, excluded []int, idempotencyKey string) (generated.PortSuggestion, error) {
	request := generated.PortSuggestionRequest{RangeStart: start, RangeEnd: end, Protocol: generated.PortSuggestionRequestProtocol(protocol)}
	if projectID != "" {
		request.ProjectId = &projectID
	}
	if len(excluded) > 0 {
		request.Excluded = &excluded
	}
	response, err := c.generated.CreatePortSuggestionWithResponse(ctx, &generated.CreatePortSuggestionParams{IdempotencyKey: idempotencyKey}, request)
	if err != nil {
		return generated.PortSuggestion{}, fmt.Errorf("suggest port: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.PortSuggestion{}, apiError("suggest port", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// GitState reads a fresh repository snapshot.
func (c *Client) GitState(ctx context.Context, projectID string) (generated.GitState, error) {
	response, err := c.generated.GetProjectGitWithResponse(ctx, projectID)
	if err != nil {
		return generated.GitState{}, fmt.Errorf("read project Git state: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.GitState{}, apiError("read project Git state", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// ProjectActions lists the trusted quick-action contract.
func (c *Client) ProjectActions(ctx context.Context, projectID string) (generated.ProjectActions, error) {
	response, err := c.generated.ListProjectActionsWithResponse(ctx, projectID)
	if err != nil {
		return generated.ProjectActions{}, fmt.Errorf("list project actions: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ProjectActions{}, apiError("list project actions", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// CreateActionOperation queues one durable audited action.
func (c *Client) CreateActionOperation(ctx context.Context, projectID, actionID string, confirm, allowOutside bool, idempotencyKey string) (generated.Operation, error) {
	request := generated.ActionExecutionRequest{ConfirmRisk: &confirm, AllowOutsideRoot: &allowOutside}
	response, err := c.generated.CreateActionOperationWithResponse(ctx, projectID, actionID,
		&generated.CreateActionOperationParams{IdempotencyKey: idempotencyKey}, request)
	if err != nil {
		return generated.Operation{}, fmt.Errorf("create action operation: %w", err)
	}
	if response.StatusCode() != http.StatusAccepted || response.JSON202 == nil {
		return generated.Operation{}, apiError("create action operation", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON202, nil
}
