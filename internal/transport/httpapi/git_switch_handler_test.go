package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
	"envarium.dev/envarium/internal/transport/contract/generated"
)

type branchServiceStub struct{ plan domain.BranchSwitchPlan }

func (s branchServiceStub) Prepare(context.Context, string, string) (domain.BranchSwitchPlan, error) {
	return s.plan, nil
}
func TestBranchSwitchRequiresLocalActorExactPlanAndConfirmation(t *testing.T) {
	t.Parallel()
	plan := domain.BranchSwitchPlan{ProjectID: "p", Root: "/repo", FromBranch: "main", FromHead: "a", ToBranch: "feature", ToHead: "b", Risk: "mutating", ReviewRestart: true}
	for _, mode := range []string{"agent", "missing-confirmation", "stale", "trailing-json", "valid"} {
		t.Run(mode, func(t *testing.T) {
			operations := &recordingOperations{}
			handler := &handler{branches: branchServiceStub{plan}, operations: operations}
			input := plan
			confirm := true
			actor := accessBrowser
			status := http.StatusAccepted
			switch mode {
			case "agent":
				actor = accessKind("agent")
				status = 403
			case "missing-confirmation":
				confirm = false
				status = 409
			case "stale":
				input.ToHead = "old"
				status = 409
			case "trailing-json":
				status = 400
			}
			body, _ := json.Marshal(struct {
				Plan        domain.BranchSwitchPlan `json:"plan"`
				ConfirmRisk bool                    `json:"confirmRisk"`
			}{input, confirm})
			if mode == "trailing-json" {
				body = append(body, []byte(` {}`)...)
			}
			request := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p/git/branch-switch", strings.NewReader(string(body)))
			request = request.WithContext(context.WithValue(request.Context(), securityContextKey{}, requestIdentity{Access: actor, ActorID: "local-user"}))
			response := httptest.NewRecorder()
			handler.CreateGitBranchSwitch(response, request, "p", generated.CreateGitBranchSwitchParams{IdempotencyKey: "reviewed-branch-switch"})
			if response.Code != status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body)
			}
			if status == 202 {
				if operations.request.Kind != "git.switch" || operations.request.ActorID != "local-user" {
					t.Fatalf("operation unaudited: %#v", operations.request)
				}
			} else if operations.request.Kind != "" {
				t.Fatal("unauthorized transition queued")
			}
		})
	}
}
