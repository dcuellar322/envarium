// Package adapters observes runtime and operating-system port bindings.
package adapters

import (
	"context"
	"errors"
	"time"

	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	portsApplication "envarium.dev/envarium/internal/ports/application"
	"envarium.dev/envarium/internal/ports/domain"
	runtimeDomain "envarium.dev/envarium/internal/runtime/domain"
)

type projectLister interface {
	ListProjects(context.Context) ([]catalogDomain.Project, error)
}

type runtimeObserver interface {
	Inspect(context.Context, string) (runtimeDomain.Observation, error)
}

// RuntimeBindings identifies Docker/process published ports already attributed to a project.
type RuntimeBindings struct {
	scopes   portsApplication.RuntimeScopeSource
	projects projectLister
	runtime  runtimeObserver
	now      func() time.Time
}

// NewRuntimeBindings creates an adapter over driver-neutral runtime observations.
func NewRuntimeBindings(projects projectLister, runtime runtimeObserver) *RuntimeBindings {
	return &RuntimeBindings{projects: projects, runtime: runtime, now: time.Now}
}

// Facts returns currently published ports attributed to trusted projects.
func (s *RuntimeBindings) Facts(ctx context.Context) ([]domain.Fact, error) {
	projects, err := s.projects.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	scopes := []portsApplication.RuntimeScope{}
	for _, project := range projects {
		if project.TrustState == catalogDomain.TrustTrusted {
			scopes = append(scopes, portsApplication.RuntimeScope{ID: project.ID, Name: project.DisplayName})
		}
	}
	if s.scopes != nil {
		additional, err := s.scopes.PortRuntimeScopes(ctx)
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, additional...)
	}
	var facts []domain.Fact
	var observationErrors []error
	for _, project := range scopes {
		observation, observeErr := s.observeBindings(ctx, project.ID)
		if observeErr != nil {
			observationErrors = append(observationErrors, observeErr)
			continue
		}
		for _, service := range observation.Services {
			// Stopped process observations retain declared ports for planning;
			// those are configuration, not evidence of a live host binding.
			if service.State != "running" && service.State != "paused" && service.State != "restarting" {
				continue
			}
			for _, port := range service.Ports {
				if port.HostPort == 0 {
					continue
				}
				host := port.HostIP
				if host == "" {
					host = "0.0.0.0"
				}
				facts = append(facts, domain.Fact{
					ID: runtimeFactID(project.ID, service.ID, host, port), Kind: domain.KindBinding,
					ProjectID: project.ID, ProjectName: project.Name, ServiceID: service.ID,
					Host: host, Port: port.HostPort, Target: port.ContainerPort, Protocol: port.Protocol,
					Source: string(observation.Driver), Evidence: "live runtime published port", ObservedAt: s.now().UTC(),
				})
			}
		}
	}
	return facts, errors.Join(observationErrors...)
}

func runtimeFactID(projectID, serviceID, host string, port runtimeDomain.PublishedPort) string {
	return stableID("runtime", projectID, serviceID, host, port.Protocol, port.HostPort)
}

func (s *RuntimeBindings) observeBindings(ctx context.Context, id string) (runtimeDomain.Observation, error) {
	if observer, ok := s.runtime.(interface {
		InspectBindings(context.Context, string) (runtimeDomain.Observation, error)
	}); ok {
		return observer.InspectBindings(ctx, id)
	}
	return s.runtime.Inspect(ctx, id)
}

// WithScopes includes independently owned worktree bindings.
func (s *RuntimeBindings) WithScopes(source portsApplication.RuntimeScopeSource) *RuntimeBindings {
	s.scopes = source
	return s
}
