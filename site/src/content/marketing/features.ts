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
      'Use your existing Compose configuration or local start commands. Preview what a start, stop, or rebuild will do, then follow its progress and errors.',
    proof: ['Preview start, stop, and rebuild operations', 'Run commands with explicit arguments and working directories', 'Inspect partially running projects and runtime failures'],
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
      'Declarations, reservations, and live bindings stay distinct, with source evidence for every conflict. Git and action adapters expose practical context without turning Envarium into a broad shell or Git GUI.',
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
    summary: 'Give coding agents typed project context and approved operations instead of another unrestricted shell.',
    description:
      'Observe, develop, maintain, and admin profiles are enforced by application services. Tool responses are bounded, structured, redacted, and audited regardless of what repository text or a model requests.',
    proof: ['No generic MCP shell', 'Permission-scoped mutations', 'Schema-constrained proposals'],
    docsHref: '/docs/mcp/',
    icon: 'agents',
  },
  {
    slug: 'terminals',
    title: 'Embedded and external terminals',
    summary: 'Open a real project PTY or hand work off to the terminal you already use.',
    description:
      'Open a terminal in the project directory, use interactive command-line tools, and reconnect to a session. You can also launch your preferred external terminal. Windows terminal support is coming soon.',
    proof: ['Use interactive terminal programs on macOS and Linux', 'Resize and reconnect to project terminal sessions', 'Open a project in an external terminal'],
    docsHref: '/docs/architecture/terminal-sessions/',
    icon: 'terminal',
  },
  {
    slug: 'plugins',
    title: 'Capability-scoped plugins',
    summary: 'Extend stable product contracts without loading arbitrary code into the daemon.',
    description:
      'Plugins run as supervised processes over a versioned protocol, declare capabilities and scopes, and must pass a conformance suite before they are trusted.',
    proof: ['Run plugins in separate supervised processes', 'Review the capabilities a plugin requests', 'Build against the documented plugin protocol'],
    docsHref: '/docs/plugin-sdk/',
    icon: 'settings',
  },
  {
    slug: 'security-local-first',
    title: 'Local-first security',
    summary: 'Keep the control plane local, repositories untrusted until review, and destructive actions explicit.',
    description:
      'Essential operation works offline. Privileged clients use user-scoped IPC, browser access stays on loopback with session and CSRF checks, and secrets remain keychain references.',
    proof: ['Run core workflows offline without an account', 'Approve repository commands before execution', 'Preview destructive actions before authorizing them'],
    docsHref: '/docs/security/threat-model/',
    icon: 'health',
  },
]
