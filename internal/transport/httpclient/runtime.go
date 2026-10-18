package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

// Runtime observes current project services through the configured driver.
func (c *Client) Runtime(ctx context.Context, projectID string) (generated.RuntimeObservation, error) {
	response, err := c.generated.GetProjectRuntimeWithResponse(ctx, projectID)
	if err != nil {
		return generated.RuntimeObservation{}, fmt.Errorf("observe project runtime: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.RuntimeObservation{}, apiError("observe project runtime", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// PlanRuntime previews one lifecycle action without executing it.
func (c *Client) PlanRuntime(ctx context.Context, projectID string, action generated.RuntimeAction, removeVolumes bool) (generated.RuntimePlan, error) {
	return c.PlanRuntimeServices(ctx, projectID, action, removeVolumes, nil)
}

// PlanRuntimeServices previews one lifecycle action for selected declared services.
func (c *Client) PlanRuntimeServices(ctx context.Context, projectID string, action generated.RuntimeAction, removeVolumes bool, services []string) (generated.RuntimePlan, error) {
	return c.PlanRuntimeSelection(ctx, projectID, action, removeVolumes, services, nil)
}

// PlanRuntimeSelection previews one lifecycle action for selected services and trusted Compose profiles.
func (c *Client) PlanRuntimeSelection(ctx context.Context, projectID string, action generated.RuntimeAction, removeVolumes bool, services, profiles []string) (generated.RuntimePlan, error) {
	request := generated.RuntimeActionRequest{Action: action, RemoveVolumes: &removeVolumes}
	if len(services) > 0 {
		request.Services = &services
	}
	if len(profiles) > 0 {
		request.Profiles = &profiles
	}
	response, err := c.generated.PlanProjectRuntimeWithResponse(ctx, projectID, generated.RuntimeActionRequest{
		Action: request.Action, RemoveVolumes: request.RemoveVolumes, Services: request.Services, Profiles: request.Profiles,
	})
	if err != nil {
		return generated.RuntimePlan{}, fmt.Errorf("plan project runtime: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.RuntimePlan{}, apiError("plan project runtime", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// CreateRuntimeOperation queues a durable lifecycle mutation.
func (c *Client) CreateRuntimeOperation(
	ctx context.Context,
	projectID string,
	action generated.RuntimeAction,
	removeVolumes bool,
	idempotencyKey string,
) (generated.Operation, error) {
	return c.CreateRuntimeOperationForServices(ctx, projectID, action, removeVolumes, nil, idempotencyKey)
}

// CreateRuntimeOperationForServices queues a lifecycle mutation for selected declared services.
func (c *Client) CreateRuntimeOperationForServices(
	ctx context.Context,
	projectID string,
	action generated.RuntimeAction,
	removeVolumes bool,
	services []string,
	idempotencyKey string,
) (generated.Operation, error) {
	return c.CreateRuntimeOperationSelection(ctx, projectID, action, removeVolumes, services, nil, idempotencyKey)
}

// CreateRuntimeOperationSelection queues a lifecycle mutation for selected services and trusted Compose profiles.
func (c *Client) CreateRuntimeOperationSelection(
	ctx context.Context,
	projectID string,
	action generated.RuntimeAction,
	removeVolumes bool,
	services, profiles []string,
	idempotencyKey string,
) (generated.Operation, error) {
	confirmRisk := action == generated.RuntimeActionTeardown
	request := generated.RuntimeActionRequest{Action: action, RemoveVolumes: &removeVolumes, ConfirmRisk: &confirmRisk}
	if len(services) > 0 {
		request.Services = &services
	}
	if len(profiles) > 0 {
		request.Profiles = &profiles
	}
	response, err := c.generated.CreateProjectOperationWithResponse(
		ctx, projectID, &generated.CreateProjectOperationParams{IdempotencyKey: idempotencyKey},
		request,
	)
	if err != nil {
		return generated.Operation{}, fmt.Errorf("create runtime operation: %w", err)
	}
	if response.StatusCode() != http.StatusAccepted || response.JSON202 == nil {
		return generated.Operation{}, apiError("create runtime operation", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON202, nil
}

// RuntimeLogs reads a bounded Docker log snapshot.
func (c *Client) RuntimeLogs(ctx context.Context, projectID, service, since, runID, operationID string, tail int) ([]generated.RuntimeLogEntry, error) {
	params := &generated.GetProjectLogsParams{Tail: &tail}
	if service != "" {
		params.Service = &service
	}
	if since != "" {
		params.Since = &since
	}
	if runID != "" {
		params.RunId = &runID
	}
	if operationID != "" {
		params.OperationId = &operationID
	}
	response, err := c.generated.GetProjectLogsWithResponse(ctx, projectID, params)
	if err != nil {
		return nil, fmt.Errorf("read project logs: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, apiError("read project logs", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// ExportRuntimeLogs returns a bounded redacted plain-text or NDJSON export.
func (c *Client) ExportRuntimeLogs(ctx context.Context, projectID, service, runID, operationID string, format generated.ExportProjectLogsParamsFormat) ([]byte, error) {
	params := &generated.ExportProjectLogsParams{Format: format}
	if service != "" {
		params.Service = &service
	}
	if runID != "" {
		params.RunId = &runID
	}
	if operationID != "" {
		params.OperationId = &operationID
	}
	response, err := c.generated.ExportProjectLogsWithResponse(ctx, projectID, params)
	if err != nil {
		return nil, fmt.Errorf("export project logs: %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, apiError("export project logs", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return response.Body, nil
}

// RuntimeMetrics reads current Compose resource samples.
func (c *Client) RuntimeMetrics(ctx context.Context, projectID, service string) ([]generated.RuntimeMetricSample, error) {
	params := &generated.GetProjectMetricsParams{}
	if service != "" {
		params.Service = &service
	}
	response, err := c.generated.GetProjectMetricsWithResponse(ctx, projectID, params)
	if err != nil {
		return nil, fmt.Errorf("read project metrics: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, apiError("read project metrics", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}
