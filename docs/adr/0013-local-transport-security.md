---
title: "ADR-0013: Local IPC and browser session security"
description: Protect privileged local clients and loopback browser mutations.
category: concept
audience: [contributor, integrator]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-07-15

## Context

CLI, desktop, MCP, and browser clients need local access to privileged
capabilities. Loopback binding alone does not prevent cross-origin requests or
access by other local users.

The original cookie-based browser authentication below is superseded by
[ADR-0025](0025-origin-proof-and-reviewed-inputs.md). Current browser clients
use origin-scoped session proof for every HTTP and WebSocket request; bearer
cookies carry no authority. Privileged IPC remains unchanged.

## Decision

Use user-permissioned Unix sockets on macOS/Linux and named pipes on Windows for
privileged local clients. Serve the browser API on loopback only. A local
bootstrap handshake issues a short-lived same-origin cookie; mutations require
CSRF validation. Validate origins for HTTP and WebSocket traffic, disallow
permissive CORS, expire idle sessions, and audit mutations.

## Consequences

Local clients share handlers while retaining transport-specific authentication.
Socket/pipe cleanup and permissions are tested. Browser JavaScript never
receives the Docker socket or generic process-execution capability.

Current refinements: [ADR-0025](0025-origin-proof-and-reviewed-inputs.md)
supersedes the browser cookie, replay retention and prelaunch compatibility
exceptions where applicable.
