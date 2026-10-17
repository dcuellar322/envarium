---
title: "Phase 12: Resource and storage intelligence"
description: Implementation evidence for Envarium product phase 12.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

## Implemented

- Added one process-wide, cancellable resource sampler with a four-project
  concurrency ceiling, per-project deadlines, active-runtime filtering, and
  idle maintenance throttling. Stopped projects are inspected for state but do
  not invoke driver statistics, health reads, or history writes.
- Added project/service aggregation across every Compose replica and every
  verified managed process-tree member. Samples carry explicit CPU, memory,
  network, disk, health, process-count, restart, partial, and availability
  evidence so a failed read is never rendered as a believable zero.
- Added SQLite migration 8 with atomic/idempotent raw samples, availability-
  aware one-minute and fifteen-minute rollups, aligned tier pruning, bounded
  history queries, and database/WAL/SHM/log/history footprint reporting.
- Added manifest-backed CPU, memory, and storage budgets with warnings only
  after three consecutive fresh samples exceed the configured threshold.
- Added a read-only Docker SDK storage inspector for containers, images,
  volumes, and build cache. It uses canonical Compose project labels and live
  references, caches Engine disk usage for two minutes, and times observations
  out after five seconds.
- Added explicit `exclusive`, `shared`, `estimated`, and `unknown` attribution
  with a human-readable reason on every resource. Image layers and build cache
  are never presented as exact project-exclusive bytes.
- Added a non-executable cleanup preview containing stable exact Engine IDs,
  per-resource attribution, known reclaimable bytes, and unknown-size counts.
  No prune/delete application port, HTTP operation, CLI command, or MCP tool
  exists.
- Added generated aggregate resource, storage inventory, cleanup preview, and
  bounded history REST/TypeScript contracts.
- Replaced the dashboard fan-out with one resource overview query, separate
  cached storage query, retained time-series charts with visible gaps and an
  accessible table, project/service consumers, sustained warnings, retention
  facts, and Envarium's own footprint.
- Replaced the project Storage placeholder with real project-scoped resource
  navigation. Project overview and service rows now distinguish unavailable
  CPU/memory and link into durable `?project=` history/storage context.
- Added daemon flags for sample interval, each metric retention tier, and the
  maximum history response size.
- Promoted log and metric retention bounds into revisioned durable settings.
  Changes are reported as pending until restart so samplers, rollups, and
  pruning jobs always begin together with one coherent policy; flags continue
  to seed a new database.

## Architecture decisions

No new ADR was required. This phase implements ADR-0005's SQLite ownership,
ADR-0006's Docker SDK observation and canonical-label attribution, ADR-0007's
PID-reuse-resistant process ownership, and ADR-0014's bounded retention model.

Resource intelligence remains in the observability bounded context. Runtime
drivers emit raw provider-neutral samples; the observability application layer
owns aggregation, budgets, retention policy, and read models. SQLite and Docker
remain consumer-owned adapters. The Docker storage interface intentionally has
no mutation method.

## Retention and overhead targets

Default metric policy:

| Tier | Interval | Retention | Maximum rows per project/service scope |
|---|---:|---:|---:|
| Exact | 10 seconds | 1 hour | 360 |
| Minute | 1 minute | 24 hours | 1,440 |
| Quarter hour | 15 minutes | 30 days | 2,880 |

History responses contain at most 1,000 points. Log retention remains seven
days or 256 MiB by default. Tier maintenance upserts only the current/new
bucket for an established scope, completes rollups before pruning their source,
and runs at most every fifteen minutes while every project is idle.

The documented Phase 12 idle targets are:

- one sampler ticker for the entire daemon;
- at most four concurrent project observations;
- zero driver-statistics and metric-history writes for stopped projects;
- a 50-stopped-project in-memory collection below 1 ms/op and 64 KiB/op on the
  reference development host;
- an empty live daemon below 1% of one CPU core and 100 MiB resident memory
  after startup settles.

`BenchmarkResourceSamplerFiftyIdleProjects` measured `88,980 ns/op`, `38,644
B/op`, and `215 allocs/op` on the Intel i9-8950HK reference host. An isolated
release daemon with no registered projects settled at `0.0%` CPU and `22.8
MiB` RSS after 32 seconds on that host.

## Tests added

- Sampler aggregation, max-four concurrency, cancellation, project failure,
  stopped-project no-write behavior, maintenance throttling, sustained/fresh
  budgets, bounded history selection, cleanup-preview safety, and the 50-idle
  benchmark.
- Active/stopped runtime adaptation, health latency, per-field availability,
  partial observations, and every runtime state.
- Canonical Docker attribution, shared images/volumes, remote/unknown volumes,
  build cache, disconnected Engine behavior, cache expiry, SDK options, and
  client closure.
- SQLite atomic/idempotent writes, availability-aware rollups, repeated
  maintenance, retention boundaries, bounded ordering, and footprint facts.
- Verified native process-tree metrics, partial member failure, service
  filters, PID reuse, and ended-run exclusion.
- Compose replica samples, disk accounting, restart/process counts, partial
  replica failure, and availability evidence.
- Resource HTTP contract/query/error mapping and the absence of an executable
  cleanup result.
- Resource loading, error, empty, partial, Docker-disconnected, filtering,
  exact preview, history control, accessible table, URL selection, and project
  Storage-tab flows.

## Acceptance criteria status

- [x] Project and service rows identify current available CPU and memory across
  Compose replicas and managed process trees.
- [x] Every storage resource and summary discloses exclusive, shared,
  estimated, or unknown attribution.
- [x] Cleanup preview lists exact IDs and known/unknown reclaimable estimates
  and has no execution capability.
- [x] Raw and downsampled history is transactionally rolled up, pruned to
  configured aligned tiers, and bounded at the response boundary.
- [x] Automated idle sampling meets the documented latency, allocation,
  concurrency, no-statistics, and no-write targets.

## Verification evidence

`make quality` passed on 2026-07-16, covering generated-artifact drift, Go and
Vue lint, architecture constraints, TypeScript checks, all Go and web unit
tests, `go test -race ./...`, schema migration to version 8, `govulncheck`, four
real-runtime browser E2E tests, seven visual-regression tests, and production
Go/web builds.

For live verification, the release binary ran against an isolated data
directory and Unix socket for 32 seconds. It measured `0.0%` CPU and `22.8 MiB`
RSS, below both idle targets. A request to `/api/v1/resources` over that socket
returned the empty-project overview, actual Docker storage inventory,
Envarium footprint, default retention policy, and explicit attribution
warning without starting or mutating any project.

## Known limitations and deferred work

- Cleanup execution and resource throttling are intentionally absent. A future
  mutation must add a new explicit authorization, preview, audit, and race-
  revalidation use case; it cannot extend this read-only inspector implicitly.
- Docker image and build-cache records do not expose enough layer ownership for
  exact project totals. Those records remain shared, estimated, or unknown.
- Network and disk values are cumulative provider counters. The retained chart
  exposes exact observed values and gaps; rate derivation with reset metadata
  can be added as a later coherent analytics slice.

## 2026-10-16 — Bound expensive Docker inventory retries

Live diagnosis found Docker Desktop VM CPU above 200% while application
containers were nearly idle. An 8-second Engine CPU profile accumulated 20.63
CPU-seconds; volume directory walks and image metadata traversal dominated.
Pausing Envarium and allowing in-flight work to drain yielded zero Engine CPU
samples in a 10-second profile. Resuming it reproduced the load.

The resource sampler cached successful storage inventories but retried failed
five-second scans every ten seconds. Engine filesystem work can continue after
the request times out. Cache failures as explicit unavailable results for five
minutes, preserve caller-cancellation semantics, allow the background inventory
30 seconds, and start cache expiry after completion. Host storage observations
also use a five-minute cache while retaining their three-second request bound
and normal CPU/memory refresh cadence. No cleanup or resource deletion occurs.

Targeted adapter and host tests passed with the race detector; targeted go vet
and the CLI build passed. Tests cover failure backoff, retry recovery, caller
cancellation, and CPU refresh while Docker storage is unavailable. The rebuilt
local bin/envarium was installed with a rollback copy at
/tmp/envarium-before-cpu-fix, and the existing daemon was restarted for live
verification. Monitoring thresholds and intervals are unchanged.

Live verification at 15:01 CDT: patched daemon reports ready; all four
monitoring HTTP readiness checks return 200, both Prometheus exporter targets
are Up, and all seven Kuma monitors have fresh Up samples. Docker VM CPU sampled
29.9% after installation versus 232.8% before. A 10-second post-fix Engine
profile recorded 510ms CPU (5.1%), down from 257.84% in the active pre-fix
profile. The normal two-second runtime watcher remains running. Slow storage
inventories can still produce bounded work every five minutes; this change
prevents continuous retries rather than deleting or suppressing inventory.

PR #50 CI follow-up: the Public site CI generated-file check detected that
these progress notes were missing from `site/public/llms-full.txt`. Regenerated
the documentation bundle through the site generator. Site typecheck, lint, all 19 unit
tests with coverage, the static build, and content/build validation passed
locally. The regenerated CLI, manifest, settings, MCP, API, and plugin reference
files remain unchanged.

## 2026-10-16 — Daemon CPU and log-memory review

### Scope and acceptance criteria

This review is a Phase 12 overhead-hardening slice, with Phase 7 log-retention
and Phase 14 terminal-buffer regression coverage. Acceptance criteria are:
remove full-history work from output appends; bound native capture across
restarts; bound archive append state; coalesce observation-only Compose CLI
work without caching lifecycle previews; preserve redaction, durable replay,
ordering, cancellation, and existing API/schema contracts; measure rather than
assume the empty-daemon and 50-idle-project targets.

### Findings and implementation

- The existing live daemon (PID 17536) sampled at 18.5–25.6% CPU and about
  41 MiB RSS. These are observations of the existing configured daemon, not a
  controlled before/after comparison. A five-second macOS stack sample showed
  recurring log collection, accepted-manifest resolution, Docker context/config
  normalization, native listener inspection, and SQLite queries. It does not
  establish a precise CPU percentage attributable to each path.
- Every Compose inspect, metrics read, log-follow retry, and event subscription
  resolved context and normalized configuration through Docker subprocesses.
  Observation calls now share a 15-second, singleflight cache, including explicit
  failures, capped at 128 entries. Changes to the resolved project input produce
  a new key. External Compose-file, `.env`, or Docker-context edits reach reads
  within 15 seconds. Lifecycle plans always normalize fresh; execution still
  uses the existing validated plan and ownership controls.
- Native capture shifted up to 10,000 entries per append and retained a separate
  buffer for every run. It now uses circular storage shared across runs for each
  project/service, with the existing `--log-ring-entries` limit (default 2,000)
  and an independent one-MiB accounting limit for retained entries. Eviction
  clears references; live forwarding preserves run/operation identity. A line
  larger than the capture budget is forwarded without being retained. Raw tails
  cover recent service history; public historical queries still use the archive.
- Native log-follow now subscribes to declared service buffers before the first
  run. It stays attached across restarts without enumerating durable run history,
  avoiding the previous once-per-second retry/query loop for never-started native
  projects. Historical log reads continue through the canonical archive.
- The archive populated a second log-body ring that no query or replay read.
  Removed that write-only copy. Append-state metadata now holds at most 128
  active segments, finalizes evicted segments with checksums, and releases
  segments idle for one minute. Minute-based maintenance also runs when producers
  are quiet, so completed runs become eligible for age/disk retention.
- Terminal output copied up to one MiB of scrollback for every 32-KiB chunk just
  to read its truncation flag, and shifted the retained bytes on overflow. It
  now writes a circular byte buffer and returns truncation metadata directly.
  Attach/reconnect still obtains an isolated chronological snapshot. Subscriber
  payloads retain their existing isolation.
- Native log-follow forwarding workers now have a child cancellation context and
  are joined when the sink fails or the follow request ends, preventing blocked
  forwarders from retaining their queues after an early return.

### Architecture decisions and files

[ADR-0019](../adr/0019-log-memory-ownership.md) amends ADR-0014's placement of
in-memory rings while preserving redaction and durable segment ownership. The
changes remain in runtime, observability, SQLite, terminal, and bootstrap;
there are no API, schema, migration, Rust, or Vue behavior changes. The CLI flag
now describes its native capture entry/byte bounds accurately.

### Tests and measurements

Regression tests cover circular wrap/ordering, exact and oversized byte limits,
independent snapshots, configured native entry limits, buffer reuse across 100
runs, live restart continuity, service isolation, sink failure, cancellation,
archive segment caps/checksums, durable replay after idle closure, retention of
quiet runs, Compose cache expiry/failure recovery/capacity/input changes,
concurrent reads, canceled requests, and uncached lifecycle planning.

Three local benchmark runs on the Intel i9-8950HK development host:

| Workload | Before | After |
|---|---|---|
| 32-KiB terminal chunk with full one-MiB scrollback | 181–186 µs; ~1,048,610 B/op; 2 allocations | 1.13–1.32 µs; 24 B/op; 1 allocation |
| Native log append after filling history | 167–181 µs; 0 allocations | 68.8–69.9 ns; 0 allocations |
| 50 idle projects, existing in-memory sampler benchmark | Previous phase evidence above | 121–126 µs; ~38,200 B/op; 214 allocations |

The native append comparison includes the intentionally smaller byte-bounded
retained window. Both append paths now avoid a full-history shift. These are
microbenchmarks, not measurements of whole-application CPU savings.

An isolated production Go build used a new temporary database/socket, an
unavailable Docker endpoint, and no registered projects or attached UI. After
10 seconds settling, a 30-second measurement recorded 0.04 CPU-seconds
(0.13% of one core), with RSS 28.03–28.17 MiB across six samples. This meets the
empty-daemon targets of <1% CPU and <100 MiB RSS for that scenario. The existing
configured daemon was neither replaced nor restarted. No real project or
container was stopped for the measurement.

Reproduce the allocation/latency checks with:

```sh
GOCACHE="$PWD/.cache/go-build" go test ./internal/terminal/application ./internal/runtime/process -run '^$' -bench 'BenchmarkTerminalOutput|BenchmarkNativeLogAppend' -benchmem -count=3
GOCACHE="$PWD/.cache/go-build" go test ./internal/observability/application -run '^$' -bench BenchmarkResourceSamplerFiftyIdleProjects -benchmem -count=3
```

### Verification results

- `go test ./...` and `go test -race ./...`: passed, including real native
  process lifecycle/cancellation fixtures. Local-port tests required running
  outside the restricted filesystem/network sandbox.
- `go vet ./...` and pinned `golangci-lint run`: passed (zero lint issues).
- `make generate-check repository-check archcheck migrate-check`: passed;
  generated OpenAPI clients, JSON Schema, sqlc code, and migrations stayed clean.
- `make vuln`: zero reachable vulnerabilities. The checker reports three
  module-level advisories without vulnerable imported packages/call paths.
- Frontend lint, typecheck, and all 54 unit tests in 24 files: passed.
- Browser checks: five E2E tests and fourteen visual tests passed.
- Documentation-site generation, typecheck, lint, all 19 tests, static build,
  and content/build validation passed.

### Acceptance status and remaining limits

- [x] Output appends avoid full retained-history copies/shifts.
- [x] Native restart count does not multiply capture-ring retention.
- [x] Archive log bodies are stored once and append state has a fixed ceiling.
- [x] Compose observation commands are coalesced and bounded; plans remain fresh.
- [x] Empty-daemon and in-memory 50-idle-project targets were measured and met.

Whole-product budgets for dozens of real configured projects, an open or
hidden desktop WebView, high-volume logs, and multi-hour sessions remain to be
measured. The empty-daemon result does not establish those budgets.

Memory still scales with distinct service identities, live PTYs, subscriber
queues, provider/OS libraries, and WebView state; the byte limit is an accounted
capture-entry bound, not a hard process RSS quota. Native buffers do not yet
expire solely because a service disappears from a manifest. Existing UI work
also remains: project detail polls runtime/health/metrics every five seconds,
Git/logs every ten seconds, the operation center polls every 1.5 seconds, and the
tray requests a CLI snapshot every fifteen seconds. Native reconciliation can
repeat listener/process scans; Compose log collectors still retry finite streams
every second. These need workload profiles before changing freshness or ownership
semantics. No claim is made that the existing configured daemon's CPU has met
a production budget after these source changes, or that the earlier Docker VM
issue is entirely eliminated in every workload.

Manual follow-up: build/install the reviewed binary with rollback available,
restart the daemon when convenient, and compare the same real project set over
several minutes with the UI open, hidden, and closed. Include daemon CPU/RSS,
Docker VM CPU, CLI child-process counts, WebView CPU/RSS, high-output terminal
sessions, repeated start/stop cycles, and an extended idle soak. Preserve project
ownership and permission checks during any further optimization.

## 2026-10-16 — Local installation and real workload measurements

Built the reviewed working tree with production web assets, semantic version
`1.0.0`, commit marker `2527d9e78a74-dirty`, and build time
`2026-10-16T21:00:06Z`. Installed the Go binary to `bin/envarium` and the native
application's bundled sidecar. This is a local source build, not a published
release. The installed binary SHA-256 is
`45073208991ee72e4bd3f5a8b2f4e07092ba5b495d92dc7fe348eee8e3f8536d`.

The native application's initial error, "running daemon reported an invalid
version", was caused by the old daemon's `dev` metadata. The updated daemon
and sidecar report valid `1.0.0` versions. The installed sidecar's startup
snapshot succeeds, includes eleven projects, and reports no warnings. The
existing app was locally ad-hoc signed before the update; its updated bundle
was re-signed in the same mode and passed deep, strict signature verification.

No lifecycle operations or embedded PTYs were active. Preserved rollback copies
of the original CLI and complete native app bundle, plus a consistent private
SQLite backup, under `.cache/resource-live-20261016/`. Sent SIGTERM only to the
old daemon, waited for a graceful exit, and started the updated binary. The
new daemon reports ready on the same local socket/address and schema version
18. All eleven runtime states are identical before/after, including the four
running `dac-stack` containers. Container IDs, start timestamps, and restart
counts are unchanged. Grafana, Metabase, Prometheus, and Uptime Kuma each return
HTTP 200; both Prometheus targets are Up.

The actual workload has eleven registered projects: ten are stopped and
`dac-stack` is running externally with four containers. No stopped project was
started just to manufacture a load scenario.

Measurements below use macOS `proc_pid_rusage` cumulative user/system CPU time,
including separately reported reaped-child command CPU, and resident memory.
Samples are five seconds apart. CPU percentages are relative to one logical
core, so 100% equals one core and values can exceed 100%. They are not whole-
machine utilization percentages. The Docker VM and backends are measured
separately, since they include the actual monitoring workload and other Docker
activity. RSS includes shared mappings; it is not equivalent to private memory.

| App state | Window | Daemon CPU | Command CPU | Combined CPU | Daemon median RSS |
|---|---|---:|---:|---:|---:|
| Before update, closed | 60 s, 16:03:47–16:04:47 CDT | 14.0% | 168.3% | 182.2% | 42.05 MiB |
| After update, closed | 60 s, 16:08:57–16:09:57 CDT | 6.7% | 60.8% | 67.5% | 41.94 MiB |
| After update, visible | 120 s, 16:35:09–16:37:09 CDT | 11.7% | 78.8% | 90.5% | 45.16 MiB |
| After update, hidden | 120 s, 16:45:49–16:47:49 CDT | 9.0% | 69.4% | 78.4% | 45.48 MiB |

The closed-app comparison reduces combined daemon/command CPU by 63.0%.
Daemon physical footprint also fell from a median 30.04 MiB to 26.20 MiB in
these windows. This is not yet a satisfactory CPU budget for a mostly stopped
catalog. The Docker VM measured 49.7% versus 56.0% CPU and about 8.33 GiB RSS;
Docker backends measured 17.4% versus 10.7% CPU and 281–295 MiB RSS. These short
windows do not establish an improvement in Docker VM CPU or memory.

A separate 22.9-second process sample observed at least eighty `lsof` children
and seventeen Docker CLI children. The sample can miss very short-lived
commands and includes one explicit native startup-snapshot verification.
Repeated listener inventory scans are a concrete next profiling target; their
individual CPU contribution has not yet been isolated.

The user confirmed that the native dashboard is visible and working. A
120-second visible-dashboard window measured 11.7% daemon CPU plus 78.8%
child-command CPU (90.5% combined), with 45.16 MiB median daemon RSS. The native
shell measured 0.44% CPU plus 0.31% child-command CPU and 63.04 MiB median RSS.
Five newly created WebKit processes remained alive throughout the window and
measured 2.31% CPU with 180.73 MiB combined median RSS. The shell/WebKit group
therefore used approximately 3.06% CPU and 243.77 MiB RSS. Shared mappings mean
that summed RSS is not a private-memory total. The WebKit cohort is attributed
by its creation after Envarium launch; privileged macOS ownership inspection
was unavailable, so attribution is not independently confirmed by that API.

During measurement the native shell was launched from the repository's debug
app bundle, not the copy in the user's Applications folder. Its requests used
the updated 1.0.0 daemon, which served the updated production web assets. After
the user quit the native app, the debug bundle was also replaced with a staged,
verified copy containing the updated Go sidecar. The original debug bundle is
preserved as `Envarium-debug-before.app` in the private rollback directory.
Both the Applications and debug bundles now report `1.0.0`, commit
`2527d9e78a74-dirty`, and the same build timestamp; both pass deep, strict
signature verification in their existing ad-hoc signing mode. The native Rust
implementation was unchanged. The daemon retained PID 67334 and its original
post-update start time throughout this second bundle replacement. The UI
windows above predate this client-side sidecar replacement; the daemon and
served web assets were already updated during both windows.

The user confirmed that Envarium was hidden with the process still running.
The same native PID, daemon PID, and five WebKit PIDs remained alive throughout
both 120-second windows. Five-second samples measured these native/UI totals:

| Window state | Native shell and sidecar CPU | WebKit cohort CPU | Combined CPU | Combined median RSS |
|---|---:|---:|---:|---:|
| Visible dashboard | 0.75% | 2.31% | 3.06% | 243.77 MiB |
| Hidden, still running | 0.71% | 0.51% | 1.22% | 247.17 MiB |

These are sums of the component RSS medians, not a private-memory accounting.
Hiding reduced WebKit activity but retained its memory, as expected for a
live application. In the hidden window, daemon RSS ranged from 45.44 to
45.49 MiB; WebKit RSS ranged from 185.27 to 189.29 MiB, plus 61.54–61.66 MiB
for the native shell. These short windows do not establish the absence of a
multi-hour memory leak or a bound for dozens of simultaneously running projects.

Docker remained separate from the native application measurement:

| Window state | Docker VM CPU | Docker VM median RSS | Docker backend CPU | Backend median RSS |
|---|---:|---:|---:|---:|
| Visible dashboard | 68.0% | 8,286 MiB | 13.2% | 296 MiB |
| Hidden, still running | 61.2% | 8,343 MiB | 11.7% | 301 MiB |

All eleven runtime states and the same four container IDs were verified again
after the measurements. The updated daemon is still ready on schema 18, with
its original post-update start time; no additional daemon restart was needed.
User-confirmed window state is used because native UI automation is unavailable
in this session. Repeated full TCP inventories in
`internal/runtime/process/identity.go` and the two-second native watcher explain
how stopped projects can continue launching listener-inspection work; the port
registry also performs full inventories. Batching/coalescing those read-side
scans, while keeping lifecycle revalidation fresh, is the next concrete CPU
optimization to evaluate. No further monitoring intervals, resource policy,
project state, or permission settings were changed to improve the numbers.

Raw private samples, workload snapshots, build identity, installation details,
and measurement scripts are retained under `.cache/resource-live-20261016/`.
The consistent SQLite rollback copy occupies about 6.6 GiB; binary and complete
app-bundle rollback copies are retained separately. These are local ignored
artifacts, not part of the public documentation bundle.


## 2026-10-16 — Coalesce native listener observation and remove macOS lsof scans

### Scope and acceptance criteria

This Phase 12 slice addresses idle/many-project CPU overhead while preserving
Phase 6 external-process classification and PID-reuse-resistant ownership.
Observation may reuse TCP binding evidence for two seconds. Startup conflict
checks and port suggestions must still query fresh sockets; managed identity
verification and signaling must never use cached fingerprints. No project or
container lifecycle action is part of performance measurement.

### Implementation and safety evidence

- macOS socket observation now reads the kernel inventory through numeric,
  full-address `netstat -anvlW -p tcp/udp`. It avoids walking every process's
  file descriptors with `lsof`. IPv4, IPv6, dual-stack wildcard, scoped IPv6,
  TCP listeners, and UDP bindings are supported. Unexpected layouts, invalid
  rows, cancellation, and output above 8 MiB return errors rather than a
  successful empty inventory. Other platforms retain gopsutil observation.
- Each native driver owns one TCP observation cache, shared across projects,
  ports, watchers, and read requests. The single snapshot expires after two
  seconds; concurrent refreshes are coalesced. Failed refreshes replace prior
  evidence and are also coalesced, avoiding retry storms. Cancelled requests
  do not populate the cache. The cache has no background polling worker.
- Native startup uses the uncached inspector. Port registry display caching
  remains ten seconds, and suggestions continue to bypass that cache using
  fresh TCP/UDP inventory. Fingerprints are still read directly from the OS,
  including for PIDs returned in a cached socket snapshot.
- No accepted architecture decision changes: these are OS adapter and
  read-observation optimizations under ADR-0002 and ADR-0007. Runtime event
  cadence, ownership, permissions, and workload are intact.

Focused race tests cover cache expiry, concurrent coalescing, failed refresh
replacement/recovery, cancellation, copied results, address parsing, malformed
inventory rejection, output bounds, actual socket opening/closing, TCP/UDP
PID attribution, and rejection of startup over an external process even after
an observed stopped snapshot. Existing reused-PID, process-tree, cancellation,
crash/restart, port suggestion, and ownership suites also pass. The platform
compile check now includes the socket adapter.

The full Go and race suites, vet, pinned golangci-lint, architecture and
repository checks, empty-database migration test, generated contract checks,
and govulncheck passed. govulncheck reports no reachable or imported-package
vulnerabilities (three advisories in required but uncalled modules). No UI,
Rust, schema, or migration code changed in this slice; prior frontend checks
remain applicable.

### Deployment and measurement

The CLI and both native app bundles were updated with a staged Go binary,
version 1.0.0, commit `2527d9e78a74-dirty`. The socket-only build was made at
`2026-10-16T22:53:40Z`; the final build, including the additional background
work below, was made at `2026-10-16T23:04:48Z`. Both complete bundles passed deep, strict ad-hoc signature
verification. Prior binaries/bundles are retained privately for rollback.
The old daemon exited gracefully and the new daemon became ready on schema 18.
All eleven runtime states and the same four container IDs were verified after
restart. The existing database was neither migrated nor restored.

A local command benchmark on the running workload measured 0.008 CPU seconds
for one TCP netstat inventory versus 0.141 seconds for one lsof inventory.
This is an isolated command cost, not a daemon CPU measurement.

Native UI controls remain unavailable in this session and the user is away.
The native application is quit, so new measurements must not be labeled visible
or hidden. Neither prior native UI numbers nor simulated request load establish
native open/hidden performance for this new build. Raw private evidence and
scripts remain in `.cache/resource-live-20261016/`.


The socket-only 120-second window measured 5.49% daemon CPU plus 9.39%
reaped-child CPU (14.88% combined), with 41.27 MiB median daemon RSS and
25.82 MiB physical footprint. A corrected 25 ms child inventory then observed
33 Docker children in 30 seconds and no lsof children; a separate 22-second
classification found ten context-show, eleven context-inspect, and ten Compose
config commands. These are sampled lower bounds, not exact launch counters.
Stack samples also showed repeated manifest resolution through idle log feeds.

Two further fixes were therefore included before the final rebuild:

- Default context selection uses a single `docker context inspect` command,
  which returns the active context name and endpoint together. Read-side
  normalization shares endpoint lookups across projects for one second with
  singleflight and a bounded cache. Cache keys include explicit context and
  Docker selection environment. Lifecycle plans continue to inspect context
  afresh. Compose configuration still expires after fifteen seconds; a read
  can include up to one additional second of cached endpoint metadata.
- When no active Compose container can produce a follow stream, log collection
  waits on project-filtered container lifecycle events instead of reopening
  empty/history streams each second. A fifteen-second timeout bounds recovery
  from missing events and configuration changes. A replay cursor preceding
  container listing covers the list/subscribe race. Snapshot log queries
  remain finite; disconnects remain errors; cancellation closes the event feed.

Additional tests cover idle log waiting/cancellation, event wakeup, unrelated
project filtering, disconnect errors, finite empty log snapshots, shared context
resolution, expiry/error recovery, cancellation, and uncached one-command
active-context resolution. Complete Go and race suites, vet, pinned lint, and
architecture/repository checks passed again. Linux and Windows compile checks
passed with the new socket package included. The final restart again preserved
all eleven project runtime states and the same four running container IDs.


Final build, native app quit, 120 seconds (2026-10-16T23:05:50.417246+00:00 through
2026-10-16T23:07:50.425874+00:00):

| Measurement | Final build |
|---|---:|
| Daemon CPU | 3.18% of one core |
| Reaped child-command CPU | 6.46% of one core |
| Combined daemon/command CPU | 9.64% of one core |
| Median daemon RSS | 40.67 MiB |
| Daemon RSS range | 39.36–40.98 MiB |
| Median physical footprint | 25.97 MiB |

This is 85.7% less combined CPU than the prior installed build's
67.47% closed-app window. It is a short-window comparison with unchanged project
states/container membership, not a controlled constant-load benchmark. Docker's
separate VM used 27.67% CPU and
8380 MiB median RSS; backend processes used
3.28% CPU. Changes in that external workload
cannot be attributed entirely to Envarium changes.

A subsequent thirty-second sample observed eight Docker children, four netstat
children, and no lsof children. Short-lived commands can evade this sample;
static call-path inspection also confirms Darwin socket observation no longer
calls gopsutil's lsof implementation. The installed Applications sidecar's
actual desktop snapshot returned eleven projects and no top-level or project
warnings against the final daemon, with matching build metadata.

The remaining CLI normalization and catalog/manifest observation costs are
measurable. These results do not establish a sub-1%-CPU budget for this real
workload, the native shell/WebKit's new-build overhead, or absence of a
multi-hour leak. The empty-daemon budget is a separate acceptance criterion.


A separate 120-second request-load check executed the installed native sidecar's
actual desktop snapshot every fifteen seconds (eight snapshots). Every snapshot
returned eleven projects with zero warnings; latency ranged from
0.15 to 1.07 seconds.
The daemon measured 4.97% CPU plus
8.40% reaped-child CPU
(13.37% combined),
with 43.31 MiB median RSS and
27.90 MiB physical footprint. Snapshot CLI processes
added 0.36% of one core, measured separately by
cumulative reaped-child CPU of the request driver. Docker VM CPU was
53.41%, separate from these daemon figures.

This is a synthetic native tray request-load check against real projects. The
native window and its WebKit processes were absent throughout both new windows;
it must not be presented as a native visible/hidden comparison or a native app
memory measurement. That requested verification remains blocked by unavailable
native UI controls while the user is away. No permission or preference was
changed to bypass that limitation.


## 2026-10-16 — Remove background CLI normalization and repeated manifest parsing

### Scope and exit criteria

The previous installed build's 9.64% closed-app CPU remained too high. This
Phase 12 slice targets recurring configuration work while retaining Phase 6
identity/lifecycle safety and ADR-0008 manifest precedence. The acceptance
checks are: background observation no longer launches Docker commands; source
changes, read errors and invalid input remain visible; lifecycle plans still
use fresh installed-CLI normalization; memory retention is bounded; the actual
projects/containers remain unchanged; and closed-app CPU is measured again.

### Implementation and compatibility

[ADR-0020](../adr/0020-background-configuration-resolution.md) records the
read/lifecycle distinction before this implementation. Read-only Compose
metadata now uses the official compose-go reference loader, pinned to v2.16.1,
with environment/dotenv handling, ordered file merges, includes/extends,
profile selection and worktree port overlays. No remote resource loaders are
registered. Invalid or unsupported inputs remain errors. The observation
window remains fifteen seconds; monitoring was not slowed to obtain savings.
Lifecycle plans continue to normalize afresh through the installed Compose CLI.

SDK observers now read bounded Docker context metadata files directly rather
than invoking `docker context inspect`. Explicit context, environment selection,
active context switching, default environment connections, TLS material paths,
and refusal of insecure TLS are preserved. No credentials helper is invoked,
and registry credentials are not decoded into an application model.

Catalog effective-manifest resolution now uses a service-owned content memo:
all current source documents are read before lookup, and the exact contents,
root and runtime override determine the key. Edits with unchanged timestamps,
new/deleted overlays, accepted revisions and source read errors are therefore
not hidden behind a timer. Results are independently decoded for callers.
Read/validation errors are not retained. Encoded results are capped at 4 MiB
and 128 entries; oversized results are never cached. No global cache is added.

### Verification

The reference loader was compared with the installed Compose CLI on a fixture
and all six actual Compose projects before deployment. Project names, active
service sets, endpoint metadata and command prefixes match. The check caught
and corrected disabled-profile service handling before installation. Context
selection tests cover explicit/env/config precedence, context switches,
missing/corrupt/oversized metadata, cancellation, relative configuration paths,
and insecure TLS rejection. Loader tests cover dotenv interpolation, file
merges, includes, optional profiles, edits, invalid inputs, containment and
cancellation. Manifest tests cover source changes, caller isolation, errors,
concurrent access and retention limits.

A small-fixture manifest benchmark measured 93.5 µs and 45,410 B/op for fresh
resolution versus 21.7 µs and 3,265 B/op with content memoization. These are
microbenchmarks, not whole-daemon CPU estimates.

Complete Go and race suites, pinned lint, architecture/repository checks,
empty-database migration, generated contracts, Linux/Windows compilation and
govulncheck passed. No reachable/imported-package vulnerability was reported;
three advisories remain in required modules whose affected code is not called.
No UI, Rust, API schema or database migration was changed in this slice.

### Deployment

The CLI and both native app bundles contain version 1.0.0, commit
`2527d9e78a74-dirty`, built at `2026-10-17T03:56:04Z`. Both bundles passed deep,
strict signature verification in their existing ad-hoc mode. Rollback binaries
and complete bundles are retained privately. The prior daemon exited gracefully;
the replacement became ready on unchanged schema 18. Eleven project states and
the same four running container IDs were verified after restart. The native app
is quit during measurement. No build or test runs overlap the measurement.


The first in-process-loader build measured 2.98% daemon CPU
plus 0.37% child CPU (3.34% combined) over 180 seconds,
with 45.56 MiB median RSS. A thirty-second child sample found only
netstat commands and no Docker CLI. The remaining daemon CPU prompted another
stack sample and database/query inspection rather than acceptance of that result.

Further repeated work was removed before the final build:

- Accepted manifests are identified by the current persisted catalog revision.
  That revision is read afresh before memo lookup; an unchanged immutable
  snapshot body is no longer queried/decoded on every observation. Local source
  contents are still reread. Changed revisions load and validate immediately;
  unknown/legacy revisions retain exact-content resolution.
- Catalog lists fetch all project tags once, rather than one query per project
  on each background list cycle. Project order, empty tag arrays, tag order,
  per-project assignment and cancellation are covered by integration tests.
- Health-history pruning formerly scanned 7,833 rows every second, with an
  observed full-table query plan. It now runs at most once per minute, separately
  from the unchanged one-second health scheduler. Errors are reported and retried
  on the maintenance cadence, rather than silently causing an immediate loop.
- Active metric points still persist at the configured interval (ten seconds
  by default). Rollup/pruning housekeeping now runs once per minute, matching
  the finest rollup tier, instead of rerunning the same rollups on every sample.
  Idle maintenance remains fifteen minutes. Downsampled history may lag live/raw
  samples by one minute; health history expiry cleanup may lag by one minute.
  Rollups still complete before any source pruning, preserving recovery/order.

Tests verify that accepted-body reads are avoided only for identical authoritative
revisions, local edits still override them, live samples continue between rollups,
and housekeeping retries/cutoffs remain correct. Full Go/race suites, lint and
architecture checks passed again. This intermediate build timestamp was
`2026-10-17T04:05:38Z`; CLI and both app bundles were replaced and verified again,
and the daemon restarted gracefully on schema 18. All eleven project states and
the same four running container IDs remained unchanged.


A further probe-path review found that scheduled health evaluation still ran
HTTP/TCP/command probes and wrote results for stopped/paused projects, although
`aggregateHealth` immediately discarded those results and returned unknown.
That work is now skipped for confirmed stopped/paused observations. Health reads
also avoid loading history that would be discarded. Starting/running/unknown
observations retain probe behavior, and `WaitRequired` still performs readiness
checks and retries. Tests assert zero evaluator calls, history reads or writes
for stopped/paused projects and preserve startup readiness evaluation.

An intermediate measurement was interrupted when this issue was found; it is
excluded from the final comparison. All Go/race tests, lint, architecture,
formatting and platform checks passed after this fix. An intermediate build
(`2026-10-17T04:09:51Z`) was installed into the CLI and both bundles, with a
graceful daemon restart and unchanged schema 18. Eleven runtime states and
four container IDs were verified unchanged. Its measurement was interrupted
for precise profiling and is also excluded from the final comparison.


### Remaining SQLite and scheduler overhead

A private diagnostic wrapper recorded a 60-second Go CPU profile using the real
CLI bootstrap after a 30-second warmup. It exposed no profiler endpoint and
added no permanent product flag. The profile attributed substantial time to
repeated catalog resolution, SQLite WAL locking and query-cancellation goroutine
wakeups. Cumulative samples included 26% in runtime project resolution, 16% in
`GetProject` and 13% in SQLite file locking; these overlap and must not be summed.
The diagnostic daemon was stopped and the normal installed daemon restored.
An initial helper failed to reap its diagnostic child before checking for exit;
the normal daemon was restored immediately afterward, and project containers
were never stopped.

Project/tag lookup now uses one query. Callers that already fetched a current,
trusted project reuse that DTO for manifest resolution. Each health scheduler
and runtime-watch refresh obtains one current catalog snapshot, retaining
individual project errors and rereading local source files. Direct lifecycle
resolution still reads and checks current trust independently. This is a batch
within each existing scheduler cycle, with no new freshness delay or shared
global cache. Cancellation remains enabled on every database query. Tests cover
batch use, failure isolation, trust revocation, local edits and cancellation.

After these changes, the complete Go and race suites, vet, pinned lint (zero
issues), architecture/repository checks, empty-database migration, generated
contracts and Linux/Windows compilation passed. Govulncheck reported no reachable
or imported-package vulnerabilities (three advisories in unused module code).
The final deployment is version 1.0.0, commit `2527d9e78a74-dirty`, built at
`2026-10-17T04:22:23Z`, SHA-256
`e407ca803eeda095d915ce57c5dc414dea8a3957b15327be768e393cc1e8744f`.
The CLI and both app bundles match; both bundles passed deep strict signature
verification. The daemon exited gracefully and its replacement became ready on
unchanged schema 18. All eleven project states and four running container IDs
were checked again before measurement.


A follow-up 60-second profile after the query batching measured 1.24 CPU seconds
(2.06% of a core while profiling). Catalog resolution's cumulative samples fell
from 0.50 to 0.21 seconds; these short sampled profiles are directional evidence,
not the reported resource measurement. Remaining samples were distributed
among current filesystem reads, Compose parsing, process-run reconciliation,
metric persistence and Go scheduler/GC overhead. The normal installed binary
was restored before repeating the resource measurement. No profiler is enabled
in the installed product.

### Closed-app measurement method

The final measurements use two independent 180-second windows with the native
app quit and no desktop/WebKit process belonging to Envarium. Eleven trusted
projects remain registered: ten stopped and dac-stack externally running its
four containers. No build, test, manual snapshot, profiling or health-validation
requests overlap these windows. macOS `proc_pid_rusage` cumulative user+system
CPU deltas are divided by elapsed monotonic time; 100% means one core. Reaped
child-command CPU is reported separately and included in the combined total.
RSS and physical footprint are sampled every five seconds. Docker's VM/backend
usage is recorded separately, since it includes the actual project workload.
Raw private samples and reproduction scripts remain under
`.cache/resource-live-20261016/`; they are excluded from version control.


| Closed-app measurement | Window A | Window B |
| --- | ---: | ---: |
| Daemon CPU | 2.24% of one core | 2.25% of one core |
| Reaped child-command CPU | 0.32% of one core | 0.31% of one core |
| Median RSS | 43.26 MiB | 44.58 MiB |
| RSS range | 42.37–44.19 MiB | 42.53–44.73 MiB |
| Median physical footprint | 27.62 MiB | 27.64 MiB |
| **Combined CPU** | **2.56% of one core** | **2.57% of one core** |

Combined CPU fell from 9.64% to 2.56–2.57% of one core,
about 73.4% lower. The Docker VM independently consumed
32.52% and 31.93% CPU; its workload is excluded from Envarium's overhead.
The native app was quit throughout. These results establish a repeatable
closed-app improvement for this workload; they do not establish a sub-1% budget,
visible/hidden native-app overhead or absence of a long-running memory leak.
The prior visible/hidden verification gap is unchanged.


After both windows, daemon readiness and schema 18 were rechecked. All eleven
runtime states, four running container IDs, exact start times and zero restart
counts match the original workload. Grafana, Metabase, Prometheus and Uptime
Kuma health/dashboard endpoints returned HTTP 200. Both app bundles passed
strict deep signature verification again; embedded metadata and all Mach-O
payload sections match the installed CLI (ad-hoc signature bytes intentionally
differ). A separate 30-second child sample observed netstat and a process caught
before exec, with no Docker or lsof executable observed. Sampling may miss
short-lived processes and is supporting evidence, not a complete exec trace.
The installed normal daemon remains running, the native app remains quit, and
rollback binaries/bundles are retained privately.


## 2026-10-16 — Target under 1% CPU with the native app quit

### Scope and acceptance checks

The repeated 2.56–2.57% result did not meet the user's resource expectation.
This slice targets under 1% combined daemon and reaped-child CPU on the same
actual workload. Keep the one-second health scheduler, two-second runtime/socket
reconciliation, fifteen-second Compose refresh window, ten-second raw resource
sampling, current source/error handling and fresh lifecycle/identity checks.
Validate native socket evidence against netstat, exact-input reuse against the
installed Compose CLI, caller isolation and query results. Then run the relevant
quality gates, install matching CLI/app sidecars, restart gracefully, recheck
projects and measure the installed normal binary in clean repeated windows.

### Changes and intermediate evidence

- ADR-0021 records direct macOS PCB inventory parsing with the numeric netstat
  compatibility fallback. Known framing/layouts, generation filtering, TCP
  listener state, IPv4/IPv6 addresses and last-owner PIDs are validated. Process
  fingerprints remain independently checked. Live TCP4/TCP6/UDP4/UDP6 parity
  passed. Unknown layouts and API errors fall back, rather than becoming a
  successful empty inventory. No monitoring interval changed.
- Compose observations reuse validated metadata only after exact-content checks
  of all known inputs and the process environment. Local env/label dependencies
  and absent optional/default dotenv files participate in invalidation. Includes,
  extends, aliases, dynamic dependency paths and explicit dotenv modes retain
  full reference loading. Inputs are checked before and after parsing; no source
  contents or environment values are retained. All six real Compose projects
  qualified and matched installed-CLI metadata on both fresh and reused reads.
- Native run history now uses one joined read for up to 200 runs and all their
  identities, preserving run order, identity order, absent identities, project
  boundaries, terminal outcomes and cancellation. Catalog lists also obtain
  projects and tags in one joined read. This removes query/lock/cancellation
  goroutine overhead without disabling query cancellation or caching trust.
- Health scheduling skips projects with no declared probes: those jobs had
  performed runtime reads and discarded their aggregate return value without
  writing any result. Direct health reads still inspect built-in runtime health;
  new/removed declarations are still detected by the unchanged scheduler.
- Effective manifest cache hits return explicit typed copies instead of repeatedly
  decoding JSON. All reference fields, including nested invocation options, are
  covered by an isolation test that also detects newly added fields. The cache
  keeps its 128-entry ceiling and a 4 MiB conservative retention-weight budget
  (eight times serialized size plus 1 KiB per entry), without retaining JSON.

The first candidate measured 1.59% CPU with zero reaped-child CPU. Its 60-second
profile attributed most samples to system calls and Go thread wakeups, with
46.2 MB allocated and nine GC cycles. It was not accepted as the final result.
A temporary two-thread trial of the old binary showed only a small improvement;
it was restored afterward. A later one-thread candidate with no-probe scheduling
measured 1.09%, also above target. Runtime tuning remains experimental until
its benefit and behavior are verified with the complete changes.


ADR-0022 records the final runtime choice: one default Go CPU execution slot for
`envarium daemon`, respecting an explicit GOMAXPROCS override and restoring the
previous setting when the command returns. This is not a 1% CPU limiter; one
slot can still execute at 100% of a core. It reduces worker wakeups in the
I/O-oriented daemon. A barrier-based test with 32 one-second probe declarations
verified four simultaneous I/O waits and completion of every probe with one
slot. Default/override/restoration tests, complete Go/race suites, vet, pinned
lint, architecture/repository/migration checks, generated contracts and
vulnerability checking passed. Linux/Windows and macOS arm64 compilation passed.
The arm64 result is compilation evidence, not a live arm64 measurement. No
frontend, Rust, public schema or migration changes were needed.


The default-runtime candidate measured 0.988% combined CPU over 180 seconds,
with zero child-command CPU. This was treated as intermediate evidence, not the
installed result. A final small optimization proves absence of the optional
manifest parent directory once per resolution, avoiding two repeated missing-file
lookups for seven actual projects. Existing directories still have both file
contents read on every resolution; source creation, deletion, unchanged-mtime
edits and errors remain covered. No metadata-only content cache was introduced.
Full Go/race suites, vet, pinned lint and architecture checks passed again.

### Installation and pre-measurement verification

Installed version 1.0.0, commit `2527d9e78a74-dirty`,
built at `2026-10-17T05:06:35Z`; CLI SHA-256
`c5d04b2470ed39bf596f68e30fc5741f83d02640538a1c7a6ef9482ee038a014`. The CLI and both app sidecars have matching metadata and
Mach-O payload sections; both complete bundles pass strict deep signature
verification in their existing ad-hoc mode. Rollback copies are retained. The
old daemon exited gracefully; replacement PID 90808 became ready on
unchanged schema 18. No project process or container was stopped.

All eleven runtime states and four running container identities, exact start
times and zero restart counts match the original workload. All four service
health/dashboard endpoints returned HTTP 200. A request check returned eleven
projects without warnings in each of four native snapshot requests: 3.22 seconds
for the initial cold request and 0.047–0.050 seconds for subsequent warm requests.
These requests ran before the clean CPU windows and are not a native GUI test.

### Installed closed-app measurement

After a thirty-second warmup, run two consecutive five-minute windows using
macOS cumulative process and reaped-child CPU time, with 100% equal to one core.
Sample RSS and physical footprint every five seconds. The native app remains
quit; ten projects remain stopped and dac-stack's four containers remain active.
No build, test, profiler, manual snapshot or validation request overlaps the
measurement. Docker VM/backend CPU is separate from Envarium's overhead.
Afterward, verify persisted raw sample availability and timing, snapshot latency,
readiness, container continuity and app signatures again.


### Five-minute periodic diagnostic spike discovered and corrected

The first installed five-minute window did **not** pass: 1.10% daemon CPU plus
0.67% reaped-child CPU, 1.77% combined. The spike occurred at
`2026-10-17T05:12:27Z–05:12:38Z`, exactly five minutes after daemon startup.
The second window was interrupted for investigation and is excluded. Project
states and running containers were rechecked and remained unchanged.

The automation scheduler was collecting full diagnoses for every project every
five minutes, even with zero configured recipes. This included fresh Git commands
and cleanup recommendations. Skipping empty recipes would have removed automatic
notifications, so that approach was rejected. ADR-0023 records a scheduled
evidence path that keeps current runtime, health, logs, ports, resource warnings,
operations, configuration and approved actions, while omitting Git and cleanup
enrichment used only by non-notifying recommendations. Scheduled diagnoses are
explicitly labeled and retain notification-producing deterministic hypotheses.
Manual diagnoses and explicit recipe evaluation still collect full evidence.
The five-minute cadence and enabled-recipe checks, cooldowns and limits remain.

Parity tests cover every notification and supported recipe trigger, including
matching summaries, evidence and suggested actions. Adapter tests verify that
scheduled collection omits Git/cleanup while manual collection retains them.
Due recipe dispatch, cooldowns, empty recipes, cancellation and collector errors
are tested. Complete Go/race suites, vet, pinned lint and architecture/platform
checks passed again. Generated contracts remain unchanged.

The completed build is version 1.0.0, commit
`2527d9e78a74-dirty`, built at `2026-10-17T05:22:04Z`,
CLI SHA-256 `496e12a48b6a5c892447cfbc4810f3465f98f0ff3c7dc8e86a0a507e88906877`. Both bundles and the CLI were replaced again,
with strict deep signature verification and a graceful daemon restart on schema
18. Rollback copies remain private. New measurement windows are six minutes each,
so both include a scheduled diagnostic pass and normal maintenance jobs.


### Completed installed measurements and acceptance evidence

Both consecutive six-minute windows passed the under-1% target with the normal
installed daemon, PID 95586, and the native app quit. Percentages below use one
CPU core as 100% and include reaped-child CPU, which was zero in both windows.

| Window (UTC, 2026-10-17) | Combined CPU | RSS median (range), MiB | Physical footprint median, MiB |
| --- | ---: | ---: | ---: |
| 05:25:13–05:31:13 | 0.965% | 40.14 (40.07–41.80) | 26.21 |
| 05:31:13–05:37:13 | 0.996% | 42.16 (41.80–42.20) | 27.09 |

The duration-weighted twelve-minute result is **0.980%**, approximately 62% below
the previous 2.56% result. Both windows include the five-minute scheduled
diagnostic job, which produced labeled diagnoses for all eleven projects while
retaining notification checks. No build, test, profiler or manual application
request overlapped these final windows. Docker VM CPU was approximately 38–39%
and is excluded from the Envarium overhead measurement. These measurements are
workload-specific evidence with a narrow margin below 1%, not a CPU ceiling or
an hours-long memory-leak assessment.

All five resource series (project aggregate plus four containers) contain 36
raw points in each six-minute window, averaging 10.000 seconds between points.
Probe intervals, probe worker count, runtime reconciliation, raw resource
interval and the scheduled notification interval were not lengthened. Tests
verify notification hypothesis/evidence/action parity, all supported recipe
triggers, current socket evidence, manifest invalidation and concurrent probes.

An initial validation assumption of exact sample deadlines and universally
available Docker metrics failed. The measured timestamps have paired short/long
gaps of approximately 4–16 seconds; collection is not a hard ten-second deadline.
A separate earlier window of the previously installed normal binary
(04:57:10–05:01:15 UTC) also had jitter, including a 14.05-second maximum gap.
Availability matches that baseline exactly: aggregate, Prometheus and Uptime Kuma
have CPU and memory in every point; Grafana and Metabase have neither metric
available in either baseline or final windows. The final check therefore verifies
full sample counts, average cadence and unchanged per-service availability, and
records actual gaps without claiming strict deadlines or complete Docker stats.
The existing missing metrics remain a limitation; no missing value is represented
as zero or claimed as newly fixed by this slice.

Post-window verification passed: all eleven project states match the starting
workload; the same four running containers retain their exact start times and
zero restart counts; all four health/dashboard endpoints return HTTP 200. CLI
and both signed app sidecars have matching version and executable payloads.
Four native snapshot requests return eleven projects with no reported warnings:
3.209 seconds for the cold request, then 0.041 seconds each when warm. Snapshot
warnings do not imply complete per-container resource availability. No native
GUI/open/hidden measurement was performed in this latest closed-app slice.

Complete Go and race suites, vet, pinned lint, architecture/repository checks,
migrations, generated contracts and platform compilation passed for the completed
source. Vulnerability checking found no reachable/imported vulnerable package;
three advisories apply only to unused modules. Private raw measurements,
installation/rollback records and validation outputs are retained under
`.cache/resource-live-20261016/` and are not published with project data.

## 2026-10-17 — Native dashboard visible and hidden with active local projects

### Scope and measurement protocol

The user authorized starting Family Hub and LeagueSaga locally, and opening and
hiding the native app without user actions. This is a Phase 12 overhead
measurement of the existing installed build, not another product change. Keep
the existing daemon and running projects intact. Preview and start the two
trusted registered Compose definitions through Envarium, verify readiness, and
collect comparable six-minute closed, visible and hidden windows. Each window
must include normal periodic jobs. Exclude startup, builds, tests and manual
daemon requests from the measurement windows, and recheck container identities,
start times, restart counts, project states and monitoring afterward.

Both startup operations succeeded. Family Hub has five running containers and
LeagueSaga has seven; dac-stack's existing four continue running. Optional
marketing profiles are not enabled. Both applications' declared API readiness
endpoints and frontends returned HTTP 200 before measurement. These are running
local services at rest, not a synthetic request or application load test.

Use macOS cumulative process CPU time, including reaped-child CPU, with 100%
equal to one core. Sample RSS and charged physical footprint every five seconds.
Report the daemon, native app, attributable helpers and combined overhead
separately. Sum process RSS explicitly; shared resident pages can appear in more
than one process. Docker VM CPU/memory is recorded separately and excluded from
Envarium's total.

A private native helper uses AppKit to locate the exact installed bundle,
activate/hide it and read its state; CoreGraphics confirms whether its windows
are on screen. State is rechecked every thirty seconds. Native helpers are
attributed using the app's macOS resource coalition, rather than including other
applications' WebKit processes. The coalition layout is checked at runtime
against the returned size; its interface is described in
[Apple's XNU process-info source](https://github.com/apple/darwin-xnu/blob/main/bsd/sys/proc_info.h).
Measurements require a stable helper cohort and readable counters throughout.
No permanent instrumentation, profiler or daemon restart is introduced.

The installed app was opened directly and its dashboard screenshot inspected.
After its initial loading state, it showed three running projects and sixteen
running services. An existing partial Docker-storage warning remains visible;
this does not prevent current runtime observations from loading. Private screen
captures and raw process data remain in the ignored local evidence directory.

### Results

The normal installed build remains version 1.0.0, commit
`2527d9e78a74-dirty`, built at `2026-10-17T05:22:04Z`, on schema 18.
Daemon PID 95586 was not restarted. The same native app PID 8526 and five
coalition-attributed WebKit helpers remained alive throughout both app windows.
Window visibility/hidden state checks passed every thirty seconds. The visible
window starts approximately one minute after opening; the hidden window starts
after allowing the hide transition and in-flight work to settle.

| State | UTC interval | Daemon CPU | Daemon child CPU | Desktop/helper CPU including children | Combined CPU |
| --- | --- | ---: | ---: | ---: | ---: |
| App quit | 13:40:59–13:46:59 | 1.519% | 0.000% | 0.000% | **1.519%** |
| Dashboard visible | 13:48:00–13:54:00 | 10.112% | 51.472% | 5.238% | **66.822%** |
| App hidden | 13:54:45–14:00:45 | 10.093% | 52.477% | 4.684% | **67.253%** |

Each interval is six minutes. The combined CPU includes short-lived reaped
commands, rather than only the long-lived process CPU visible at one instant.
The native app alone plus its helpers is a small part of the increase; most CPU
is charged to daemon child commands. Hiding the app did not materially reduce
the complete-window average. These are sequential single windows, with normal
workload variation; they do not establish a precise causal difference between
the two app states.

| State | Daemon median RSS, MiB | Desktop/helpers median RSS, MiB | Combined median RSS (range), MiB | Combined median physical footprint, MiB |
| --- | ---: | ---: | ---: | ---: |
| App quit | 41.86 | 0.00 | 41.86 (41.83–41.86) | 28.02 |
| Dashboard visible | 42.59 | 208.31 | 250.88 (228.69–268.40) | 192.29 |
| App hidden | 42.95 | 232.41 | 275.34 (264.05–279.77) | 186.96 |

Memory rows cover the daemon, native app and five persistent WebKit helpers.
Short-lived command memory/peaks are not captured by these persistent-process
RSS/footprint totals, even though their CPU is included. RSS grew over the app
windows, while the hidden-window median charged footprint was lower. These
short measurements do not establish a memory leak. Docker VM CPU averaged
approximately 219%, 277% and 271% respectively and is excluded, along with its
memory, from Envarium overhead.

### Hidden-state child-command evidence

After the clean CPU windows, a separate thirty-second inventory sampled only
the daemon's child PIDs every 25 milliseconds, with no manual snapshot request.
It observed 336 distinct Git child PIDs: 237 first seen at `/usr/bin/git` and
99 at the Command Line Tools Git executable. Nine other children were first
seen at the Envarium executable, which can occur before a forked child execs.
Each PID is counted once; the sample can miss very short commands and does not
attribute every CPU second to an executable. It nevertheless confirms recurring
Git work while the same app is hidden. The dashboard's source requests fresh
Git observations for trusted projects on its fifteen-second snapshot refresh;
this is a concrete path to investigate alongside redundant/overlapping calls.
No product optimization or rebuild was performed in this measurement slice.

### Workload continuity and monitoring limits

Post-window checks passed for all eleven project states and all sixteen running
container identities, exact start times and zero restart counts. Both local
applications' API readiness endpoints and frontends, plus the four existing
stack health/dashboard endpoints, returned HTTP 200. Envarium reports ready with
the same daemon PID/build. Family Hub and LeagueSaga remain running, and the
native app remains hidden with no on-screen window.

The resource timing check found a real limitation under the larger workload.
Its initial nominal-count assertion failed and was retained as investigation
evidence; actual values were then recorded explicitly. Each of the nineteen raw
series (three project aggregates plus sixteen containers) has these counts:

| State | Raw points per series | Nominal points at ten seconds | Largest observed gap |
| --- | ---: | ---: | ---: |
| App quit | 34 | 36 | 39.135 s |
| Dashboard visible | 35 | 36 | 33.746 s |
| App hidden | 35 | 36 | 26.301 s |

Several container CPU/memory series contain unavailable values in all three
states. Full per-service availability and strict sample deadlines are therefore
not verified. These gaps occur in the closed baseline too; the smaller earlier
four-container workload is not an adequate accuracy comparison for this new
sixteen-container load. Every window contains labeled scheduled diagnoses for
all eleven projects, with the scheduled notification hypothesis filter intact.

The measurement request is complete. The native dashboard remains a major CPU
issue in both visible and hidden states, and the larger-workload sampling gaps
need investigation. The earlier under-1% closed-app result applies to its
documented four-container workload, not all workloads or app states. Private
evidence files are `native-loaded-{closed,visible,hidden}.json`,
`native-loaded-hidden-child-commands.json`,
`native-loaded-pre-measurement.json` and `native-loaded-verification.json`
under the ignored `.cache/resource-live-20261016/` directory. Documentation
indexes were regenerated and site validation and whitespace checks passed.

## 2026-10-17 — Tune the active native dashboard and reliable sampling

The user authorized iterative fixes, rebuilds, installation, native app control
and measurements without further questions. ADR-0024 maps this slice to Phase
12's overhead and resource-accuracy criteria, with Git freshness from Phase 8
and presentation scheduling from Phase 9. Initial acceptance targets are under
1% closed CPU, hidden within 0.5 percentage points of the matching closed
baseline, under 5% focused visible CPU, complete service metrics and consistent
ten-second raw sampling. Count child CPU and inspect longer memory behavior.
Keep all sixteen real containers running throughout.

### Implementation and first install

- Runtime events had invalidated the complete dashboard every two seconds,
  repeatedly loading unrelated Git/actions for all projects even while hidden.
  Dashboard observations now have independent shared keys. Runtime/health events
  invalidate only their corresponding facts. Resource cards read the sampled
  aggregate overview once rather than issuing direct driver-stat requests for
  every project. Git and action refresh retain their independent cadence.
- Presentation activity uses browser visibility plus the thin native shell's
  focus fact. Hidden or unfocused desktop presentation suspends UI queries and
  the UI event connection, cancels cancellable requests, and refreshes/resumes
  from the last sequence on return. Native tray observation, notifications,
  daemon monitoring and project processes continue. Native state carries only
  a boolean presentation fact; Rust contains no new product policy.
- Trusted Git reads coalesce in-flight work with four concurrent read slots.
  Working-tree status, remotes, worktree inventory and merge/rebase markers stay
  fresh. Validated Git directory discovery and immutable commit metadata are
  reused in a 128-entry cache; retained commit strings are bounded. Replacements,
  grafts and linked-worktree metadata fall back to fresh commit reads. Caller
  result copies, cancellation, trust, same-mtime edits and new HEADs are tested.
- Storage scans move to an independently canceled five-minute worker. The metric
  collector and overview use its explicitly pending/partial snapshot; explicit
  storage/cleanup requests keep their inventory behavior. A running sampler
  cannot be rescheduled by overview requests, which disclose stale observations.
- Container stats use up to eight workers per project and three-second individual
  deadlines inside the unchanged four-second project deadline. Results reach
  sinks serially in stable order. This retains Docker's previous CPU sample
  semantics while preventing early containers from starving later services.

The first installed candidate is version 1.0.0, commit
`2527d9e78a74-dirty`, built at `2026-10-17T16:17:41Z`, CLI SHA-256
`de02c8cfc203880d7fef2e036c1c94de8b4d71500a2f639812b5b91f6824a29c`.
CLI and both complete native bundles were staged, ad-hoc signed and strictly
verified. The native app quit gracefully before replacement. Daemon PID 95586
exited gracefully and replacement PID 76431 became ready on schema 18.
Rollback copies are retained; no database restore was performed.

All sixteen container identities, exact start times and zero restart counts
match the earlier workload. Family Hub and LeagueSaga are now classified as
`running_external` after reconciliation by the new daemon; their actual
containers were not restarted. The initial strict project-state comparison
flagged this origin-label change, so it is recorded separately from container
continuity and active/stopped status. The first eight sampling cycles contain
CPU and memory for all sixteen services and all three aggregates, including
those previously unavailable.

Full Go/race suites, vet, pinned Go lint, architecture/repository checks,
migrations, generated contracts, Linux/Windows compilation, native tests,
frontend unit/coverage/type/lint checks, five browser end-to-end tests and
fourteen visual tests passed during preparation. Final native lint and clean
performance windows are still pending; this is not yet an accepted result.

### First-candidate observations and second iteration

The clean first-candidate closed window measured 1.615% CPU over six minutes,
zero reaped-child CPU and 44.02 MiB median RSS (29.28 MiB charged footprint).
This did not meet the closed target. Sampling did pass: every one of the nineteen
series had 36 raw points with CPU and memory available throughout. This window
is retained as intermediate evidence, not presented as a final accepted result.

A temporary wrapper around the normal daemon entry point then collected a
sixty-second CPU profile after a thirty-second warmup and restored the normal
installed binary automatically. No profiling endpoint was added. It attributed
substantial work to runtime reconciliation, SQLite and HTTP I/O; the completed
closed window's journal contained 516 runtime observations, all Docker exec
lifecycle chatter (172 each of create/start/die). The Compose watcher now skips
only these known non-state-changing events. Tests preserve health-status,
lifecycle and unknown future actions. Monitoring intervals and probes remain
unchanged. The normal daemon and all containers were verified after profiling.

The second candidate is built at `2026-10-17T16:30:13Z`, with the same version,
commit and schema. It changes only this event filter beyond the first candidate;
focused race tests and pinned lint passed before installation. Further CPU,
visibility and sampling measurements remain required.

### Third iteration: bound presentation and history work

The second candidate's probes were still above budget: 1.296% closed over
90 seconds, 15.301% focused/visible over two minutes, and 2.429% hidden over
two minutes. The hidden probe was exploratory (a read-only child-process sampler
overlapped its tail); final windows will run without instrumentation beyond the
measurement harness. No performance target is claimed from these probes.

The dashboard now requests fresh working-tree/branch status through a dedicated
documented endpoint, leaving full commit/remote/worktree observation unchanged
for detail, CLI and MCP. Warm status reads need one Git process per project
instead of three; status, trust, cancellation and enrichment boundaries are
tested. Card view objects survive unrelated fetch-progress changes. The native
tray compares only rendered menu inputs before replacing an unchanged menu;
notification tracking and snapshot polling continue every fifteen seconds.

Exact manifest resolution now carries its canonical declaration hash internally,
avoiding repeated JSON serialization by background observers. SQLite footprint
queries separate count and indexed minimum seeks; rollup watermarks are resolved
once per service. Tests cover independent watermarks, newly observed services,
late points, repeat maintenance and exact oldest/count results. Schema remains
18. API clients are regenerated rather than edited. The generation check's Git
comparison reports the intentional new endpoint; a second complete generation
reproduces every generated artifact byte for byte.

Third-candidate CLI SHA-256:
a83d529c029291a60e47f261e0319c21f603d7483693d19a2655d6e7732fc035.
Version 1.0.0, dirty commit 2527d9e78a74, build 2026-10-17T16:47:26Z.
Both native bundles were rebuilt, staged, signed and verified before replacement.
Daemon 80867 exited gracefully; normal replacement 92960 became ready. Rollback
binaries/bundles are retained and no database restoration occurred.

Go tests/race, vet, pinned Go lint (zero issues), architecture/repository checks,
61 frontend tests, frontend types/lint, sixteen native tests, native clippy and
five browser end-to-end flows pass. One visual test exposed outdated mocks for
the new Git/status and sampled-resource endpoints; fixtures are corrected and
the visual suite is being rechecked. Final resource measurements remain pending.

The third candidate's closed three-minute probe measured 1.124% CPU and
46.74 MiB median RSS. This still misses the target. A fourth iteration retains
fresh one-second health source checks while avoiding full manifest copies:
bounded runtime projections are keyed by current declaration hash, root and
slug. Trust, local source bytes, accepted revision and environment overlays remain
fresh; nested caller-owned values are copied. New tests cover caller mutation,
retention bounds and source edits/errors through identity-only resolution.

The fourth candidate was built at 2026-10-17T16:54:05Z, CLI SHA-256
5b28e5a53f9b0eca78603b3a3b01fbfff0f70aa7aedf4893d006e089323ad061.
Daemon 92960 exited gracefully and normal replacement 95521 became ready on
schema 18. Full Go/race/vet, pinned lint, architecture and the complete fourteen
visual checks pass. Closed/visible/hidden probes remain in progress.

Visual review also identified a display accuracy detail: unavailable resource
samples must show an unknown value, not zero. Known stopped projects can show
zero, while disconnected or partial observations retain their warning. The final
frontend candidate includes explicit CPU/memory availability rendering and
tests for both cases.

### Bounded transport and native framing

The projected candidate's closed two-minute probe remained at 1.145% CPU.
A second sixty-second CPU profile measured 1.11%, 46.15 MB allocated, seven GC
cycles and 118 goroutines. It identified repeated SQLite preparation and metric
batch work. Read plans now have a per-database 64-statement/64-KiB bound; query
parameters/results remain fresh, preparation failures are not retained, and
metric insert/upsert statements are prepared once per transaction.

That candidate was followed by a prepared-read build at 2026-10-17T17:02:17Z
(SHA-256 98a184d219ad08247a943be15d3b4ac8a0df39cca60a199cb696b395abd0f472).
The focused visible probe still used 9.864% CPU: 2.124% daemon own, 3.015% Git
children, 1.026% native own/children and 3.698% attributable WebKit helpers.
Combined median RSS was 294.0 MiB. A hidden exploratory probe immediately after
hiding measured 2.471%; it includes the transition and the five-minute disk
enrichment boundary and is not a final steady-state result.

The remaining work now addresses transport costs directly:
- An application facade reads one observation kind per bounded request
  (runtime, Git status or actions), with at most 64 IDs/four concurrent reads,
  ordered results, generic per-project warnings and cancellation. Underlying
  services retain current trust and source checks. Dashboard queries remain
  independent by kind; runtime events cannot invalidate Git.
- Connected idle operation presentation uses event invalidation and a fifteen-
  second safety poll. Active operations and disconnected presentation retain
  1.5-second polling. An event replay gap invalidates current facts explicitly.
- Apple's Git toolchain launcher is resolved once per observer; custom PATH
  selection, cancellation and removal fallback are tested. A local ten-read
  comparison on this repository measured 0.721 child CPU seconds via the launcher
  versus 0.483 through its selected Git executable.
- Native commands request compact JSON with the same stable envelope. The shell
  plugin frames stdout by line; compact output removes hundreds of line events
  from each snapshot without dropping data.
- Host capacity reads share the resource worker's disk summary. They no longer
  start a second full Docker disk scan; unknown sizes remain unknown and Engine
  connectivity can still be checked with a cheap ping.

The new frontend suite passes 67 tests, native tests pass, Go/race/vet and
architecture checks pass, pinned lint has zero issues, and five browser flows
plus fourteen visual checks pass. The next candidate is built at
2026-10-17T17:21:46Z. Final installation, workload continuity, sampling accuracy,
CPU and longer memory windows are still required.

### Remove remaining health-history and diagnostic-cycle overhead

The clean six-minute batched-candidate closed window still measured 1.124% CPU,
zero reaped-child CPU and 47.99 MiB median RSS (32.38 MiB charged footprint).
Its two-minute focused native probe improved to 5.950% combined CPU and
283.58 MiB median RSS. Native own CPU fell to 0.154%; the probe still overlaps
a scheduled cycle and does not satisfy the final acceptance target.

A read-only, transaction-consistent comparison across all eleven real projects
showed identical latest health results while reducing total query time from
37.67 ms to 0.95 ms. The store now enumerates indexed check scopes and seeks one
latest timestamp/ID row per scope, preserving ties and cross-project isolation.
Worktree-registry and log-dedup reads reuse bounded prepared statements.

The diagnostic scheduler now constructs an explicit collector scope per cycle.
Global port/resource snapshots are read once; project trust, logs, health and
approved actions remain current per project. Captured source failures remain
visible partial warnings, and interactive collection is unchanged. Tests verify
attribution, partial failures, cancellation and that the scope does not leak into
independent collection.

The same one-core I/O execution budget now applies at the CLI entry point as
well as daemon entry, preserving explicit overrides and leaving subprocess
environment variables untouched. This avoids constraining monitored project
processes. A five-snapshot client comparison reduced child CPU from 0.278 to
0.228 seconds. A separate Git index-thread experiment was slower and was not
adopted; Git retains its existing index/preload behavior.

The candidate is built at 2026-10-17T17:48:36Z. Full Go/race/vet, focused SQLite
and diagnostic tests, and pinned Go lint pass. Native Rust is unchanged from the
validated compact-output bundle. Final installation and complete performance,
memory and sampling validation remain pending.

The installed cycle candidate's complete six-minute closed window measured
1.063% CPU and 48.79 MiB median RSS (33.39 MiB charged footprint). This still
misses the target. A new three-minute profile attributes 170 ms to metric
maintenance and shows repeated SQLite page reads/locking; its sampled total
was 0.88% CPU, which does not replace the process-counter measurement.

The final rollup plan enumerates source series and uses the existing history
index for each target watermark and incremental source range. A read-only,
transaction-consistent live comparison returned identical complete rows: raw
rollup reads fell from roughly 22–26 ms to 4 ms, and minute rollup reads from
13 ms to 7 ms. Earlier scope-only experiments lacked the explicit indexed range
plan and were not adopted. New tests cover identical service names in different
projects, target-only retained history, repeat maintenance and cancellation;
existing late-sample and retention tests remain required. Final installed
measurements are still pending.

The indexed-rollup candidate (build 2026-10-17T18:11:57Z) passed full Go/race,
vet, lint and architecture checks, but its closed six-minute run was still
1.028% CPU with 43.52 MiB median RSS. The next slice removes discarded Docker
server-version requests from metric/log/event connections; runtime inspection
continues fetching current server metadata and every connection still negotiates
the API version. An actual local Engine HTTP fixture verifies one negotiation
request and no unused metadata requests. Footprint aggregates now share one
SQLite read snapshot while retaining exact counts and indexed oldest seeks.

The lean candidate (build 2026-10-17T18:20:43Z, SHA-256
67a8b2cc8d55a05784387d90aeb81b8e139bbc1dcbe1186571e61e5d08175c38)
measured 0.994% closed CPU and 44.69 MiB median RSS over six minutes. Its
focused native six-minute window measured 4.983% combined CPU: daemon own
1.903%, daemon children 1.434%, native own/children 0.508%, and attributable
WebKit helpers 1.138%. Median combined RSS was 290.81 MiB; charged footprint
254.44 MiB. These windows include scheduled work, but are provisional because
review found the new runtime batch omitted the single endpoint's health-derived
degraded/starting state reconciliation.

The correction moves that reconciliation into an application read shared by
both endpoints. An additive `status` batch returns the same runtime and health
facts once per project; native snapshots replace their two-HTTP-requests-per-
project loop with chunks of at most 64 IDs. Existing trust checks, partial
warnings and fifteen-second tray/notification polling remain. Older daemons
retain the single-endpoint fallback, and denied/unavailable batches are not
retried through another path. Tests cover health-derived states, partial source
failures, cancellation, 130 trusted projects across 64/64/2 chunks, untrusted
projects and legacy fallback. New generated clients and final installed CPU,
memory and sampling evidence remain pending.

The status-batch build at 2026-10-17T18:45:44Z passes Go/race/vet, lint,
architecture checks, 67 frontend tests, types/build, five browser flows and
fourteen visual checks. All 24 generated artifacts reproduce byte for byte.
Installed CLI and sidecar payloads/signatures match. Live verification shows
one status-batch request, no per-project runtime/health HTTP requests, and
matching runtime/health/container identities for all eleven projects. Its closed
six-minute measurement is still 1.020% CPU, 44.77 MiB median RSS, so a small
additional reduction is required.

The next candidate reuses up to eight idle Engine SDK clients across explicit
local Unix endpoint identities. Clients are leased independently and closed on
driver shutdown; failed negotiation is discarded, and environment/TLS clients
remain fresh. Every lease still performs API negotiation and every runtime
status read still fetches current server metadata and container state. A real
local HTTP fixture verifies two current negotiations with a changed API version
over one physical socket, and concurrent/bound/failure/shutdown race tests pass.
The measurement harness now includes live-child CPU at window boundaries in
addition to reaped child CPU; snapshots racing child reaping are retried.

The installed pooled-Engine candidate (build 2026-10-17T18:57:26Z, SHA-256
0af3bb4f48e036299ad0e623dbdc88256617e921d27adb0d456b7fa5d08de84f)
passes Go/race/vet, pinned lint, architecture and five runtime browser flows.
Its closed six-minute window is 1.012% CPU / 44.40 MiB median RSS. Its native
visible window is 5.512% combined CPU / 288.68 MiB median RSS. A separate
post-window WebKit sample finds periodic query/microtask and paint work; it
does not establish a continuous animation problem. The first hidden six-minute
window is 1.605% / 296.40 MiB RSS and 182.60 MiB charged footprint, with two
scheduled cycles falling inside that window. A further twelve-minute hidden
observation is in progress; targets remain provisional.

The final cleanup uses shallow query observation for immutable server payloads,
disables inactive dashboard/host/operation presentation timers, skips empty
global invalidation timers, and contains card layout/paint invalidation without
changing card geometry. Native operation-notice deduplication now has a 1,024-ID
bound, with a 5,000-operation regression test. Fixed health write plans share
the existing statement budget. Log fingerprinting returns its exact encoded
byte count so segment-rotation estimation no longer serializes the same entry
twice. Tests and final rebuilt/installed measurements are pending; all monitoring
and native notification intervals remain unchanged.

The pooled candidate's twelve-minute extension stopped after approximately
10.5 minutes when a previously idle WebContent helper exited. The native main
process, daemon and projects remained running. The attempt is incomplete and
is not counted as a successful twelve-minute window. Observed RSS stayed near
297 MiB before retirement. Final long-run accounting now uses the native
resource coalition's cumulative CPU ledger, including exited tasks, plus
separate daemon/child counters; instantaneous memory still uses live process
samples. A synthetic retirement test verifies that exited CPU stays counted
while released memory is removed. The primary implementation reference is
[Apple's resource-coalition accounting](https://github.com/apple-oss-distributions/xnu/blob/main/osfmk/kern/coalition.c).


### Final installed resource acceptance — 2026-10-17

This final result supersedes the provisional candidate measurements above. The
installed build is `1.0.0`, commit `2527d9e78a74-dirty`, built at
`2026-10-17T19:50:12Z`; CLI SHA-256 is
`3dc4d27c9e5c1bcd1bd97fcf4cd0b90d15ba9eab887c4d294392b561e7998eaf`.
The CLI and both native app sidecars have identical Mach-O payload sections;
both bundles pass deep/strict signature verification. The native build remains
the existing development build. The daemon restarted gracefully from PID 40782
to PID 56888, without restoring or changing the schema-18 database. Its ready
metadata matches the installed binary.

Family Hub, LeagueSaga and dac-stack remained running throughout. All eleven
registered projects retain their activity state, and all sixteen containers
retain their exact IDs, start times, restart counts and running status. Eight
local HTTP/readiness endpoints return 200. A current compact native snapshot
uses one status-batch request and no per-project runtime/health requests; its
runtime and health results match independent current application reads for all
eleven projects.

CPU below is the time-window average as a percentage of one logical core.
Totals include the daemon, its child commands, native app, its CLI commands and
WebKit helpers; project workloads and measurement tools are excluded. RAM is
the median sum of sampled live-process RSS, with charged footprint shown
separately. No builds, tests or profilers ran during these final windows.

| Native app state | Duration | Combined CPU | Median RSS | Median charged footprint |
| --- | --- | --- | --- | --- |
| Closed | 10 minutes | 0.980% | 47.31 MiB | 32.06 MiB |
| Dashboard visible and focused | 10 minutes | 4.876% | 285.68 MiB | 217.87 MiB |
| Hidden, still running | 20 minutes | 1.503% | 286.19 MiB | 133.60 MiB |

The primary closed target below 1% and visible target below 5% are met in these
windows. Hidden overhead is 0.522 percentage points, slightly above the 0.5-point
stretch budget by 0.022 points; this is not recorded as a full three-budget
pass. These are measured averages for this workload, not CPU caps or guarantees
across machines and workloads. No monitoring or notification interval was
lengthened to obtain them.

Visible CPU comprises 1.942% daemon own time, 1.434% daemon child commands and
1.500% native coalition time. Hidden comprises 1.182% daemon and 0.321% native
coalition time. Closed daemon child CPU was zero in this window. Final native
CPU accounting uses the OS cumulative resource-coalition ledger so normal
helper retirement does not erase prior CPU. A two-minute counter calibration
agrees with the independent process-counter method within 0.024 percentage
points. The shorter initial 30-second calibration differed by 0.056 points and
was not accepted as the calibration evidence.

The hidden run completed despite an idle WebContent helper retiring normally.
Its first-five-minute median RSS was 286.14 MiB; the last-five-minute median was
265.64 MiB. Corresponding charged footprints were 142.28 and 125.38 MiB. No
sustained upward memory trend appeared in this twenty-minute observation;
this does not establish absence of leaks over hours or days. Five-second RAM
sampling can miss short-lived child peaks, and summed RSS can include shared
pages. The OS CPU ledger includes exited native tasks regardless of whether
they were present at a RAM sample.

Sampling acceptance passes for all nineteen series: sixteen service series
plus three project aggregates. Every series has sixty raw samples in each
ten-minute window and 120 in the twenty-minute hidden window, with CPU and
memory available on every point. Mean intervals are approximately ten seconds.
Maximum gaps are 10.071 seconds closed, 11.027 seconds visible and 10.212
seconds hidden, below the fifteen-second bound. Scheduled diagnoses run for
all eleven projects in every window, retaining the explicit scheduled-Git
omission warning and all notification-eligible hypotheses.

Final validation passes full Go tests, race tests, vet, pinned Go lint (zero
issues), formatting, architecture and repository checks; 68 frontend unit tests
across 29 files, lint, type checking and production build; five browser flows
and fourteen visual checks; and seventeen native Rust tests, formatting and
Clippy. The native bundle and disk image rebuild succeeds. Empty-database
migrations and Linux/Windows compile smoke checks pass; those cross-platform
checks do not execute tests on those operating systems. Govulncheck reports
zero reachable vulnerabilities, with three advisories confined to required
modules whose affected code is not imported/called. The 24 generated contract
artifacts reproduce byte for byte from the updated schema. The Git-baseline
`generate-check` is not recorded as clean because it includes the intentional
new contract changes in this uncommitted workspace.

Raw reports, per-process samples, parity checks, workload/metric verification,
build/test logs and the original database backup remain in the private ignored
`.cache/resource-live-20261016` evidence directory. The final reports are
`retention-final-{closed,visible,hidden}.json`,
`retention-final-verification.json`, `retention-final-live-parity.json`,
`retention-binary-verification.json`, `retention-counter-validation.json` and
`retention-final-memory-trend.json`. The temporary awake assertion was removed.
After measurement the verified native app was reopened and left running hidden;
the daemon and project containers remain running.
