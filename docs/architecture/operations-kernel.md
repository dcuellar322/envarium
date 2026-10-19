---
title: Durable operations and local transport
description: Operation state, idempotency, cancellation, local IPC, browser sessions, and audit.
category: concept
audience: [contributor, integrator]
since: 1.0.0
lastVerified: 2026-10-19
---

The operations service coordinates durable mutations for project, runtime,
workspace, UI, CLI, and agent adapters.

## Operation lifecycle

Every mutation is represented by one durable operation:

```text
queued -> running -> succeeded
                  -> failed
                  -> cancelled
                  -> partially_succeeded
queued            -> cancelled
                  -> failed
```

Terminal states never transition again. Operation steps and audit events are
append-only. State changes use a compare-and-swap update so a concurrent writer
cannot silently overwrite another terminal decision.

An idempotency key is unique within a project. Repeating the same request
returns the existing operation and never schedules a second executor. HTTP
mutation adapters require an opaque `Idempotency-Key` between 8 and 128
characters; application services enforce the durable uniqueness rule.

## Scheduling and cancellation

A keyed gate permits one active lifecycle mutation per project. Operations for
different projects can execute concurrently. Cancellation is durable before
the live executor receives context cancellation, so a daemon restart cannot
forget the request.

Daemon shutdown stops accepting requests, cancels active operation contexts,
waits for terminal transitions, closes HTTP and IPC listeners, and only then
closes SQLite.

## Restart behavior

Recovery is deterministic:

| Persisted state | Recovery action |
|---|---|
| `queued` | Resume through the registered executor. |
| `queued` with cancellation requested | Mark `cancelled`. |
| `running` | Mark `failed` with `DAEMON_RESTARTED`; never guess whether an external side effect completed. |
| terminal | Leave unchanged. |

Runtime executors reconcile observed external state before a user retries
an interrupted operation.

## Events and replay

Operation events are inserted into the SQLite journal before live publication.
The sequence is monotonic. WebSocket clients reconnect with
`/ws/v1/events?after=<last-sequence>`. The daemon replays at most 500 events;
`system.refresh_required` tells a client to refresh query state when that bound
is exceeded. Slow live subscribers are disconnected and recover through replay.

## Local transport and browser sessions

CLI traffic uses HTTP semantics over a user-permissioned Unix-domain socket on
macOS and Linux. Their socket and database use mode `0600`, and the data
directory uses `0700`. The Windows source adapter uses a named pipe with a
protected DACL granting access to LocalSystem and the daemon owner.
**Windows support is coming soon**; see [platform support](../platform-support.md).

The loopback browser API is a separate router:

1. `envarium ui` requests a one-time, one-minute bootstrap token over local
   IPC.
2. The browser exchanges that token once for an origin-scoped proof, kept in
   session storage. No bearer cookie is issued. Idle and absolute expiry apply.
3. Every browser API request carries the proof in `X-CSRF-Token`. Mutations
   also require an exact Origin and an idempotency key.
4. WebSockets require the proof in an authentication subprotocol and an exact
   same-origin `Origin`. Only the public `envarium.v1` protocol is negotiated.
   The server never echoes the secret as its selected
   protocol. A cookie from another loopback port has no authority.

The replay journal retains the latest 100,000 sequence positions, with at most
1,023 pending cleanup between compactions. Pruned cursors receive
`system.refresh_required`; clients reload authoritative state. Audit records
are retained separately. SQLite reuses freed pages without a blocking VACUUM.

Bootstrap and session responses are not cacheable. Browser responses set a
restrictive content-security policy, no-referrer policy, MIME sniffing
protection, and frame denial.
