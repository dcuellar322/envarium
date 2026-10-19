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

Resolve sibling optional manifests through one confined, identity-checked
directory handle per resolution. Read each current file and validate its opened
descriptor and byte cap independently; retain no directory or byte cache across
resolutions. A missing directory is checked again on the next resolution.
This removes repeated parent walks without allowing symlink aliases or stale
manifest contents.

Read registered environments and their exact leases in one adapter-owned SQL
snapshot. Scheduler snapshots may consume a typed batch of those current
environment allocations instead of listing them and immediately fetching each
one again. Preserve independent direct lifecycle reads, per-target trust and
availability checks, caller-owned maps, cancellation and explicit errors.

Drive background health, event-subscription reconciliation and log inventory from
one fresh one-second project observation feed. Each subscriber owns a copied,
bounded latest snapshot; slow consumers do not block monitoring or queue stale
cycles. Failed cycles remain explicit. Direct requests and lifecycle execution
continue resolving current trust and files independently. The feed has an
explicit daemon-owned lifetime and retains no interactive query cache. This
preserves the fastest existing configuration check cadence while eliminating
independent duplicate catalog, manifest and environment reads.

On Darwin and Linux, traverse manifest parent directories using pinned directory
file descriptors and kernel-enforced `openat(O_DIRECTORY|O_NOFOLLOW)` boundaries.
This replaces redundant pre/post directory stats with an atomic no-symlink
open, while retaining pre-open regular-file checks, nonblocking/no-follow file
opens, opened-descriptor identity checks and post-read byte caps. Synchronize
close against opens to prevent descriptor reuse from changing the capability.
Other platforms retain the rooted Go filesystem implementation. Retain no
handles or bytes between observations, and verify symlink, FIFO, replacement,
close/cancellation and cross-platform behavior before accepting the change.

Native process watchers retain their two-second reconciliation cadence and
immediate lifecycle events, but share one fresh batch of durable run records per
cycle. Batch only through the run store's explicit domain interface (at most 64
project IDs per SQL read); preserve the newest 200 runs per project and every
stored fingerprint. Each subscriber receives owned values, failed reads stop
that watch explicitly, slow subscribers retain only the latest cycle, and the
feed stops when its final watcher exits. Do not cache run results between cycles
or weaken PID/group ownership verification.

Resolve current trusted runtime inputs once per combined health or resource
read. The runtime application service can return both its resolved project and
current observation, or observation plus statistics, to the owning monitoring
adapter. Reuse only that call's typed inputs; preserve direct trust/file checks,
active/inactive classification, existing status coalescing, driver failures and
metric sampling. Alternate adapters retain their existing independent reads.

Native port attribution is also an explicit inventory concern: a process being
alive does not prove every configured socket exists. The native driver exposes
an on-demand binding read that intersects current TCP listeners with freshly
verified run/process-group owners and retains actual IPv4/IPv6 addresses. The
port registry uses that application boundary. Its failures do not alter ordinary
runtime, health or metric observations, and no background socket poll is added.

The Compose metric cycle can share a single current container inventory and each
successful post-statistics inspection with status assembly. Its evidence map is
owned by that call and discarded immediately afterward. It is never an aged
restart-count cache: every new sample cycle reads new Engine statistics and
inspections, stopped containers do no statistics work, partial failures remain
explicit, and the pre-inventory ownership token still guards reconciliation.
Direct lifecycle inspection and planning retain their existing fresh paths.

Keep the native tray snapshot a presentation DTO containing the actual menu and
notification inputs. Do not serialize full runtime service inventories,
repository paths, workspace recipes or operation details that the native shell
does not consume. Preserve system/host values, project state and health,
workspace identities, operation transitions, notification contents, conflicts
and unavailable observations. This does not change the public HTTP contract or
ordinary CLI views, nor any observation source or interval.

The short-lived `desktop snapshot` child limits Go parallelism to one and skips
percentage-triggered collection while retaining a 64 MiB soft runtime memory
limit. A read-only benchmark against sixteen running containers reduced child
CPU by about 25% and output from 34 KiB to 7 KiB without increasing peak RSS.
Apply these environment overrides only to that exact leaf command; daemon
startup, UI initialization and mutations retain their ordinary runtime policy.
Tests enforce that boundary. The daemon and managed processes never inherit
these overrides.

The native shell may retain one read-only `desktop snapshot --stream` bridge
instead of launching a process for each fifteen-second tray refresh. Each exact
`snapshot` input line requests the same fresh application reads, with a bounded
deadline and one bounded envelope; no timer or mutation policy moves into Rust.
EOF/cancellation closes the bridge. Native failure, invalid output or timeout
discards the owned child and reconnects on the next ordinary poll. Keep normal
GC and a soft memory limit for this longer-lived adapter. Daemon startup and
mutations never inherit its resource policy. Shared presentation queries must
also share their observer/timer, not merely a query cache key. Card rendering
may skip timestamp-only changes while retaining every displayed value and live
action inputs. Monitoring and presentation freshness intervals stay unchanged.
