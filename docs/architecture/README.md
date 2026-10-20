---
title: Architecture overview
description: Process topology, bounded contexts, dependency direction, and accepted Envarium decisions.
category: concept
audience: [contributor, integrator, maintainer]
since: 1.0.0
lastVerified: 2026-10-19
slug: docs/architecture
---

Envarium is a local-first modular monolith delivered primarily as one Go
binary. The process hosts a daemon, CLI commands, local HTTP/WebSocket and IPC
transports, and an MCP façade. A Vue application and thin Tauri desktop shell
consume the same application behavior.

Use the focused architecture guide for the boundary you are changing:

- [Durable operations and local transport](operations-kernel.md): state
  transitions, recovery, replay, IPC, and browser authentication.
- [Project onboarding](project-onboarding.md): discovery, repository trust,
  manifest validation, and configuration provenance.
- [Docker Compose](docker-compose-runtime.md) and [native processes](native-process-runtime.md):
  lifecycle drivers, identity evidence, and ownership limits.
- [Observability](observability.md): redaction, log retention, health, metrics,
  history, and storage attribution.
- [Ports, source control, and actions](developer-workflows.md): port evidence,
  Git observation, reviewed commands, and audit recovery.
- [Dashboard and product UI](dashboard-alpha.md): routes, server-state ownership,
  command boundaries, accessibility, and degraded states.
- [Agent integration](agent-integration.md) and [AI-assisted onboarding](ai-assisted-onboarding.md):
  scoped MCP tools, isolated providers, bounded evidence, and human approval.
- [Workspaces and worktrees](workspaces.md): dependency execution, checkout
  isolation, exact ports, routing, and recovery.
- [Terminal sessions](terminal-sessions.md): PTY ownership, authenticated streams,
  detach behavior, scrollback, and agent metadata.
- [Desktop shell](desktop-shell.md): sidecar compatibility, native presentation,
  notifications, and signed updates.
- [Plugin architecture](plugins.md) and [SDK](../plugin-sdk.md): executable trust,
  protocol negotiation, capabilities, supervision, and authoring.
- [Diagnostics and automation](intelligent-diagnostics.md): evidence validation,
  deterministic rules, reviewed actions, recipes, and local feedback.
- [Federation](../federation.md) and [team configuration](../team-configuration.md):
  optional peer identity, mutual TLS, signed bundles, encrypted sync, and policy.

These guides describe the current source. Check [platform support](../platform-support.md)
and [downloads](/download/) for released packages and supported platforms.

## Process topology

```text
CLI          Browser/Vue          Tauri          MCP clients
 |                |                 |                 |
 +---------- local IPC / loopback HTTP / stdio -------+
                              |
                   transport adapters
                              |
                   application use cases
                              |
                       domain modules
                              |
          focused infrastructure adapter interfaces
               /        |        |        \
           SQLite     Docker   processes   Git/OS
```

The daemon owns project truth. Clients never call Docker, SQL, or operating
system process APIs directly. Tauri/Rust contains packaging and native-shell
integration only.

## Dependency direction

```text
transport and UI adapters -> application -> domain
infrastructure adapters ----- implement ---> application ports
```

A domain owns its model, invariants, use cases, persistence interfaces, and
adapter contracts. Domain code must not import transport, database, Docker,
operating-system command, desktop, or AI-provider packages. A domain cannot
read another domain's tables. Cross-domain behavior uses an explicit
application interface or a typed event.

## Bounded contexts

| Context | Responsibility |
|---|---|
| `catalog` | Projects, locations, tags, and trust/approval |
| `manifest` | Configuration, overlays, provenance, proposals, schema versions |
| `operations` | Durable operations, progress, cancellation, serialization |
| `runtime` | Driver contracts, runs, reconciliation, lifecycle execution |
| `observability` | Logs, health, events, metrics, and retention |
| `ports` | Declarations, reservations, bindings, conflicts, and leases |
| `sourcecontrol` | Git repository facts and worktrees |
| `actions` | Typed, risk-classified project actions |
| `workspace` | Multi-project graphs and coordinated lifecycle |
| `agents` | MCP, provider adapters, proposals, sessions, and agent audit |
| `diagnostics` | Bounded evidence, deterministic findings, reviewed suggestions, local feedback, alerts, and automation policy |
| `terminal` | Owned PTYs, bounded streams, typed interactive launches, and session audit |
| `plugins` | External package discovery, executable trust, scopes, supervision, and typed actions |
| `fleet` | Authenticated peer identities, bounded inventory, grants, remote operations, and audit |
| `team` | Publisher trust, signed portable bundles, restrictive policy, registry metadata, and encrypted sync |
| `telemetry` | Explicit consent, fixed usage counters, payload preview, periodic delivery, and opt-out deletion |
| `settings` | Revisioned local preferences, safe project roots, adapter choices, and restart-effect reporting |
| `platform` | OS capabilities without product policy |

`foundation` is limited to stable primitives such as clocks, IDs, event
envelopes, pagination, and problem details. It is not a miscellaneous shared
code layer. Root packages named `utils`, `common`, or `helpers` are forbidden.

## Composition

Construct dependencies explicitly at a small composition root. Interfaces live
with their consumer and exist only for a real boundary, alternate adapter, or
test seam. Long-running work receives a context, has defined cancellation, and
records progress and outcome. Architecture checks enforce dependency rules.

Accepted decisions and their consequences are indexed in
[`docs/adr`](../adr/README.md).
