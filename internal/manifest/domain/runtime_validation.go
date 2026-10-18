package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func validateRuntime(runtime Runtime) []error {
	var problems []error
	if runtime.Driver == "compose" && (runtime.Compose == nil || len(runtime.Compose.Files) == 0) {
		problems = append(problems, errors.New("compose runtime requires at least one Compose file"))
	}
	if runtime.Driver == "process" && (runtime.Process == nil || len(runtime.Process.Processes) == 0) {
		problems = append(problems, errors.New("process runtime requires at least one process definition"))
	}
	if runtime.Compose != nil && runtime.Process != nil {
		problems = append(problems, errors.New("runtime cannot declare both compose and process configuration"))
	}
	if runtime.Compose != nil {
		seen := make(map[string]struct{}, len(runtime.Compose.Profiles))
		for _, profile := range runtime.Compose.Profiles {
			if strings.TrimSpace(profile) == "" {
				problems = append(problems, errors.New("compose profiles cannot contain an empty name"))
				continue
			}
			if _, duplicate := seen[profile]; duplicate {
				problems = append(problems, fmt.Errorf("compose profile %q is declared more than once", profile))
			}
			seen[profile] = struct{}{}
		}
	}
	if runtime.Process != nil {
		problems = append(problems, validateProcesses(*runtime.Process)...)
	}
	return problems
}

func validateProcesses(config ProcessConfig) []error {
	var problems []error
	processIDs := make([]string, 0, len(config.Processes))
	for _, process := range config.Processes {
		processIDs = append(processIDs, process.ID)
		if len(process.Command) == 0 || strings.TrimSpace(process.Command[0]) == "" {
			problems = append(problems, fmt.Errorf("process %q requires a command argument array", process.ID))
		}
		if !process.Shell && usesShellExecutable(process.Command) {
			problems = append(problems, fmt.Errorf("process %q uses shell syntax without shell: true", process.ID))
		}
		if process.Shell && len(process.Command) != 1 {
			problems = append(problems, fmt.Errorf("shell process %q must provide exactly one command string", process.ID))
		}
		if process.Restart.Mode != "" && process.Restart.Mode != "never" && process.Restart.Mode != "on-failure" {
			problems = append(problems, fmt.Errorf("process %q has unknown restart mode %q", process.ID, process.Restart.Mode))
		}
		problems = append(problems, validateEnvironment("process "+process.ID, process.Environment, process.Secrets)...)
	}
	problems = append(problems, uniqueIDs("process", processIDs)...)
	problems = append(problems, validateEnvironment("process runtime", config.Environment, config.Secrets)...)
	return problems
}

func validateProcessBindings(runtime Runtime, services []Service) []error {
	var problems []error
	for _, service := range services {
		if runtime.Driver == "process" && service.Source.Process == "" {
			problems = append(problems, fmt.Errorf("process runtime service %q must reference a process definition", service.ID))
		}
		if runtime.Driver == "compose" && service.Source.ComposeService == "" {
			problems = append(problems, fmt.Errorf("compose runtime service %q must reference a Compose service", service.ID))
		}
	}
	if runtime.Process == nil {
		return problems
	}
	definitions := make(map[string]struct{}, len(runtime.Process.Processes))
	for _, process := range runtime.Process.Processes {
		definitions[process.ID] = struct{}{}
	}
	for _, service := range services {
		if service.Source.Process == "" {
			continue
		}
		if _, ok := definitions[service.Source.Process]; !ok {
			problems = append(problems, fmt.Errorf("service %q references unknown process %q", service.ID, service.Source.Process))
		}
	}
	if cycle := dependencyCycle(services); len(cycle) > 0 {
		problems = append(problems, fmt.Errorf("service dependency cycle: %s", strings.Join(cycle, " -> ")))
	}
	return problems
}

func dependencyCycle(services []Service) []string {
	dependencies := make(map[string][]string, len(services))
	for _, service := range services {
		dependencies[service.ID] = service.Dependencies
	}
	state := make(map[string]uint8, len(services))
	var stack []string
	var visit func(string) []string
	visit = func(id string) []string {
		if state[id] == 1 {
			for index, item := range stack {
				if item == id {
					return append(append([]string(nil), stack[index:]...), id)
				}
			}
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		stack = append(stack, id)
		for _, dependency := range dependencies[id] {
			if cycle := visit(dependency); len(cycle) > 0 {
				return cycle
			}
		}
		stack = stack[:len(stack)-1]
		state[id] = 2
		return nil
	}
	for _, service := range services {
		if cycle := visit(service.ID); len(cycle) > 0 {
			return cycle
		}
	}
	return nil
}

func usesShellExecutable(command []string) bool {
	if len(command) == 0 || strings.ContainsAny(command[0], " \t\r\n|&;<>()$`") {
		return true
	}
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(command[0]), ".exe"))
	return slices.Contains([]string{"sh", "bash", "zsh", "fish", "cmd", "powershell", "pwsh"}, base)
}

var environmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateEnvironment(owner string, values map[string]string, secrets map[string]SecretRef) []error {
	var problems []error
	for key := range values {
		if !environmentName.MatchString(key) {
			problems = append(problems, fmt.Errorf("%s has invalid environment name %q", owner, key))
		}
		if _, exists := secrets[key]; exists {
			problems = append(problems, fmt.Errorf("%s environment %q cannot be both a value and secret reference", owner, key))
		}
	}
	for key, reference := range secrets {
		if !environmentName.MatchString(key) {
			problems = append(problems, fmt.Errorf("%s has invalid secret environment name %q", owner, key))
		}
		if reference.Provider != "keychain" || strings.TrimSpace(reference.Key) == "" {
			problems = append(problems, fmt.Errorf("%s secret %q requires a keychain key", owner, key))
		}
	}
	return problems
}
