package adapters

import (
	"context"

	diagnosticsApplication "envarium.dev/envarium/internal/diagnostics/application"
	operationsApplication "envarium.dev/envarium/internal/operations/application"
	operationsDomain "envarium.dev/envarium/internal/operations/domain"
)

type operationSubmitter interface {
	Submit(context.Context, operationsApplication.SubmitRequest) (operationsDomain.Operation, error)
}

// ActionSubmitter routes automation through the existing operation and action permission kernels.
type ActionSubmitter struct{ operations operationSubmitter }

// NewActionSubmitter creates an automation operation adapter.
func NewActionSubmitter(operations operationSubmitter) *ActionSubmitter {
	return &ActionSubmitter{operations: operations}
}

// SubmitAction emits only the existing action.run operation contract.
func (s *ActionSubmitter) SubmitAction(ctx context.Context, projectID, actionID, recipeID, diagnosisID string) (string, error) {
	operation, err := s.operations.Submit(ctx, operationsApplication.SubmitRequest{
		ProjectID: projectID, Kind: "action.run", IdempotencyKey: "automation:" + recipeID + ":" + diagnosisID,
		Input: diagnosticsApplication.ActionOperationInput(actionID, recipeID), ActorType: "automation", ActorID: recipeID,
	})
	return operation.ID, err
}
