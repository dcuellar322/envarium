package application

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"envarium.dev/envarium/internal/runtime/domain"
)

// projectNativePorts expands only explicit decimal port references. It never
// guesses which numeric arguments or arbitrary environment variables are ports.
func projectNativePorts(project *domain.ProjectRuntime) error {
	if project.Process == nil {
		return nil
	}
	config := project.Process
	for i := range config.Processes {
		process := &config.Processes[i]
		effective := maps.Clone(config.Environment)
		if effective == nil {
			effective = map[string]string{}
		}
		for key := range config.Secrets {
			delete(effective, key)
		}
		maps.Copy(effective, process.Environment)
		for key := range process.Secrets {
			delete(effective, key)
		}
		used := map[string]bool{}
		for _, value := range effective {
			if _, err := expandNativePorts(value, project.Ports, used); err != nil {
				return err
			}
		}
		for j, argument := range process.Command {
			if j == 0 && !process.Shell && strings.Contains(argument, "${ports.") {
				return fmt.Errorf("process %s cannot select its executable using a port reference", process.ID)
			}
			value, err := expandNativePorts(argument, project.Ports, used)
			if err != nil {
				return err
			}
			process.Command[j] = value
		}
		for key, value := range process.Environment {
			expanded, err := expandNativePorts(value, project.Ports, nil)
			if err != nil {
				return err
			}
			process.Environment[key] = expanded
		}
		process.PortBindings = slices.Sorted(maps.Keys(used))
	}
	for key, value := range config.Environment {
		expanded, err := expandNativePorts(value, project.Ports, nil)
		if err != nil {
			return err
		}
		config.Environment[key] = expanded
	}
	return nil
}

func expandNativePorts(value string, ports map[string]domain.PortDeclaration, used map[string]bool) (string, error) {
	var result strings.Builder
	for {
		prefix, rest, found := strings.Cut(value, "${ports.")
		result.WriteString(prefix)
		if !found {
			return result.String(), nil
		}
		id, tail, closed := strings.Cut(rest, "}")
		port, exists := ports[id]
		if !closed || !exists || port.Host < 1 || port.Host > 65535 {
			return "", fmt.Errorf("native process references an unknown or invalid port %q", id)
		}
		result.WriteString(strconv.Itoa(port.Host))
		if used != nil {
			used[id] = true
		}
		value = tail
	}
}
