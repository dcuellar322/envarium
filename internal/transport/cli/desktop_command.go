package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

type desktopSnapshot struct {
	System            generated.SystemInfo               `json:"system"`
	Host              *generated.HostObservation         `json:"host,omitempty"`
	Projects          []desktopProjectSnapshot           `json:"projects"`
	Workspaces        []generated.Workspace              `json:"workspaces"`
	Operations        []generated.Operation              `json:"operations"`
	Notifications     []generated.DiagnosticNotification `json:"diagnosticNotifications"`
	PortConflictCount int                                `json:"portConflictCount"`
	Warnings          []string                           `json:"warnings"`
}

type desktopProjectSnapshot struct {
	Project  generated.Project             `json:"project"`
	Runtime  *generated.RuntimeObservation `json:"runtime,omitempty"`
	Health   *generated.ProjectHealth      `json:"health,omitempty"`
	Warnings []string                      `json:"warnings"`
}

func newDesktopCommand(options *rootOptions) *cobra.Command {
	command := &cobra.Command{Use: "desktop", Short: "Support the thin native desktop adapter"}
	command.AddCommand(newDesktopSnapshotCommand(options))
	return command
}

func newDesktopSnapshotCommand(options *rootOptions) *cobra.Command {
	var stream bool
	command := &cobra.Command{
		Use:   "snapshot",
		Short: "Read one bounded tray and notification snapshot",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if stream && (!options.json || !options.compact || options.jsonl) {
				return usageError("DESKTOP_STREAM_FORMAT", "native snapshot streaming requires --json --compact")
			}
			client, err := daemonClient(command.Context(), options)
			if err != nil {
				return err
			}
			if stream {
				return streamDesktopSnapshots(command.Context(), command.InOrStdin(), options.stdout, func(ctx context.Context) (desktopPresentation, error) {
					snapshot, err := readDesktopSnapshot(ctx, client)
					return nativePresentation(snapshot), err
				})
			}
			snapshot, err := readDesktopSnapshot(command.Context(), client)
			if err != nil {
				return err
			}

			return writeResult(options, "desktop.snapshot", nativePresentation(snapshot), func(writer io.Writer) error {
				_, err := fmt.Fprintf(
					writer,
					"daemon=%s version=%s projects=%d workspaces=%d conflicts=%d warnings=%d\n",
					snapshot.System.Status,
					snapshot.System.Version,
					len(snapshot.Projects),
					len(snapshot.Workspaces),
					snapshot.PortConflictCount,
					len(snapshot.Warnings),
				)
				return err
			})
		},
	}
	command.Flags().BoolVar(&stream, "stream", false, "serve bounded native snapshot requests on standard input")
	_ = command.Flags().MarkHidden("stream")
	return command
}
