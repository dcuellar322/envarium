package httpapi

import (
	runtime "envarium.dev/envarium/internal/runtime/application"
	sourcecontrol "envarium.dev/envarium/internal/sourcecontrol/application"
	"errors"
	"net/http"
)

func writeCheckoutError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case errors.Is(err, runtime.ErrNativeEnvironmentUnsupported):
		writeProblem(w, r, http.StatusConflict, "ENVIRONMENT_RUNTIME_UNSUPPORTED", "Native worktree starts unavailable", err.Error())
	case errors.Is(err, runtime.ErrCheckoutBusy):
		writeProblem(w, r, http.StatusServiceUnavailable, "CHECKOUT_BUSY", "Checkout coordination busy", err.Error())
	case errors.Is(err, sourcecontrol.ErrUnsafeCheckout), errors.Is(err, sourcecontrol.ErrCheckoutChanged), errors.Is(err, runtime.ErrCheckoutActive), errors.Is(err, runtime.ErrPlanStale):
		writeProblem(w, r, http.StatusConflict, "CHECKOUT_REVIEW_REQUIRED", "Checkout needs attention", err.Error())
	default:
		return false
	}
	return true
}
