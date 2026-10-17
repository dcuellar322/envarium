package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

// StatusBatch reads current runtime/health facts through the bounded facade.
func (c *Client) StatusBatch(ctx context.Context, ids []string) (generated.ProjectObservationBatch, error) {
	response, err := c.generated.GetProjectObservationsWithResponse(ctx, &generated.GetProjectObservationsParams{
		Kind: generated.GetProjectObservationsParamsKind("status"), ProjectId: ids,
	})
	if err != nil {
		return generated.ProjectObservationBatch{}, fmt.Errorf("read project status batch: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ProjectObservationBatch{}, apiError("read project status batch", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}
