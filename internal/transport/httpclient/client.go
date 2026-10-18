// Package httpclient is the typed local API adapter used by CLI clients.
package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"envarium.dev/envarium/internal/platform/localipc"
	"envarium.dev/envarium/internal/transport/contract/generated"
)

// Client wraps the generated transport without leaking response mechanics.
type Client struct {
	generated *generated.ClientWithResponses
}

const (
	actorTypeHeader = "X-Envarium-Actor-Type"
	actorIDHeader   = "X-Envarium-Actor-ID"
)

// APIError preserves a stable problem code and HTTP status for CLI exit mapping.
type APIError struct {
	Operation string
	Status    int
	Code      string
	Detail    string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Operation, e.Detail)
	}
	return fmt.Sprintf("%s: unexpected HTTP %d", e.Operation, e.Status)
}

// NewIPC creates a typed client over privileged local IPC.
func NewIPC(address string) (*Client, error) {
	return newIPC(address, "", "")
}

// NewIPCForAgent creates a privileged local client whose mutations retain agent identity in durable audits.
func NewIPCForAgent(address, actorID string) (*Client, error) {
	if actorID == "" {
		return nil, fmt.Errorf("create agent IPC client: actor identity is required")
	}
	return newIPC(address, "agent", actorID)
}

func newIPC(address, actorType, actorID string) (*Client, error) {
	options := []generated.ClientOption{generated.WithHTTPClient(localipc.HTTPClient(address))}
	if actorType != "" {
		options = append(options, generated.WithRequestEditorFn(func(_ context.Context, request *http.Request) error {
			request.Header.Set(actorTypeHeader, actorType)
			request.Header.Set(actorIDHeader, actorID)
			return nil
		}))
	}
	client, err := generated.NewClientWithResponses(
		"http://envarium.local/api/v1",
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("create local IPC client: %w", err)
	}
	return &Client{generated: client}, nil
}

// Health returns the latest persisted and currently evaluated project health.
func (c *Client) Health(ctx context.Context, projectID string) (generated.ProjectHealth, error) {
	response, err := c.generated.GetProjectHealthWithResponse(ctx, projectID)
	if err != nil {
		return generated.ProjectHealth{}, fmt.Errorf("read project health: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.ProjectHealth{}, apiError("read project health", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// System returns the generated system contract.
func (c *Client) System(ctx context.Context) (generated.SystemInfo, error) {
	response, err := c.generated.GetSystemWithResponse(ctx)
	if err != nil {
		return generated.SystemInfo{}, fmt.Errorf("request system status: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.SystemInfo{}, fmt.Errorf("request system status: unexpected HTTP %d", response.StatusCode())
	}
	return *response.JSON200, nil
}

// Host returns the bounded aggregate host observation used by native adapters.
func (c *Client) Host(ctx context.Context) (generated.HostObservation, error) {
	response, err := c.generated.GetHostWithResponse(ctx)
	if err != nil {
		return generated.HostObservation{}, fmt.Errorf("request host observation: %w", err)
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return generated.HostObservation{}, apiError("request host observation", response.StatusCode(), response.ApplicationproblemJSONDefault)
	}
	return *response.JSON200, nil
}

// BrowserBootstrap requests a one-time browser credential over local IPC.
func (c *Client) BrowserBootstrap(ctx context.Context) (generated.BrowserBootstrap, error) {
	response, err := c.generated.CreateBrowserBootstrapTokenWithResponse(ctx)
	if err != nil {
		return generated.BrowserBootstrap{}, fmt.Errorf("request browser bootstrap: %w", err)
	}
	if response.StatusCode() != http.StatusCreated || response.JSON201 == nil {
		return generated.BrowserBootstrap{}, fmt.Errorf("request browser bootstrap: unexpected HTTP %d", response.StatusCode())
	}
	return *response.JSON201, nil
}

func unexpected(operation string, status int) error {
	return fmt.Errorf("%s: unexpected HTTP %d", operation, status)
}

func apiError(operation string, status int, problem *generated.Problem) error {
	result := &APIError{Operation: operation, Status: status}
	if problem != nil {
		result.Code = problem.Code
		if problem.Detail != nil {
			result.Detail = *problem.Detail
		}
	}
	return result
}
