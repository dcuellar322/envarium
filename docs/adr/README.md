---
title: Architecture decision records
description: Accepted technical decisions and the process for superseding them.
category: contributor
audience: [contributor, maintainer, integrator]
lastVerified: 2026-10-19
slug: docs/adr
---

ADRs use the status values `Proposed`, `Accepted`, `Superseded`, or `Rejected`.
Changing an accepted decision requires a new ADR that links to and supersedes
the old one.

## Accepted decisions

- [ADR-0001: Modular monolith and dependency direction](0001-modular-monolith.md)
- [ADR-0002: Go control plane](0002-go-control-plane.md)
- [ADR-0003: One Envarium binary](0003-one-binary.md)
- [ADR-0004: REST, WebSocket, and OpenAPI](0004-rest-websocket-openapi.md)
- [ADR-0005: SQLite, sqlc, migrations, and log segments](0005-sqlite-persistence.md)
- [ADR-0006: Docker SDK observation and Compose CLI lifecycle](0006-docker-compose-runtime.md)
- [ADR-0007: Native process ownership and reconciliation](0007-process-ownership.md)
- [ADR-0008: Manifest precedence and provenance](0008-manifest-precedence.md)
- [ADR-0009: Thin Tauri desktop shell](0009-thin-tauri-shell.md)
- [ADR-0010: MCP-first provider-neutral agent integration](0010-mcp-first-agents.md)
- [ADR-0011: Agent permissions and no generic shell tool](0011-agent-permissions.md)
- [ADR-0012: Out-of-process plugin protocol](0012-out-of-process-plugins.md)
- [ADR-0013: Local IPC and browser session security](0013-local-transport-security.md)
- [ADR-0014: Log retention and redaction](0014-log-retention-redaction.md)
- [ADR-0015: Platform support order](0015-platform-order.md)
- [ADR-0016: Optional peer federation and signed shared configuration](0016-optional-federation.md)
- [ADR-0017: Static public site and documentation portal](0017-static-public-site.md)
- [ADR-0018: Project-first runtime operations](0018-project-runtime-operations.md)

- [ADR-0019: Log memory ownership](0019-log-memory-ownership.md)

- [ADR-0020: Background configuration resolution without CLI polling](0020-background-configuration-resolution.md)

- [ADR-0021: Direct macOS socket inventory](0021-direct-macos-socket-inventory.md)

- [ADR-0022: Bound default daemon CPU execution parallelism](0022-daemon-execution-parallelism.md)

- [ADR-0023: Scheduled diagnostic evidence](0023-scheduled-diagnostic-evidence.md)

- [ADR-0024: Shared observations and sampling isolation](0024-shared-observation-and-sampling-isolation.md)

- [ADR-0025: Origin proof and immutable reviewed inputs](0025-origin-proof-and-reviewed-inputs.md)

- [ADR-0026: Reviewed checkout transitions](0026-reviewed-checkout-transitions.md)

- [ADR-0027: Native worktree ownership and reviewed transitions](0027-native-worktree-runtimes.md)

- [ADR-0028: Published release updater feed](0028-published-release-updater-feed.md)

- [ADR-0029: Retired documentation routes](0029-retired-documentation-routes.md)

- [ADR-0030: First release platform scope](0030-first-release-platform-scope.md)
