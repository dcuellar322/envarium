package mcpserver

import (
	"context"
	"errors"

	agents "envarium.dev/envarium/internal/agents/application"
	"envarium.dev/envarium/internal/transport/contract/generated"
)

// Workspace operations carry an explicit workspace identity. Resolve members
// from the immutable submission scope; a synthetic project prefix grants no access.
func (s *Server) authorizeOperation(ctx context.Context, operation generated.Operation, capability agents.Capability) error {
	if operation.WorkspaceId == nil {
		if capability == "" {
			return s.scope.AuthorizeRead(operation.ProjectId)
		}
		return s.scope.Authorize(capability, operation.ProjectId)
	}
	if operation.WorkspaceMemberIds == nil || len(*operation.WorkspaceMemberIds) == 0 {
		return errors.New("workspace operation has no captured participant scope; resubmit it")
	}
	for _, memberID := range *operation.WorkspaceMemberIds {
		projectID, err := s.workspaceMemberProject(ctx, memberID)
		if err != nil {
			return err
		}
		if capability == "" {
			err = s.scope.AuthorizeRead(projectID)
		} else {
			err = s.scope.Authorize(capability, projectID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
