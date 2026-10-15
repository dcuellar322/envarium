---
title: MCP tools, resources, and permission profiles
description: Bounded Envarium MCP tools, resources, prompts, permissions, and agent workflows.
category: reference
audience: [user, integrator]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-07-17
searchTerms: [Codex MCP, Claude Code MCP, envarium_project_start, permission profile]
---

Envarium's MCP server makes local project operations available to Codex,
Claude Code, and other MCP clients without asking them to guess Docker or shell
commands. It uses stdio and connects to the ordinary Envarium daemon over its
privileged local IPC endpoint.

## Install

Build or install the `envarium` binary, then from a repository root run:

```bash
envarium agent install codex --profile develop
envarium agent install claude --profile develop
```

Run only the installer matching the client you use. `observe` is the safer
default and has no mutation tools. Use `--project PROJECT_ID` one or more times
to constrain reads and mutations to explicit registered projects. A restricted
server cannot scan/register a new project, and proposal acceptance resolves the
proposal's owner before authorization. Inspect all options with
`envarium agent install --help`.

To configure another MCP client directly:

```json
{
  "type": "stdio",
  "command": "/absolute/path/to/envarium",
  "args": [
    "--data-dir", "/absolute/path/to/data",
    "mcp", "serve",
    "--transport", "stdio",
    "--provider", "generic",
    "--agent-id", "my-client",
    "--profile", "observe"
  ]
}
```

MCP owns stdout while running. Diagnostics go to stderr.

## Tools

All tool names are stable and prefixed with `envarium_`. Inputs reject
unknown fields through SDK-generated JSON Schemas. Identifiers are opaque;
obtain project, service, action, proposal, and operation IDs from read tools.

### Observe

| Tool | Key input | Bound or behavior |
|---|---|---|
| `envarium_system_info` | none | Daemon/API/schema readiness |
| `envarium_projects_list` | `limit` | Default 50, maximum 100 |
| `envarium_project_get` | `projectId` | One catalog project |
| `envarium_project_status` | `projectId` | Catalog + runtime + health |
| `envarium_project_services` | `projectId` | Maximum 100 services |
| `envarium_project_logs_query` | filters, `tail` | Redacted; default 100, maximum 500 |
| `envarium_project_health` | `projectId` | Current persisted health |
| `envarium_project_health_wait` | `projectId`, `timeoutSeconds` | Wait for healthy; maximum 30 seconds |
| `envarium_project_git_status` | `projectId` | Bounded structured Git facts |
| `envarium_ports_list` | optional `projectId`, `limit` | Maximum 500 facts and 100 conflicts |
| `envarium_ports_suggest` | range, protocol, `requestId` | At most 10,001 candidates; no reservation |
| `envarium_actions_list` | `projectId` | Maximum 100 accepted actions |
| `envarium_operation_get` | `operationId` | Durable current state |
| `envarium_operation_wait` | `operationId`, `timeoutSeconds` | Terminal-state wait; maximum 30 seconds |
| `envarium_manifest_explain` | `projectId` | Effective manifest and provenance |

### Develop and above

`envarium_project_start`, `envarium_project_stop`,
`envarium_project_restart`, `envarium_project_kill`, `envarium_project_pause`, and
`envarium_project_resume` accept:

```json
{
  "projectId": "project-opaque-id",
  "serviceIds": ["api"],
  "requestId": "stable-request-123"
}
```

Omit `serviceIds` for the whole project. A targeted start includes required
native-process dependencies. Targeted Compose operations use only declared
service names. Teardown cannot be targeted.

`envarium_action_run` accepts `projectId`, `actionId`, `requestId`, and the
explicit `confirmRisk` and `allowOutsideRoot` gates. The develop profile cannot
run a destructive action even when confirmation is supplied.

`envarium_operation_cancel` accepts an operation ID and request ID. It asks
the durable coordinator to cancel and returns the updated operation.

### Maintain and admin

Maintain adds `envarium_project_deploy`, `envarium_project_rebuild`, and
`envarium_manifest_proposal_create`. Proposal creation performs deterministic
safe reads and does not execute repository code.

Admin adds `envarium_manifest_proposal_accept` and
`envarium_project_teardown`. Teardown accepts `removeVolumes`; the boolean
must be explicit and the tool is annotated destructive. Accepting a proposal
is an explicit trust transition and never happens during proposal creation.

## Resources and prompts

Static resources are `envarium://system` and `envarium://projects`.
Project templates provide catalog, status, manifest, services, Git, ports, and
at most 50 recent redacted errors under
`envarium://projects/{projectId}/...`.

The server publishes four prompts:

- `envarium_onboard_project`
- `envarium_diagnose_project`
- `envarium_start_and_verify`
- `envarium_resolve_port_conflict`

Prompts describe safe tool workflows; they grant no extra capability.

## Typical lifecycle

1. Call `envarium_projects_list` and select an opaque project ID.
2. Call `envarium_project_status`.
3. Call `envarium_project_start` with a new stable request ID.
4. Call `envarium_operation_wait` with the returned operation ID.
5. Call `envarium_project_health_wait`.
6. If unhealthy, read `envarium_project_health` and bounded redacted logs.
7. To restart one declared service, call `envarium_project_restart` with its
   service ID, wait for the operation, and verify health again.
8. Use `envarium_project_deploy` only when image pulling/reconciliation is
   intended, and `envarium_project_kill` only when graceful stop is unsuitable.

Never retry a different mutation with the same request ID.

## Inspector

The repository pins the smoke-test version and runs it with:

```bash
make test-mcp-inspector
```

The test builds the production binary, starts an isolated daemon, invokes the
MCP Inspector CLI over stdio, and validates discovery plus default destructive
tool omission.
