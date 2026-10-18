package httpapi

import (
	"slices"
	"testing"

	"envarium.dev/envarium/internal/operations/domain"
)

func TestOperationResponseRetainsSubmittedWorkspaceMembers(t *testing.T) {
	operation := domain.Operation{WorkspaceID: "workspace-one", Input: []byte(`{"workspaceRevision":1,"workspaceMemberIds":["project-a","project-b"]}`)}
	response := operationResponse(operation)
	if response.WorkspaceMemberIds == nil || !slices.Equal(*response.WorkspaceMemberIds, []string{"project-a", "project-b"}) {
		t.Fatalf("scope=%#v", response.WorkspaceMemberIds)
	}
}
