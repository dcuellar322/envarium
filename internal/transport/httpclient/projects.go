package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

// Projects lists registered projects through the local API.
func (c *Client) Projects(ctx context.Context) ([]generated.Project, error) {
	response, err := c.generated.ListProjectsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, unexpected("list projects", response.StatusCode())
	}
	return *response.JSON200, nil
}

// Project reads one registered project by opaque ID.
func (c *Client) Project(ctx context.Context, projectID string) (generated.Project, error) {
	response, err := c.generated.GetProjectWithResponse(ctx, projectID)
	if err != nil {
		return generated.Project{}, fmt.Errorf("get project: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.Project{}, apiError("get project", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// TrustProject accepts the latest valid proposal for one project.
func (c *Client) TrustProject(ctx context.Context, projectID, idempotencyKey string) (generated.AcceptedManifestProposal, error) {
	response, err := c.generated.TrustProjectWithResponse(ctx, projectID, &generated.TrustProjectParams{IdempotencyKey: idempotencyKey})
	if err != nil {
		return generated.AcceptedManifestProposal{}, fmt.Errorf("trust project: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.AcceptedManifestProposal{}, apiError("trust project", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// RemoveProject removes catalog data without touching the repository.
func (c *Client) RemoveProject(ctx context.Context, projectID, idempotencyKey string) error {
	response, err := c.generated.RemoveProjectWithResponse(ctx, projectID, &generated.RemoveProjectParams{IdempotencyKey: idempotencyKey})
	if err != nil {
		return fmt.Errorf("remove project: %w", err)
	}
	if response.StatusCode() != http.StatusNoContent {
		return apiError("remove project", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return nil
}
