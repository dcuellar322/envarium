---
title: "ADR-0024: Share observations and isolate resource sampling"
description: Bound presentation-driven work while preserving live monitoring and explicit freshness.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

- Status: Accepted
- Date: 2026-10-17
- Refines: ADR-0004, ADR-0006, ADR-0009 and ADR-0022

## Context

With sixteen real containers running, the installed native dashboard consumes
about 67% of one core whether visible or hidden. Most CPU is charged to Git
children. Runtime event invalidations refetch unrelated Git/actions across all
projects and bypass background polling suppression. Metric sampling blocks on
storage scans and reads container statistics sequentially inside a four-second
project deadline; raw history has gaps of up to 39 seconds and unavailable values.

## Decision

Keep runtime, health and metric observation active independently of presentation.
Scope query invalidations to their actual data, use independent shared query
keys for Git/actions/runtime/metrics, and suppress presentation requests while
hidden. Refresh current data when shown. Use document visibility and native
shell visibility/focus signals as presentation facts only (unfocused desktop
presentation also pauses). Suspend the UI event connection while inactive and
resume from its last sequence when shown; tray notifications
and daemon monitoring continue. Transport cancellation reaches the server.

Coalesce concurrent trusted Git requests without weakening trust checks or
caching mutable working-tree status. Reuse immutable commit metadata by HEAD and
validated Git directory discovery; re-read current status, remotes, worktree
inventory and operation markers. Bound retained metadata and copy caller-owned
results. Failed and canceled reads are not successful cache entries.

The dashboard consumes a dedicated fresh working-tree status read; full Git
reads retain commit, remote and worktree enrichment for detail, CLI and MCP
consumers. Give these reads different query/coalescing keys, document omitted
enrichment, and preserve identical trust, branch, edit and operation checks.
The native tray compares only its rendered menu inputs before rebuilding;
snapshot polling and notification processing continue regardless.

Compute the canonical manifest declaration identity once during exact-input
resolution, retaining it internally with the typed copy. Runtime observers use
that identity without repeatedly serializing unchanged declarations. SQLite
footprint reads separate fast counts from indexed oldest-point seeks; rollups
resolve each service's watermark once while retaining late-bucket/idempotent
semantics.

Move storage enrichment onto an independently canceled five-minute worker,
matching the existing storage cache horizon. Metric collection and overview
reads use its latest explicitly partial snapshot and cannot wait for disk scans.
Explicit storage/cleanup preview reads retain their existing inventory behavior.
Collect container statistics with bounded concurrency and serialize sink writes;
one slow container cannot consume the deadline for every later service. Preserve
partial/error reporting and resource intervals. Test freshness, cancellation,
concurrency, failure, scheduling and client isolation.

## Acceptance evidence

Measure the normal installed CLI and native bundle with the same sixteen
containers, counting child CPU and attributable WebKit helpers. Aim for under
1% closed CPU, hidden overhead within 0.5 percentage points of closed, and under
5% visible CPU. Require complete current service metrics and approximately
ten-second raw cadence, including storage refresh and scheduled diagnostics.
Longer memory observations must distinguish RSS from charged footprint and
include sampled transient children. Record achieved values and remaining limits
in the Phase 12 progress log; targets are not hard CPU limits.

The first live candidate still reconciles every Docker health-check exec event.
Ignore only known exec lifecycle chatter in the Compose watcher: it changes
neither container identity nor lifecycle/health state. Continue forwarding actual
health-status, lifecycle, metadata and unknown future events. Resource sampling
and health probes remain independent. This reduces redundant observation and
journal writes without hiding state transitions or inventing longer intervals.

The health scheduler must retain its one-second source checks. Reuse bounded
runtime projections by freshly resolved canonical manifest hash, project root
and slug, rather than cloning actions/provenance and rebuilding identical
runtime declarations. Trust is read before every lookup; current local bytes,
accepted revision, source errors and environment availability/overlays are
checked on every resolution. Retained runtime projections have a four-MiB
weighted budget and 128-entry cap, and returned nested values are copied.

Reuse at most 64 prepared fixed SQL statements / 64 KiB of SQL per SQLite adapter
instance; parameters and result values are never cached. Preparation failures
and canceled requests do not poison later reads, and the database owns statement
closure. Prepare metric inserts/upserts once per transaction rather than once
per point. Keep transaction atomicity, live query semantics and current schema.

The next native probe still spends substantial CPU in per-project HTTP traffic
and the Apple Git launcher. Batch presentation reads by kind (runtime, Git
status, or actions), with at most 64 unique IDs, four application-service calls
in flight, ordered results, explicit per-project warnings and cancellation.
Each owning service retains current permission/freshness checks. Separate
kind-specific TanStack Query keys prevent runtime events from refetching Git.
Single-project detail/CLI/MCP reads retain their APIs. The dashboard resource
sampler remains independent. Refresh all current facts on an event replay gap.

Connected idle operation presentation uses event invalidation plus a fifteen-
second safety poll; active operations and disconnected presentation retain the
1.5-second poll. Native sidecars request compact JSON to avoid per-line subprocess
framing overhead, retaining the exact envelope. Resolve Apple's selected Git
toolchain once per observer instead of invoking its launcher for every status;
honor custom PATH and fall back if the selected executable disappears.

Host presentation consumes the resource worker's cached disk summary rather
than initiating a second Docker filesystem traversal. Unknown storage remains
unknown; a cheap ping can still report Engine connectivity. Explicit disk reads
and the independent five-minute enrichment worker retain their behavior.

Each scheduled diagnostic cycle captures global port/resource read models once
through an explicit scoped collector; project trust, health, logs and approved
actions remain independently current. Global-source failures remain partial
warnings in each affected bundle. Interactive collection retains its own reads.
Health history seeks the latest indexed row per check, preserving timestamp/ID
tie-breaking and all check scopes, instead of ranking every historical payload.

Apply the existing one-core I/O execution budget at the CLI process entry as
well as daemon entry, respecting explicit GOMAXPROCS overrides. Only the current
Go runtime is configured: environment variables are unchanged, so autostarted
daemons and project subprocesses retain their own configured/default settings.

Incremental metric rollups enumerate source series, seek each target watermark,
and read each source's indexed timestamp range. Keep the existing schema,
chronological input order, per-series watermarks, late-current-bucket behavior
and idempotent upserts. Do not scan all retained payload rows merely to discard
those preceding the watermark.

Only runtime status observation requests Docker server-version metadata.
Metrics, log and event connections retain API negotiation and report their own
request failures without fetching unused version payloads. Read exact resource
footprint aggregates in one SQLite snapshot, retaining independent COUNT and
indexed MIN plans while avoiding repeated WAL lock acquisition.

Native tray snapshots use a bounded `status` batch containing the same current
runtime and health facts, rather than two HTTP requests per project. Health
reconciliation of degraded/starting runtime states belongs to the shared
application read used by both single-project HTTP and batched readers. Preserve
per-project partial warnings, current trust checks, notification inputs and the
fifteen-second native polling interval. Older daemons may use the existing
single-project fallback when the additive batch kind is unsupported.

Reuse at most eight idle Docker SDK clients for explicit local Unix endpoints,
owned by the Compose driver and closed on shutdown. Each active lease has its
own client; all API negotiation, status metadata and runtime/metric requests stay
fresh. Environment-selected and non-Unix/TLS connections keep fresh clients so
credential/environment changes cannot be hidden by pooling. Failed negotiation
discards its client. Bound retained clients across all endpoint identities;
returning a lease after shutdown closes it immediately.

Treat query payloads as immutable replacement values: use shallow Vue Query
state for dashboard/host/operation reads, and disable their presentation timers
while inactive. Skip global invalidation timers when an event changes none of
the subscribed read models. Retain monitoring and native notification polling.

Bound the native notification adapter's recently seen operation identities to
1,024 entries. Cache fixed health write plans under the same database statement
budget, without caching result values. Reuse the already serialized log entry's
size when computing its fingerprint rather than serializing it again solely to
estimate segment rotation.

Contain dashboard-card layout and painting within each card, preserving its
intrinsic sizing, keyboard controls and existing visual hierarchy. Validate
desktop/narrow/hover rendering and actual native performance before accepting.
