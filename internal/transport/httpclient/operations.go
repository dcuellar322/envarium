package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

// Operations lists recent durable operations with an optional project filter.
func (c *Client) Operations(ctx context.Context, projectID string, limit int64) ([]generated.Operation, error) {
	params := &generated.ListOperationsParams{Limit: &limit}
	if projectID != "" {
		params.ProjectId = &projectID
	}
	response, err := c.generated.ListOperationsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list operations: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, apiError("list operations", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// Operation reads one durable operation.
func (c *Client) Operation(ctx context.Context, operationID string) (generated.Operation, error) {
	response, err := c.generated.GetOperationWithResponse(ctx, operationID)
	if err != nil {
		return generated.Operation{}, fmt.Errorf("get operation: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.Operation{}, apiError("get operation", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// CancelOperation requests durable idempotent cancellation.
func (c *Client) CancelOperation(ctx context.Context, operationID, idempotencyKey string) (generated.Operation, error) {
	response, err := c.generated.CancelOperationWithResponse(ctx, operationID, &generated.CancelOperationParams{IdempotencyKey: idempotencyKey})
	if err != nil {
		return generated.Operation{}, fmt.Errorf("cancel operation: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.Operation{}, apiError("cancel operation", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}
