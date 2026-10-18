package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	operations "envarium.dev/envarium/internal/operations/application"
	sourcecontrol "envarium.dev/envarium/internal/sourcecontrol/application"
	"envarium.dev/envarium/internal/sourcecontrol/domain"
	"envarium.dev/envarium/internal/transport/contract/generated"
)

type branchService interface {
	Prepare(context.Context, string, string) (domain.BranchSwitchPlan, error)
}

func (h *handler) PlanGitBranchSwitch(w http.ResponseWriter, r *http.Request, projectID generated.ProjectId, _ generated.PlanGitBranchSwitchParams) {
	var request struct {
		Branch string `json:"branch"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF || request.Branch == "" || len(request.Branch) > 256 {
		writeProblem(w, r, 400, "REQUEST_INVALID", "Branch invalid", "Choose one existing local branch.")
		return
	}
	if h.branches == nil {
		writeProblem(w, r, 503, "GIT_SWITCH_UNAVAILABLE", "Branch switch unavailable", "The checkout service is not configured.")
		return
	}
	plan, err := h.branches.Prepare(r.Context(), projectID, request.Branch)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	writeJSON(w, 200, plan)
}
func (h *handler) CreateGitBranchSwitch(w http.ResponseWriter, r *http.Request, projectID generated.ProjectId, params generated.CreateGitBranchSwitchParams) {
	actor := identityFrom(r.Context())
	if actor.Access != accessBrowser && actor.Access != accessIPC {
		writeProblem(w, r, 403, "GIT_SWITCH_HUMAN_REQUIRED", "Local review required", "Branch switching requires a local user review.")
		return
	}
	var request struct {
		Plan        domain.BranchSwitchPlan `json:"plan"`
		ConfirmRisk bool                    `json:"confirmRisk"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF || request.Plan.ProjectID != projectID {
		writeProblem(w, r, 400, "REQUEST_INVALID", "Branch switch invalid", "Provide the reviewed plan for this project.")
		return
	}
	if !request.ConfirmRisk {
		writeProblem(w, r, 409, "GIT_SWITCH_CONFIRMATION_REQUIRED", "Review required", "Confirm the exact checkout change and separate restart review.")
		return
	}
	if h.branches == nil {
		writeProblem(w, r, 503, "GIT_SWITCH_UNAVAILABLE", "Branch switch unavailable", "The checkout service is not configured.")
		return
	}
	current, err := h.branches.Prepare(r.Context(), projectID, request.Plan.ToBranch)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if current != request.Plan {
		writeApplicationError(w, r, sourcecontrol.ErrCheckoutChanged)
		return
	}
	input, _ := json.Marshal(struct {
		Plan      domain.BranchSwitchPlan `json:"plan"`
		ActorType string                  `json:"actorType"`
	}{current, string(actor.Access)})
	operation, err := h.operations.Submit(r.Context(), operations.SubmitRequest{ProjectID: projectID, Kind: "git.switch", Input: input, IdempotencyKey: params.IdempotencyKey, ActorType: string(actor.Access), ActorID: actor.ActorID})
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, operationResponse(operation))
}
