package application

import (
	"maps"
	"slices"

	"envarium.dev/envarium/internal/manifest/domain"
)

// cloneResolution keeps service-owned typed results immutable to callers without
// repeatedly serializing/decoding their complete provenance and declarations.
func cloneResolution(value EffectiveManifest) EffectiveManifest {
	result := value
	result.Provenance = maps.Clone(value.Provenance)
	result.Sources = slices.Clone(value.Sources)
	m := &result.Manifest
	m.Metadata.Tags = slices.Clone(m.Metadata.Tags)
	if c := m.Runtime.Compose; c != nil {
		cloned := *c
		cloned.Files = slices.Clone(c.Files)
		cloned.Profiles = slices.Clone(c.Profiles)
		m.Runtime.Compose = &cloned
	}
	if p := m.Runtime.Process; p != nil {
		cloned := *p
		cloned.Environment = maps.Clone(p.Environment)
		cloned.Secrets = maps.Clone(p.Secrets)
		cloned.Processes = slices.Clone(p.Processes)
		for i := range cloned.Processes {
			process := &cloned.Processes[i]
			process.Command = slices.Clone(process.Command)
			process.Environment = maps.Clone(process.Environment)
			process.Secrets = maps.Clone(process.Secrets)
		}
		m.Runtime.Process = &cloned
	}
	m.Lifecycle.Start = cloneInvocation(m.Lifecycle.Start)
	m.Lifecycle.Stop = cloneInvocation(m.Lifecycle.Stop)
	m.Lifecycle.Restart = cloneInvocation(m.Lifecycle.Restart)
	m.Lifecycle.Rebuild = cloneInvocation(m.Lifecycle.Rebuild)
	m.Lifecycle.Teardown = cloneInvocation(m.Lifecycle.Teardown)
	m.Services = slices.Clone(m.Services)
	for i := range m.Services {
		service := &m.Services[i]
		service.Dependencies = slices.Clone(service.Dependencies)
		service.HealthChecks = slices.Clone(service.HealthChecks)
		for j := range service.HealthChecks {
			check := &service.HealthChecks[j]
			check.Command = slices.Clone(check.Command)
			check.Members = slices.Clone(check.Members)
		}
	}
	m.Ports = slices.Clone(m.Ports)
	m.Endpoints = slices.Clone(m.Endpoints)
	m.Actions = slices.Clone(m.Actions)
	for i := range m.Actions {
		action := &m.Actions[i]
		action.Command = slices.Clone(action.Command)
		action.Environment = maps.Clone(action.Environment)
	}
	return result
}

func cloneInvocation(value *domain.Invocation) *domain.Invocation {
	if value == nil {
		return nil
	}
	result := *value
	if value.Options != nil {
		result.Options = cloneJSONValue(value.Options).(map[string]any)
	}
	return &result
}

// Invocation options have passed JSON decoding before entering the resolution
// cache, so their only mutable values are JSON objects and arrays.
func cloneJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		result := maps.Clone(v)
		for key, child := range result {
			result[key] = cloneJSONValue(child)
		}
		return result
	case []any:
		result := slices.Clone(v)
		for i, child := range result {
			result[i] = cloneJSONValue(child)
		}
		return result
	default:
		return value
	}
}
