package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	agents "envarium.dev/envarium/internal/agents/application"
	"envarium.dev/envarium/internal/agents/providers"
	"envarium.dev/envarium/internal/agents/providers/claude"
	"envarium.dev/envarium/internal/agents/providers/codex"
)

func newAgentCommand(options *rootOptions) *cobra.Command {
	command := &cobra.Command{Use: "agent", Short: "Install agent-native Envarium integration"}
	command.AddCommand(newAgentInstallCommand(options))
	return command
}

func newAgentInstallCommand(options *rootOptions) *cobra.Command {
	var scopeName, root, profileName, agentID, executable string
	var projectIDs []string
	command := &cobra.Command{
		Use:       "install [codex|claude]",
		Short:     "Install MCP configuration, shared skill, and repository guidance",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"codex", "claude"},
		RunE: func(_ *cobra.Command, args []string) error {
			provider := args[0]
			if provider != "codex" && provider != "claude" {
				return usageError("AGENT_PROVIDER_UNSUPPORTED", "supported providers: codex, claude")
			}
			profile, err := agents.ParseProfile(profileName)
			if err != nil {
				return usageError("AGENT_PROFILE_INVALID", err.Error())
			}
			scope := providers.InstallScope(scopeName)
			if scope != providers.ScopeProject && scope != providers.ScopeUser {
				return usageError("AGENT_SCOPE_INVALID", "supported install scopes: project, user")
			}
			resolvedRoot, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			if executable == "" {
				executable, err = os.Executable()
				if err != nil {
					return err
				}
			}
			executable, err = filepath.Abs(executable)
			if err != nil {
				return err
			}
			dataDir, err := filepath.Abs(options.dataDir)
			if err != nil {
				return err
			}
			request := providers.InstallRequest{Scope: scope, Root: resolvedRoot, Home: home, Executable: executable, DataDir: dataDir, Profile: profile, AgentID: agentID, ProjectIDs: projectIDs}
			var paths []string
			switch provider {
			case "codex":
				paths, err = codex.Install(request)
			case "claude":
				paths, err = claude.Install(request)
			}
			if err != nil {
				return err
			}
			result := map[string]any{"provider": provider, "scope": scope, "profile": profile, "files": paths}
			return writeResult(options, "agent.install", result, func(writer io.Writer) error {
				_, err := fmt.Fprintf(writer, "Installed Envarium for %s (%s, %s)\n", provider, scope, profile)
				for _, path := range paths {
					if _, writeErr := fmt.Fprintf(writer, "  %s\n", path); err == nil {
						err = writeErr
					}
				}
				return err
			})
		},
	}
	command.Flags().StringVar(&scopeName, "scope", string(providers.ScopeProject), "installation scope: project or user")
	command.Flags().StringVar(&root, "root", ".", "project root for project-scoped files")
	command.Flags().StringVar(&profileName, "profile", string(agents.ProfileObserve), "permission profile: observe, develop, maintain, or admin")
	command.Flags().StringVar(&agentID, "agent-id", "envarium", "agent identity recorded in audit events")
	command.Flags().StringSliceVar(&projectIDs, "project", nil, "restrict access to a project ID (repeatable)")
	command.Flags().StringVar(&executable, "command", "", "absolute Envarium executable path (defaults to this executable)")
	return command
}
