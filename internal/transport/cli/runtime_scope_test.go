package cli

import (
	"context"
	"envarium.dev/envarium/internal/transport/contract/generated"
	"errors"
	"testing"
)

type runtimeScopeStub struct {
	environment  generated.ProjectEnvironment
	err          error
	catalogReads int
}

func (s *runtimeScopeStub) Projects(ctx context.Context) ([]generated.Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.catalogReads++
	return []generated.Project{{Id: "project-p", Slug: "p"}}, s.err
}
func (s *runtimeScopeStub) Environment(ctx context.Context, _ string) (generated.ProjectEnvironment, error) {
	if err := ctx.Err(); err != nil {
		return generated.ProjectEnvironment{}, err
	}
	return s.environment, s.err
}
func TestRuntimeSelectionAcceptsExactEnvironmentAndKeepsCatalogSelection(t *testing.T) {
	client := &runtimeScopeStub{environment: generated.ProjectEnvironment{Id: "env-f", Name: "feature"}}
	scope, err := resolveRuntimeScope(context.Background(), client, "env-f")
	if err != nil || scope.ID != "env-f" || scope.Name != "feature" || client.catalogReads != 1 {
		t.Fatal("environment not selected directly", scope, err)
	}
	scope, err = resolveRuntimeScope(context.Background(), client, "p")
	if err != nil || scope.ID != "project-p" {
		t.Fatal("catalog selection changed", scope, err)
	}
	client.err = errors.New("not registered")
	if _, err := resolveRuntimeScope(context.Background(), client, "env-gone"); err == nil || client.catalogReads != 3 {
		t.Fatal("missing environment fell back to another project", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := resolveRuntimeScope(ctx, client, "env-f"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled selection proceeded", err)
	}
}
