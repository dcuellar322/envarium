export interface Integration {
  slug: string
  name: string
  summary: string
  description: string
  setup: string[]
  docsHref: string
}

export const integrations: Integration[] = [
  {
    slug: 'docker-compose',
    name: 'Docker Compose',
    summary: 'Start Compose projects and inspect service status and logs.',
    description: 'Use the Docker context and Compose files already configured on your machine. Envarium groups their services under the project and shows status, logs, and health.',
    setup: ['Install Docker with the Compose plugin and check that the engine is reachable.', 'Add a repository with a Compose file.', 'Review discovered services, commands, ports, and health checks before approving the project.'],
    docsHref: '/docs/architecture/docker-compose-runtime/',
  },
  {
    slug: 'native-processes',
    name: 'Native processes',
    summary: 'Start local development commands and follow their output alongside Compose services.',
    description: 'Run start commands in the project directory with the arguments and environment references you approve. Read stdout and stderr in the project log view.',
    setup: ['Install the tools your project needs, such as uv, npm, pnpm, or Make.', 'Add the repository and review its start command, working directory, ports, and health checks.', 'Approve the configuration, then start the project and follow its logs.'],
    docsHref: '/docs/architecture/native-process-runtime/',
  },
  {
    slug: 'codex',
    name: 'Codex',
    summary: 'Let Codex inspect project state and request approved operations.',
    description: 'Connect Codex to Envarium to read project status, logs, and health. Start with read-only access, then choose whether to allow service operations and reviewed project actions.',
    setup: ['Open the repository where you want project-local Codex configuration.', 'Run the installer below and review the files it reports changing.', 'Reload the client’s MCP configuration and ask it to list Envarium projects.'],
    docsHref: '/docs/mcp/',
  },
  {
    slug: 'claude-code',
    name: 'Claude Code',
    summary: 'Connect Claude Code to project status, logs, and operations you permit.',
    description: 'Connect Claude Code to Envarium over MCP. Choose the same read-only or operation permissions available to other clients; agent changes are permission-checked and audited.',
    setup: ['Open the repository where you want project-local Claude Code configuration.', 'Run the installer below and review the files it reports changing.', 'Reload the client’s MCP configuration and ask it to list Envarium projects.'],
    docsHref: '/docs/mcp/',
  },
  {
    slug: 'vscode',
    name: 'VS Code',
    summary: 'Open a project or worktree in VS Code from its reviewed editor action.',
    description: 'Open a project directory in VS Code from Envarium. macOS launches the installed Visual Studio Code app; Linux uses the code command on your PATH.',
    setup: ['Install Visual Studio Code. On Linux, ensure the code command is on your PATH.', 'Review and approve the project’s editor action and working directory.', 'Run Open VS Code from the project actions. For a worktree, use its own project entry.'],
    docsHref: '/docs/architecture/developer-workflows/',
  },
  {
    slug: 'mcp',
    name: 'Generic MCP clients',
    summary: 'Connect a local stdio MCP client to project tools with explicit permissions.',
    description: 'Configure a client that supports stdio MCP servers to launch Envarium. The client can inspect projects and request only the operations allowed by its permission profile.',
    setup: ['Find the absolute path to your installed Envarium binary.', 'Add a stdio server entry to your client using the example below.', 'Reload the client’s MCP configuration and request the Envarium project list.'],
    docsHref: '/docs/mcp/',
  },
]
