---
title: Generated MCP tool index
description: Tool names, profiles, risks, and descriptions generated from the MCP server registration.
category: reference
audience: [user, integrator, contributor]
since: 1.0.0
lastVerified: 2026-07-17
---

Tools call application services and are omitted when the selected permission
profile cannot use them. Do not edit this generated page manually.

| Tool | Title | Minimum profile | Risk | Purpose |
|---|---|---|---|---|
| `envarium_action_run` | Run trusted action | develop | declared | Queue one reviewed manifest action; risk confirmation is enforced. |
| `envarium_actions_list` | List trusted actions | observe | read-only | List bounded trusted project actions and their risk metadata. |
| `envarium_environments_list` | List project environments | observe | read-only | List registered Git worktrees and exact runtime leases. |
| `envarium_environments_register` | Register worktree environments | develop | filesystem-read | Reconcile trusted Git worktrees and allocate exact ports without running repository code. |
| `envarium_manifest_explain` | Explain effective manifest | observe | read-only | Read the effective trusted manifest and field provenance. |
| `envarium_manifest_proposal_accept` | Accept manifest proposal | admin | trust-decision | Accept a previously validated proposal as an explicit trust decision. |
| `envarium_manifest_proposal_create` | Create manifest proposal | maintain | filesystem-read | Scan a local repository without executing its code and return an untrusted proposal for review. |
| `envarium_operation_cancel` | Cancel operation | develop | mutating | Request cooperative cancellation of one durable operation. |
| `envarium_operation_get` | Get operation | observe | read-only | Read one durable operation and its terminal or cancellation state. |
| `envarium_operation_wait` | Wait for operation | observe | read-only | Wait up to 30 seconds for an operation and emit MCP progress notifications. |
| `envarium_ports_list` | List local ports | observe | read-only | Read bounded declarations, leases, listeners, and conflicts. |
| `envarium_ports_suggest` | Suggest local port | observe | read-only | Find an available port in an explicit bounded range without reserving it. |
| `envarium_project_deploy` | Deploy project | maintain | networked-mutating | Pull newer Compose images when available and reconcile the current trusted definition, or restart native services with their current definition. |
| `envarium_project_get` | Get Envarium project | observe | read-only | Read one registered project by opaque identifier. |
| `envarium_project_git_status` | Get project Git status | observe | read-only | Read bounded Git branch, changes, remotes, worktrees, and last commit. |
| `envarium_project_health_wait` | Wait for project health | observe | read-only | Wait up to 30 seconds for structured project health to become healthy. |
| `envarium_project_health` | Get project health | observe | read-only | Read structured persisted health diagnostics. |
| `envarium_project_kill` | Force-stop project | develop | interrupting-mutating | Forcefully terminate a project or selected declared services without waiting for graceful shutdown. |
| `envarium_project_logs_query` | Query project logs | observe | read-only | Read bounded, redacted recent project logs. |
| `envarium_project_pause` | Pause project | develop | mutating | Queue an idempotent pause for a project or selected declared services. |
| `envarium_project_rebuild` | Rebuild project | maintain | mutating | Queue an explicit rebuild for a project or selected declared services. |
| `envarium_project_restart` | Restart project | develop | mutating | Queue a restart for a project or selected declared services. |
| `envarium_project_resume` | Resume project | develop | mutating | Queue an idempotent resume for a project or selected declared services. |
| `envarium_project_services` | List project services | observe | read-only | Read bounded service observations for one project. |
| `envarium_project_start` | Start project | develop | mutating | Queue an idempotent start for a project or selected declared services. |
| `envarium_project_status` | Get project status | observe | read-only | Read catalog, runtime, and health status for one project. |
| `envarium_project_stop` | Stop project | develop | mutating | Queue an idempotent stop for a project or selected declared services. |
| `envarium_project_teardown` | Tear down project | admin | destructive | Tear down Compose resources; volume removal is explicit and destructive. |
| `envarium_projects_list` | List Envarium projects | observe | read-only | List a bounded set of registered projects and trust states. |
| `envarium_routes_list` | List local routes | observe | read-only | List friendly localhost routes visible to this agent scope. |
| `envarium_system_info` | Envarium system info | observe | read-only | Read daemon version, API, schema, and readiness. |
| `envarium_workspace_get` | Get workspace | observe | read-only | Read one workspace dependency graph and latest execution. |
| `envarium_workspace_start` | Start workspace | develop | mutating | Queue dependency-ordered workspace start with bounded concurrency. |
| `envarium_workspace_stop` | Stop workspace | develop | conditional-destructive | Queue dependency-safe workspace stop; data removal is separately authorized. |
| `envarium_workspaces_list` | List workspaces | observe | read-only | List bounded workspace graphs visible to this agent scope. |
