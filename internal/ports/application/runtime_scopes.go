package application

import "context"

// RuntimeScope names an independently owned runtime within a trusted catalog project.
type RuntimeScope struct{ ID, Name string }

// RuntimeScopeSource supplies additional trusted scopes without exposing their storage.
type RuntimeScopeSource interface {
	PortRuntimeScopes(context.Context) ([]RuntimeScope, error)
}
