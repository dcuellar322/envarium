package cli

import (
	"context"

	"envarium.dev/envarium/internal/transport/contract/generated"
	"envarium.dev/envarium/internal/transport/httpclient"
)

func readDesktopSnapshot(ctx context.Context, client *httpclient.Client) (desktopSnapshot, error) {
	system, err := client.System(ctx)
	if err != nil {
		return desktopSnapshot{}, err
	}
	projects, err := client.Projects(ctx)
	if err != nil {
		return desktopSnapshot{}, err
	}

	snapshot := desktopSnapshot{
		System: system, Projects: make([]desktopProjectSnapshot, 0, len(projects)),
		Workspaces: []generated.Workspace{}, Operations: []generated.Operation{},
		Notifications: []generated.DiagnosticNotification{}, Warnings: []string{},
	}
	if host, hostErr := client.Host(ctx); hostErr == nil {
		snapshot.Host = &host
	} else {
		snapshot.Warnings = append(snapshot.Warnings, "host observation unavailable")
	}
	if workspaces, workspaceErr := client.Workspaces(ctx); workspaceErr == nil {
		snapshot.Workspaces = workspaces
	} else {
		snapshot.Warnings = append(snapshot.Warnings, "workspace snapshot unavailable")
	}
	if operations, operationErr := client.Operations(ctx, "", 25); operationErr == nil {
		snapshot.Operations = operations
	} else {
		snapshot.Warnings = append(snapshot.Warnings, "operation snapshot unavailable")
	}
	if notifications, notificationErr := client.DiagnosticNotifications(ctx, "", false); notificationErr == nil {
		snapshot.Notifications = notifications
	} else {
		snapshot.Warnings = append(snapshot.Warnings, "diagnostic notification snapshot unavailable")
	}
	if ports, portErr := client.PortRegistry(ctx); portErr == nil {
		snapshot.PortConflictCount = len(ports.Conflicts)
	} else {
		snapshot.Warnings = append(snapshot.Warnings, "port conflict snapshot unavailable")
	}

	snapshot.Projects, err = desktopProjects(ctx, client, projects)
	if err != nil {
		return desktopSnapshot{}, err
	}

	return snapshot, nil
}
