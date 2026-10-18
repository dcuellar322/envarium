package mcpserver

import (
	"context"
	"errors"
	"testing"

	agents "envarium.dev/envarium/internal/agents/application"
	"envarium.dev/envarium/internal/transport/contract/generated"
)

type workspaceScopeBackend struct {
	fakeBackend
	workspace generated.Workspace
	err       error
}

func (f *workspaceScopeBackend) Workspaces(context.Context) ([]generated.Workspace, error) {
	return []generated.Workspace{f.workspace}, f.err
}
func (f *workspaceScopeBackend) Workspace(context.Context, string) (generated.Workspace, error) {
	return f.workspace, f.err
}
func (f *workspaceScopeBackend) CreateWorkspaceOperation(context.Context, string, generated.WorkspaceOperationRequest, string) (generated.Operation, error) {
	return f.operation, nil
}

func TestScopedWorkspaceOperationGetWaitCancel(t *testing.T) {
	ctx := context.Background()
	id := "workspace-one"
	backend := &workspaceScopeBackend{workspace: generated.Workspace{Id: id, Members: []generated.WorkspaceMember{{ProjectId: "project-1"}, {ProjectId: "project-2"}}}}
	backend.operation = generated.Operation{Id: "op", ProjectId: "workspace:" + id, WorkspaceId: &id, WorkspaceMemberIds: &[]string{"project-1", "project-2"}, State: generated.OperationStateSucceeded}
	scope, _ := agents.NewScope("fixture", "agent", agents.ProfileDevelop, []string{"project-1", "project-2"})
	server := &Server{backend: backend, scope: scope}
	if _, _, err := server.operationGet(ctx, nil, operationInput{OperationID: "op"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.operationWait(ctx, nil, operationWaitInput{OperationID: "op", TimeoutSeconds: 1}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.operationCancel(ctx, nil, cancelInput{OperationID: "op", RequestID: "request-1"}); err != nil {
		t.Fatal(err)
	}
	if !backend.cancelled {
		t.Fatal("authorized cancellation not submitted")
	}
	backend.workspace.Members = []generated.WorkspaceMember{{ProjectId: "project-1"}}
	server.scope, _ = agents.NewScope("fixture", "agent", agents.ProfileDevelop, []string{"project-1"})
	if err := server.authorizeOperation(ctx, backend.operation, ""); !errors.Is(err, agents.ErrPermissionDenied) {
		t.Fatalf("cross-member read=%v", err)
	}
	if err := server.authorizeOperation(ctx, backend.operation, agents.CapabilityOperationCancel); !errors.Is(err, agents.ErrPermissionDenied) {
		t.Fatalf("cross-member cancel=%v", err)
	}
	backend.operation.WorkspaceId = nil
	if err := server.authorizeOperation(ctx, backend.operation, ""); !errors.Is(err, agents.ErrPermissionDenied) {
		t.Fatalf("synthetic prefix bypass=%v", err)
	}
	backend.operation.WorkspaceId = &id
	server.scope, _ = agents.NewScope("fixture", "agent", agents.ProfileDevelop, []string{"project-1", "project-2"})
	backend.err = errors.New("workspace deleted")
	if err := server.authorizeOperation(ctx, backend.operation, ""); err != nil {
		t.Fatalf("captured scope depends on current workspace: %v", err)
	}
	backend.operation.WorkspaceMemberIds = nil
	if err := server.authorizeOperation(ctx, backend.operation, ""); err == nil {
		t.Fatal("missing captured scope accepted")
	}
}
