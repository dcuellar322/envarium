---
title: "ADR-0019: Log memory ownership"
description: Bound transient capture memory and use durable replay without a duplicate archive cache.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-16
---

- Status: Accepted
- Date: 2026-10-16
- Amends: [ADR-0014](0014-log-retention-redaction.md), in-memory ring placement only

## Context

The archive's per-service ring is populated on every append but has no readers:
queries and reconnect replay already read indexed, redacted NDJSON segments.
Native capture separately retains 10,000 entries per run, including all ended
runs for the daemon's lifetime. Entry counts do not bound large message payloads.

## Decision

Keep native capture rings per project/service across runs, bounded by both the configured entry
count (`--log-ring-entries`, default 2,000) and a one-MiB payload budget. Use circular storage so appending a line
does not shift the entire retained history. Run and operation identity remain on
each entry; historical user queries continue through the canonical archive.

Remove the archive's write-only duplicate ring. The existing `--log-ring-entries`
option bounds native capture directly; it does not allocate an archive cache. Subscribers remain bounded and redaction precedes persistence
and fanout. Archive append state retains at most 128 segments and closes segments idle for
one minute, allowing retention to reclaim completed runs. Log maintenance also
runs once per minute without waiting for another producer write.

Terminal scrollback uses a circular byte buffer with the existing one-MiB limit.
Copy scrollback only when attaching, not on every output chunk.

## Consequences

Restart count no longer multiplies native log memory, oversized lines cannot
inflate capture history without a bound, and archive memory does not duplicate
log bodies. Raw native tails share one recent history per service across runs;
durable log queries retain run filters and existing disk policy. Memory still
scales with distinct observed services and live subscriber queues; this is not
a process-wide memory quota. Slow canonical sinks can still exhaust a bounded
capture window, so this change does not promise lossless overload handling.
