import type { BrandIconName } from '../../../../design/brand/envarium-v1/icons'

export interface Feature {
  slug: string
  title: string
  summary: string
  description: string
  proof: string[]
  docsHref: string
  icon: BrandIconName
}

export const features: Feature[] = [
  {
    slug: 'projects',
    title: 'Project command center',
    summary: 'See the repository, services, health, Git, endpoints, actions, and resources as one project.',
    description:
      'Keep each repository’s start commands, services, URLs, and health checks together. The dashboard shows which projects are running and which need attention.',
    proof: ['Review detected commands and ports before running a project', 'Save project configuration in a portable manifest', 'Distinguish managed services from externally started processes'],
    docsHref: '/docs/architecture/project-onboarding/',
    icon: 'projects',
  },
  {
    slug: 'runtimes',
    title: 'Compose and native runtimes',
    summary: 'Operate Docker Compose, uv, npm, Make, and other native processes from the same lifecycle.',
    description:
      'Use your existing Compose configuration or local start commands. Preview service changes before starting or stopping a project, then follow progress and errors. Compose projects also support rebuilding containers.',
    proof: ['Preview start and stop operations, or rebuild Compose services', 'Run commands with explicit arguments and working directories', 'Inspect partially running projects and runtime failures'],
    docsHref: '/docs/architecture/docker-compose-runtime/',
    icon: 'services',
  },
  {
    slug: 'logs-health-resources',
    title: 'Logs, health, and resources',
    summary: 'Read service logs, check readiness, and see which projects use CPU and memory.',
    description:
      'A running process may still be unhealthy. Inspect service health alongside logs, CPU, memory, and storage usage to understand why a project is not ready.',
    proof: ['Follow live logs and inspect retained output', 'Wait for health checks after starting services', 'Set log retention limits and inspect resource usage'],
    docsHref: '/docs/architecture/observability/',
    icon: 'resources',
  },
  {
    slug: 'ports-git-actions',
    title: 'Ports, Git, and trusted actions',
    summary: 'Catch collisions before startup and run the project actions you have already reviewed.',
    description:
      'See which process owns a port and reserve ports for stopped projects. Check the current branch and working-tree changes beside service status, then review project actions before running them.',
    proof: ['Reserve ports even while a project is stopped', 'See branch and working-tree changes beside runtime status', 'Review the risk of an action before running it'],
    docsHref: '/docs/architecture/developer-workflows/',
    icon: 'ports',
  },
  {
    slug: 'workspaces-worktrees',
    title: 'Workspaces and worktrees',
    summary: 'Start related projects in dependency order and run isolated feature worktrees side by side.',
    description:
      'Start a group of projects in dependency order and wait for their health checks. Give separate worktrees their own runtime identity and ports so feature branches can run alongside each other.',
    proof: ['Declare which projects must start first', 'Inspect failures without losing the status of other projects', 'Assign separate runtime identities and ports to worktrees'],
    docsHref: '/docs/architecture/workspaces/',
    icon: 'workspaces',
  },
  {
    slug: 'agents-mcp',
    title: 'Codex, Claude Code, and MCP',
    summary: 'Let coding agents inspect project status and request operations within permissions you choose.',
    description:
      'Connect Codex, Claude Code, or another MCP client to project status, logs, and operations. Choose a permission profile for each client. Agent mutations are checked and audited, and AI-generated configuration proposals require human approval.',
    proof: ['Start with read-only project access', 'Choose which project operations an agent may request', 'Review AI-generated configuration before accepting it'],
    docsHref: '/docs/mcp/',
    icon: 'agents',
  },
  {
    slug: 'terminals',
    title: 'Embedded and external terminals',
    summary: 'Run interactive commands in a project terminal or open the terminal app you already use.',
    description:
      'Open a terminal in the project directory, use interactive command-line tools, and reconnect to a session. You can also launch your preferred external terminal. Windows terminal support is coming soon.',
    proof: ['Use interactive terminal programs on macOS and Linux', 'Resize and reconnect to project terminal sessions', 'Open a project in an external terminal'],
    docsHref: '/docs/architecture/terminal-sessions/',
    icon: 'terminal',
  },
  {
    slug: 'plugins',
    title: 'Capability-scoped plugins',
    summary: 'Add integrations through a documented protocol and review the capabilities each plugin requests.',
    description:
      'Review a plugin’s executable fingerprint and requested capabilities before granting trust. Plugins run as separate supervised processes under your user account. Capability checks govern access to Envarium; they do not provide an operating-system sandbox.',
    proof: ['Run plugins in separate supervised processes', 'Review the capabilities a plugin requests', 'Build against the documented plugin protocol'],
    docsHref: '/docs/plugin-sdk/',
    icon: 'settings',
  },
  {
    slug: 'security-local-first',
    title: 'Local-first security',
    summary: 'Keep the control plane local, repositories untrusted until review, and destructive actions explicit.',
    description:
      'Manage local projects without an Envarium account. Review repository commands before execution and confirm destructive actions. Credentials use the operating-system secret store. Your projects and optional integrations may still need internet access.',
    proof: ['Run core workflows offline without an account', 'Approve repository commands before execution', 'Preview destructive actions before authorizing them'],
    docsHref: '/docs/security/threat-model/',
    icon: 'health',
  },
]
