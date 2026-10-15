package adapters

import (
	agents "envarium.dev/envarium/internal/agents/application"
	manifestApplication "envarium.dev/envarium/internal/manifest/application"
	manifestDomain "envarium.dev/envarium/internal/manifest/domain"
)

// ManifestValidator adapts the canonical validation use case to assisted onboarding.
type ManifestValidator struct{}

// Validate performs the same side-effect-free checks used before catalog acceptance.
func (ManifestValidator) Validate(root string, candidate manifestDomain.Manifest) agents.CandidateValidation {
	result := manifestApplication.Validate(root, candidate)
	return agents.CandidateValidation{Valid: result.Valid, Errors: result.Errors, Warnings: result.Warnings}
}
