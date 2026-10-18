package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

// CreateManifestProposal scans a repository through the privileged local API.
func (c *Client) CreateManifestProposal(ctx context.Context, path, idempotencyKey string) (generated.ManifestProposal, error) {
	return c.CreateManifestProposalWithRootOverride(ctx, path, idempotencyKey, false)
}

// CreateManifestProposalWithRootOverride carries an explicit one-off approval
// when a local user intentionally scans outside configured roots.
func (c *Client) CreateManifestProposalWithRootOverride(ctx context.Context, path, idempotencyKey string, allowOutsideRoots bool) (generated.ManifestProposal, error) {
	response, err := c.generated.CreateManifestProposalWithResponse(ctx,
		&generated.CreateManifestProposalParams{IdempotencyKey: idempotencyKey},
		generated.CreateManifestProposalRequest{Path: path, AllowOutsideRoots: &allowOutsideRoots},
	)
	if err != nil {
		return generated.ManifestProposal{}, fmt.Errorf("create manifest proposal: %w", err)
	}
	if response.StatusCode() != http.StatusCreated || response.JSON201 == nil {
		return generated.ManifestProposal{}, unexpected("create manifest proposal", response.StatusCode())
	}
	return *response.JSON201, nil
}

// ManifestProposal reads one deterministic proposal and its owning project ID.
func (c *Client) ManifestProposal(ctx context.Context, proposalID string) (generated.ManifestProposal, error) {
	response, err := c.generated.GetManifestProposalWithResponse(ctx, proposalID)
	if err != nil {
		return generated.ManifestProposal{}, fmt.Errorf("read manifest proposal: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ManifestProposal{}, apiError("read manifest proposal", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// ValidateManifestProposal reruns proposal validation through the local API.
func (c *Client) ValidateManifestProposal(ctx context.Context, proposalID, idempotencyKey string) (generated.ManifestProposal, error) {
	response, err := c.generated.ValidateManifestProposalWithResponse(ctx, proposalID, &generated.ValidateManifestProposalParams{IdempotencyKey: idempotencyKey})
	if err != nil {
		return generated.ManifestProposal{}, fmt.Errorf("validate manifest proposal: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ManifestProposal{}, unexpected("validate manifest proposal", response.StatusCode())
	}
	return *response.JSON200, nil
}

// AcceptManifestProposal records a trust decision through the local API.
func (c *Client) AcceptManifestProposal(ctx context.Context, proposalID, idempotencyKey string) (generated.AcceptedManifestProposal, error) {
	response, err := c.generated.AcceptManifestProposalWithResponse(ctx, proposalID, &generated.AcceptManifestProposalParams{IdempotencyKey: idempotencyKey})
	if err != nil {
		return generated.AcceptedManifestProposal{}, fmt.Errorf("accept manifest proposal: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.AcceptedManifestProposal{}, unexpected("accept manifest proposal", response.StatusCode())
	}
	return *response.JSON200, nil
}

// ExplainManifest returns effective fields and provenance.
func (c *Client) ExplainManifest(ctx context.Context, projectID string) (generated.EffectiveManifest, error) {
	response, err := c.generated.ExplainProjectManifestWithResponse(ctx, projectID)
	if err != nil {
		return generated.EffectiveManifest{}, fmt.Errorf("explain project manifest: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.EffectiveManifest{}, unexpected("explain project manifest", response.StatusCode())
	}
	return *response.JSON200, nil
}

// DiffManifest compares the accepted and effective manifests.
func (c *Client) DiffManifest(ctx context.Context, projectID string) (generated.ManifestDiff, error) {
	response, err := c.generated.DiffProjectManifestWithResponse(ctx, projectID)
	if err != nil {
		return generated.ManifestDiff{}, fmt.Errorf("diff project manifest: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ManifestDiff{}, unexpected("diff project manifest", response.StatusCode())
	}
	return *response.JSON200, nil
}

// ValidateProjectManifest validates the fully resolved project manifest.
func (c *Client) ValidateProjectManifest(ctx context.Context, projectID string) (generated.ManifestValidation, error) {
	response, err := c.generated.ValidateProjectManifestWithResponse(ctx, projectID)
	if err != nil {
		return generated.ManifestValidation{}, fmt.Errorf("validate project manifest: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ManifestValidation{}, unexpected("validate project manifest", response.StatusCode())
	}
	return *response.JSON200, nil
}
