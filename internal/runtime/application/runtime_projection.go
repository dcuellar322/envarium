package application

import (
	"encoding/json"
	"maps"
	"slices"
	"sync"

	catalog "envarium.dev/envarium/internal/catalog/domain"
	"envarium.dev/envarium/internal/runtime/domain"
)

// Current trust, source bytes and environment overlays are checked separately
// on every resolution. This cache retains only immutable typed projections.
type runtimeProjections struct {
	mu      sync.Mutex
	entries map[string]runtimeProjection
	bytes   int
}
type runtimeProjection struct {
	value  domain.ProjectRuntime
	weight int
}

func (c *runtimeProjections) load(project catalog.Project, hash string) (domain.ProjectRuntime, bool) {
	c.mu.Lock()
	entry, ok := c.entries[project.ID]
	c.mu.Unlock()
	if !ok || entry.value.ManifestHash != hash || entry.value.Root != project.PrimaryLocation || entry.value.ProjectSlug != project.Slug {
		return domain.ProjectRuntime{}, false
	}
	return cloneRuntimeProjection(entry.value), true
}

func (c *runtimeProjections) store(value domain.ProjectRuntime) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return
	}
	weight := 1024 + 8*len(encoded)
	const maximumBytes = 4 << 20
	if weight > maximumBytes {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[string]runtimeProjection)
	}
	if old, ok := c.entries[value.ProjectID]; ok {
		c.bytes -= old.weight
	}
	if len(c.entries) >= 128 || c.bytes+weight > maximumBytes {
		clear(c.entries)
		c.bytes = 0
	}
	c.entries[value.ProjectID] = runtimeProjection{value: cloneRuntimeProjection(value), weight: weight}
	c.bytes += weight
}

func cloneRuntimeProjection(value domain.ProjectRuntime) domain.ProjectRuntime {
	result := value
	result.Ports = maps.Clone(value.Ports)
	if value.Compose != nil {
		compose := *value.Compose
		compose.Files = slices.Clone(compose.Files)
		compose.Profiles = slices.Clone(compose.Profiles)
		compose.PortOverrides = maps.Clone(compose.PortOverrides)
		result.Compose = &compose
	}
	if value.Process != nil {
		process := *value.Process
		process.Environment = maps.Clone(process.Environment)
		process.Secrets = maps.Clone(process.Secrets)
		process.Processes = slices.Clone(process.Processes)
		for i := range process.Processes {
			p := &process.Processes[i]
			p.Command = slices.Clone(p.Command)
			p.Environment = maps.Clone(p.Environment)
			p.Secrets = maps.Clone(p.Secrets)
		}
		result.Process = &process
	}
	result.Services = slices.Clone(value.Services)
	for i := range result.Services {
		service := &result.Services[i]
		service.Dependencies = slices.Clone(service.Dependencies)
		service.HostPorts = slices.Clone(service.HostPorts)
		service.HealthChecks = slices.Clone(service.HealthChecks)
		for j := range service.HealthChecks {
			check := &service.HealthChecks[j]
			check.Command = slices.Clone(check.Command)
			check.Members = slices.Clone(check.Members)
		}
	}
	return result
}
