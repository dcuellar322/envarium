package cli

import "envarium.dev/envarium/internal/transport/contract/generated"

// This is the native adapter's presentation contract, not a second API model.
// Keep every tray/notification input, but do not re-encode full runtime service
// inventories, repository paths, workspace recipes or operation details the
// native shell never reads. The ordinary CLI and HTTP views remain complete.
type desktopPresentation struct {
	System            generated.SystemInfo           `json:"system"`
	Host              *generated.HostObservation     `json:"host,omitempty"`
	Projects          []desktopPresentationProject   `json:"projects"`
	Workspaces        []desktopPresentationWorkspace `json:"workspaces"`
	Operations        []desktopPresentationOperation `json:"operations"`
	Notifications     []desktopPresentationNotice    `json:"diagnosticNotifications"`
	PortConflictCount int                            `json:"portConflictCount"`
	Warnings          []string                       `json:"warnings"`
}
type desktopPresentationProject struct {
	Project  desktopPresentationIdentity `json:"project"`
	Runtime  *desktopPresentationState   `json:"runtime,omitempty"`
	Health   *desktopPresentationHealth  `json:"health,omitempty"`
	Warnings []string                    `json:"warnings"`
}
type desktopPresentationIdentity struct {
	ID   string `json:"id"`
	Name string `json:"displayName"`
}
type desktopPresentationState struct {
	State string `json:"state"`
}
type desktopPresentationHealth struct {
	Status string `json:"status"`
}
type desktopPresentationWorkspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type desktopPresentationOperation struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	State string `json:"state"`
}
type desktopPresentationNotice struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Detail      string `json:"detail"`
	Occurrences int    `json:"occurrences"`
}

func nativePresentation(source desktopSnapshot) desktopPresentation {
	result := desktopPresentation{System: source.System, Host: source.Host, Projects: []desktopPresentationProject{}, Workspaces: []desktopPresentationWorkspace{}, Operations: []desktopPresentationOperation{}, Notifications: []desktopPresentationNotice{}, PortConflictCount: source.PortConflictCount, Warnings: source.Warnings}
	for _, item := range source.Projects {
		project := desktopPresentationProject{Project: desktopPresentationIdentity{item.Project.Id, item.Project.DisplayName}, Warnings: item.Warnings}
		if item.Runtime != nil {
			project.Runtime = &desktopPresentationState{string(item.Runtime.State)}
		}
		if item.Health != nil {
			project.Health = &desktopPresentationHealth{string(item.Health.Status)}
		}
		result.Projects = append(result.Projects, project)
	}
	for _, item := range source.Workspaces {
		result.Workspaces = append(result.Workspaces, desktopPresentationWorkspace{item.Id, item.Name})
	}
	for _, item := range source.Operations {
		result.Operations = append(result.Operations, desktopPresentationOperation{item.Id, item.Kind, string(item.State)})
	}
	for _, item := range source.Notifications {
		result.Notifications = append(result.Notifications, desktopPresentationNotice{item.Id, item.Title, item.Detail, item.Occurrences})
	}
	return result
}
