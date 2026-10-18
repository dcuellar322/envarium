package application

import (
	"context"
	"errors"
	"testing"

	"envarium.dev/envarium/internal/workspace/domain"
)

func TestReviewedWorkspaceRevisionRejectsChangedExecutionAndRecipes(t *testing.T) {
	workspace := orchestrationWorkspace()
	workspace.Revision = 2
	operator := &fakeProjectOperator{}
	service := NewService(newFakeRepository(workspace), operator, noHealthGate{}, allowMembers{})
	if _, err := service.Execute(context.Background(), workspace.ID, ExecuteRequest{Kind: domain.ExecutionStart, ExpectedRevision: 1}, nil); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("execution=%v", err)
	}
	if len(operator.eventsCopy()) != 0 {
		t.Fatal("changed workspace started projects")
	}
	if err := service.ExecuteRecipes(context.Background(), workspace.ID, nil, 1); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("recipes=%v", err)
	}
	if _, err := service.Execute(context.Background(), workspace.ID, ExecuteRequest{Kind: domain.ExecutionStart, ExpectedRevision: 2}, nil); err != nil {
		t.Fatalf("reviewed revision=%v", err)
	}
}
