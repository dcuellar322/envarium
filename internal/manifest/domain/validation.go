package domain

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// Validate checks invariants that JSON Schema cannot express clearly.
func (m Manifest) Validate() error {
	var problems []error
	if m.SchemaVersion != SchemaVersion {
		problems = append(problems, fmt.Errorf("schemaVersion must be %q", SchemaVersion))
	}
	if m.Kind != KindProject {
		problems = append(problems, fmt.Errorf("kind must be %q", KindProject))
	}
	if m.Metadata.ID == "" || m.Metadata.Name == "" {
		problems = append(problems, errors.New("metadata.id and metadata.name are required"))
	}
	if len(m.Metadata.ID) > 63 || !manifestID.MatchString(m.Metadata.ID) {
		problems = append(problems, errors.New("metadata.id must be a lowercase slug no longer than 63 characters"))
	}
	if m.Repository.Root == "" {
		problems = append(problems, errors.New("repository.root is required"))
	}
	problems = append(problems, uniqueIDs("service", serviceIDs(m.Services))...)
	problems = append(problems, uniqueIDs("port", portIDs(m.Ports))...)
	problems = append(problems, uniqueIDs("action", actionIDs(m.Actions))...)
	problems = append(problems, uniqueIDs("endpoint", endpointIDs(m.Endpoints))...)
	problems = append(problems, validateServices(m.Services)...)
	problems = append(problems, validateRuntime(m.Runtime)...)
	problems = append(problems, validateProcessBindings(m.Runtime, m.Services)...)
	problems = append(problems, validatePorts(m.Ports, m.Services)...)
	problems = append(problems, validateEndpoints(m.Endpoints)...)
	problems = append(problems, validateActions(m.Actions)...)
	return errors.Join(problems...)
}

func validateServices(services []Service) []error {
	serviceSet := make(map[string]struct{}, len(services))
	for _, service := range services {
		serviceSet[service.ID] = struct{}{}
	}
	var problems []error
	for _, service := range services {
		sources := 0
		if service.Source.ComposeService != "" {
			sources++
		}
		if service.Source.Process != "" {
			sources++
		}
		if sources != 1 {
			problems = append(problems, fmt.Errorf("service %q must declare exactly one source", service.ID))
		}
		for _, dependency := range service.Dependencies {
			if _, ok := serviceSet[dependency]; !ok {
				problems = append(problems, fmt.Errorf("service %q references unknown dependency %q", service.ID, dependency))
			}
		}
		problems = append(problems, validateServiceHealth(service)...)
	}
	return problems
}

func validateServiceHealth(service Service) []error {
	var problems []error
	checkIDs := make(map[string]struct{}, len(service.HealthChecks))
	for index, check := range service.HealthChecks {
		checkID := healthCheckID(check, index)
		if _, exists := checkIDs[checkID]; exists {
			problems = append(problems, fmt.Errorf("service %q has duplicate health check ID %q", service.ID, checkID))
		}
		checkIDs[checkID] = struct{}{}
		problems = append(problems, validateHealthCheck(service.ID, checkID, check)...)
	}
	for index, check := range service.HealthChecks {
		if check.Type != "composite" {
			continue
		}
		checkID := healthCheckID(check, index)
		for _, member := range check.Members {
			if member == checkID {
				problems = append(problems, fmt.Errorf("service %q composite health check %q cannot reference itself", service.ID, checkID))
			} else if _, exists := checkIDs[member]; !exists {
				problems = append(problems, fmt.Errorf("service %q composite health check %q references unknown member %q", service.ID, checkID, member))
			}
		}
	}
	return append(problems, validateCompositeCycles(service)...)
}

func validateHealthCheck(serviceID, checkID string, check HealthCheck) []error {
	var problems []error
	if check.Severity != "" && check.Severity != "info" && check.Severity != "warning" && check.Severity != "critical" {
		problems = append(problems, fmt.Errorf("service %q health check %q has invalid severity %q", serviceID, checkID, check.Severity))
	}
	switch check.Type {
	case "http":
		if check.URL == "" {
			problems = append(problems, fmt.Errorf("service %q HTTP health check requires a URL", serviceID))
		}
	case "tcp":
		if check.Address == "" {
			problems = append(problems, fmt.Errorf("service %q TCP health check requires an address", serviceID))
		}
	case "process", "docker":
	case "command":
		if len(check.Command) == 0 {
			problems = append(problems, fmt.Errorf("service %q command health check requires an argument array", serviceID))
		}
	case "composite":
		if len(check.Members) == 0 {
			problems = append(problems, fmt.Errorf("service %q composite health check requires members", serviceID))
		}
		if check.Mode != "" && check.Mode != "all" && check.Mode != "any" {
			problems = append(problems, fmt.Errorf("service %q composite health check mode must be all or any", serviceID))
		}
	default:
		problems = append(problems, fmt.Errorf("service %q has unknown health check type %q", serviceID, check.Type))
	}
	return problems
}

func healthCheckID(check HealthCheck, index int) string {
	if check.ID != "" {
		return check.ID
	}
	return fmt.Sprintf("%s-%d", check.Type, index+1)
}

func validateCompositeCycles(service Service) []error {
	graph := make(map[string][]string)
	for index, check := range service.HealthChecks {
		if check.Type != "composite" {
			continue
		}
		id := healthCheckID(check, index)
		graph[id] = append([]string(nil), check.Members...)
	}
	state := make(map[string]uint8, len(graph))
	var problems []error
	var visit func(string)
	visit = func(id string) {
		if state[id] == 1 {
			problems = append(problems, fmt.Errorf("service %q composite health checks contain a cycle at %q", service.ID, id))
			return
		}
		if state[id] == 2 {
			return
		}
		state[id] = 1
		for _, member := range graph[id] {
			if _, composite := graph[member]; composite {
				visit(member)
			}
		}
		state[id] = 2
	}
	for id := range graph {
		visit(id)
	}
	return problems
}

func validatePorts(ports []Port, services []Service) []error {
	var problems []error
	for _, port := range ports {
		if port.Host < 1 || port.Host > 65535 || port.Target < 1 || port.Target > 65535 {
			problems = append(problems, fmt.Errorf("port %q must use values between 1 and 65535", port.ID))
		}
		if port.Service != "" && !slices.Contains(serviceIDs(services), port.Service) {
			problems = append(problems, fmt.Errorf("port %q references unknown service %q", port.ID, port.Service))
		}
		if port.Protocol != "tcp" && port.Protocol != "udp" {
			problems = append(problems, fmt.Errorf("port %q protocol must be tcp or udp", port.ID))
		}
	}
	return problems
}

func validateEndpoints(endpoints []Endpoint) []error {
	var problems []error
	for _, endpoint := range endpoints {
		if !strings.Contains(endpoint.URL, "${ports.") {
			if parsed, err := url.ParseRequestURI(endpoint.URL); err != nil || parsed.Scheme == "" {
				problems = append(problems, fmt.Errorf("endpoint %q has invalid URL", endpoint.ID))
			}
		}
	}
	return problems
}

func validateActions(actions []Action) []error {
	var problems []error
	for _, action := range actions {
		if (action.Type == "command" || action.Type == "command.run" || action.Type == "tests.run" || action.Type == "migration.run") && len(action.Command) == 0 {
			problems = append(problems, fmt.Errorf("command action %q requires an argument array", action.ID))
		}
		if action.Shell && len(action.Command) != 1 {
			problems = append(problems, fmt.Errorf("shell action %q requires exactly one command string", action.ID))
		}
		if action.Type == "agent.start" && action.Provider == "" {
			problems = append(problems, fmt.Errorf("agent action %q requires a provider", action.ID))
		}
		if action.Type == "browser.open" && action.Target == "" {
			problems = append(problems, fmt.Errorf("browser action %q requires a target", action.ID))
		}
		problems = append(problems, validateEnvironment("action "+action.ID, action.Environment, nil)...)
	}
	return problems
}

func serviceIDs(values []Service) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func portIDs(values []Port) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func actionIDs(values []Action) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func endpointIDs(values []Endpoint) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func uniqueIDs(kind string, values []string) []error {
	seen := make(map[string]struct{}, len(values))
	var problems []error
	for _, value := range values {
		if value == "" {
			problems = append(problems, fmt.Errorf("%s id is required", kind))
			continue
		}
		if _, ok := seen[value]; ok {
			problems = append(problems, fmt.Errorf("duplicate %s id %q", kind, value))
		}
		seen[value] = struct{}{}
	}
	return problems
}
