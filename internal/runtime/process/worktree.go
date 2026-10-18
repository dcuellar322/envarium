package process

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"envarium.dev/envarium/internal/runtime/domain"
)

func validateWorktreeServices(project domain.ProjectRuntime, services []servicePlan) error {
	if project.EnvironmentID == "" {
		return nil
	}
	root, err := filepath.EvalSymlinks(project.Root)
	if err != nil {
		return fmt.Errorf("worktree root unavailable: %w", err)
	}
	for _, service := range services {
		command := previewCommand(project.Root, service.definition)
		directory, err := filepath.EvalSymlinks(command.WorkingDirectory)
		if err != nil {
			return fmt.Errorf("worktree service %s working directory unavailable: %w", service.service.ID, err)
		}
		relative, err := filepath.Rel(root, directory)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("worktree service %s working directory must stay inside the selected checkout", service.service.ID)
		}
		for id, port := range project.Ports {
			if port.Service == service.service.ID && !port.Allocated {
				return fmt.Errorf("worktree port %s has no exact allocation; reconcile worktrees before starting", id)
			}
			if port.Service == service.service.ID && !slices.Contains(service.definition.PortBindings, id) {
				return fmt.Errorf("native worktree service %s must reference ${ports.%s} in its command or effective environment to consume its allocated port", service.service.ID, id)
			}
		}
	}
	return nil
}

// Never adopt or terminate an unrelated listener when a leased port is occupied.
func (d *Driver) checkWorktreePorts(ctx context.Context, project domain.ProjectRuntime, service servicePlan) error {
	if project.EnvironmentID == "" {
		return nil
	}
	for id, port := range project.Ports {
		if port.Service != service.service.ID {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port.Host))
		if port.Protocol == "udp" {
			listener, err := net.ListenPacket("udp", address)
			if err != nil {
				return fmt.Errorf("worktree port %s (%d/udp) is unavailable: %w", id, port.Host, err)
			}
			_ = listener.Close()
		} else {
			listener, err := net.Listen("tcp", address)
			if err != nil {
				return fmt.Errorf("worktree port %s (%d/tcp) is unavailable: %w", id, port.Host, err)
			}
			_ = listener.Close()
		}
	}
	return nil
}
