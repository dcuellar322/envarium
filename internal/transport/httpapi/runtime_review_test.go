package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"envarium.dev/envarium/internal/runtime/domain"
	"envarium.dev/envarium/internal/transport/contract/generated"
)

type reviewedRuntimeStub struct{ runtimeStub }

func (reviewedRuntimeStub) PlanSelection(ctx context.Context, id string, action domain.Action, remove bool, services, profiles []string) (domain.Plan, error) {
	plan, err := (runtimeStub{}).PlanSelection(ctx, id, action, remove, services, profiles)
	plan.SourceCheckoutRevision = "fresh-checkout"
	plan.SourceManifestHash = "fresh-manifest"
	return plan, err
}
func TestRuntimeQueueBindsReviewedCheckoutAndManifest(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"stale-checkout", "stale-manifest", "current", "direct"} {
		t.Run(mode, func(t *testing.T) {
			operations := &recordingOperations{}
			handler := &handler{runtime: reviewedRuntimeStub{}, operations: operations}
			revision, hash := "fresh-checkout", "fresh-manifest"
			status := 202
			switch mode {
			case "stale-checkout":
				revision = "old"
				status = 409
			case "stale-manifest":
				hash = "old"
				status = 409
			case "direct":
				revision = ""
				hash = ""
			}
			body := map[string]any{"action": "start", "confirmRisk": true}
			if revision != "" {
				body["expectedCheckoutRevision"] = revision
			}
			if hash != "" {
				body["expectedManifestHash"] = hash
			}
			raw, _ := json.Marshal(body)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p/operations", strings.NewReader(string(raw)))
			response := httptest.NewRecorder()
			handler.CreateProjectOperation(response, request, "p", generated.CreateProjectOperationParams{IdempotencyKey: "review-runtime-context"})
			if response.Code != status {
				t.Fatalf("status=%d %s", response.Code, response.Body)
			}
			if status == 202 {
				var input map[string]any
				if err := json.Unmarshal(operations.request.Input, &input); err != nil {
					t.Fatal(err)
				}
				if input["expectedCheckoutRevision"] != "fresh-checkout" || input["expectedManifestHash"] != "fresh-manifest" {
					t.Fatalf("queued context unbound: %#v", input)
				}
			} else if operations.request.Kind != "" {
				t.Fatal("stale intent queued")
			}
		})
	}
}
