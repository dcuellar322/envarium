---
title: "ADR-0006: Docker SDK observation and Compose CLI lifecycle"
description: Observe Docker through its API and execute lifecycle through the installed Compose CLI.
category: concept
audience: [contributor, integrator]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-07-15

## Context

The Docker Engine API is strong for inspection, events, logs, and metrics, but
Compose lifecycle semantics also depend on the user's installed plugin,
contexts, credential helpers, and configuration.

Current refinement: [ADR-0020](0020-background-configuration-resolution.md)
uses an in-process loader for background observation. The installed Compose CLI
remains authoritative for lifecycle planning and execution.

## Decision

Use the official Docker Go SDK with API negotiation for observation and live
streams. Use the installed `docker compose` CLI for normalized configuration
and lifecycle execution. Identify projects and services through canonical
Compose labels, never container-name prefixes.

## Consequences

Envarium matches commands users already trust while retaining typed
observation. Command building, execution, config reading, event watching, logs,
metrics, and reconciliation remain focused collaborators. Docker failure cannot
disable non-Docker project behavior.
