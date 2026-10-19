---
title: Operational experience and resource refresh
description: Execution plan and verification ledger for performance, source control, ports and page usability.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-18
---


Requested scope: consistently low monitoring overhead, usable Git and worktree
inspection/switching, project-oriented port topology, a working Team page, and
functional and visual verification of every application route.

## Acceptance plan

Implement coherent slices and record their evidence before moving on.

| Slice | Roadmap / exit criteria | Acceptance evidence |
|---|---|---|
| Team correctness | Phase 19: optional configuration preserves local-only mode and explicit authorization | Empty policy renders; valid trust/install works; invalid files and changed review inputs cannot silently proceed; failures remain actionable |
| Monitoring overhead | Phase 12 and ADR-0024: accurate current resource attribution with bounded overhead | Three consecutive five-minute installed closed-app windows below 1% of one core, including child CPU and housekeeping; visible below 5%; hidden within 0.5 percentage points of closed; unchanged user containers and approximately ten-second complete raw metrics |
| Git / worktrees | Phases 8 and 13: current repository state and isolated environments | Staged/unstaged/untracked/conflicted file lists, branches and worktrees; previewed durable branch/runtime transition; dirty/conflicted/in-progress repositories fail safely; no forced checkout, reset or automatic stash; exercise mutations only on owned fixtures |
| Ports topology | Phases 8 and 9: explainable project/service port evidence and accessible UI | Active application/service/endpoints grouped as a topology; conflict evidence, project navigation, explicit optional stopped/unmanaged inventory; accessible narrow-screen representation |
| Complete page audit | Phase 9: functional, accessible, consistent loading/error/disconnected experiences | Dashboard, Projects/detail tabs, Ports, Resources, Logs, Discovery, Workspaces, Machines/companion, Team, Agents, Plugins, Settings and global search/operations; interaction and visual checks at desktop/narrow widths |
| Delivery | Phases 15/18/20: shared behavior and verified installation | Relevant full tests, generated contracts, migrations and architecture checks; installed CLI/native bundles; real workload measurement; clean master and all exact-commit hosted CI jobs passing |

## Guardrails

User projects remain running during profiling and observation. Branch, worktree,
trust, destructive runtime and configuration mutations are tested on owned
fixtures rather than silently applied to the user's repositories. Monitoring
intervals, cancellation, trust, file confinement and fresh lifecycle checks must
not be weakened to hit CPU targets. Native presentation may pause while hidden;
daemon observation must continue.

## Evidence and current status

- Baseline: `7e6efe8`; prior closed samples 1.02–1.12%, visible 6.00%, hidden
  1.80%, including daemon children and attributable native helpers.
- Isolated baseline CPU profile spans startup warmup and a complete five-minute
  housekeeping horizon; profiling code is injected through a Go build overlay
  and is not part of the shipped daemon.
- Team investigation: the application returns `null` for two required policy
  arrays with no installed policy. The template dereferences those arrays and
  fails to render. The repair must preserve the distinction between no installed
  team policy and an installed deny-all policy.
- Team slice: fixed required empty policy arrays, including deny-all packs;
  visible bounded JSON payload review, input-bound confirmation resets,
  malformed/oversized/deep file rejection, selection-race protection and retry.
  Go application tests, frontend typecheck/lint and 12 component
  regressions pass. A real browser against disposable daemon data rendered the
  empty state, trusted an ephemeral public key, verified and installed its
  signed policy, and displayed the resulting restrictions. No user trust or
  policy records were changed. A real-API empty-Team regression was added to
  the existing browser integration flow; full-suite execution is pending.
- Other implementation slices, complete page-audit matrix, installed live
  measurements and final CI: pending.

### Current investigation evidence

- Team fix committed locally as `c3d52d5`; 12 component regressions and real
  signature installation pass.
- First 900-second candidate (`batch-a`) fails the CPU target: consecutive
  five-minute means **2.51%, 2.47%, 2.95%**, overall **2.65%**, median RSS
  **48.34 MiB**, charged footprint **32.64 MiB**. Native app stayed closed;
  daemon children consumed no measurable CPU. Earlier short baseline samples
  are insufficient to characterize periodic work. This candidate is not the
  accepted performance result.
- Shared one-second background configuration feed implemented and tested with
  subscriber isolation, fresh changes, failure propagation, bounded slow-reader
  delivery and cancellation. Direct lifecycle reads remain independent. A
  second isolated CPU profile is running; production instrumentation does not
  ship and the runner restores the installed daemon.
- Git detailed inventory is a separate on-demand endpoint. Added literal
  staged/unstaged/conflicted/untracked file lists, local branches, worktree paths
  and explicit worktree lifecycle previews. Reviewed branch switching is wired
  through durable operations and a root-scoped runtime/checkout gate; fixture
  tests cover stale refs, revoked trust, active/dirty checkouts, hooks and ignored
  files. Full integration/visual verification remains pending.
- Ports now group applications, services and host endpoints; configured ports
  and other host listeners are optional, complete evidence is expandable, and
  conflicts remain visible. Stopped runtime port declarations no longer become
  binding facts. Component/topology and protocol-specific conflict tests pass.
- Relevant Go race tests, Go lint, frontend type checks and 18 focused component
  tests currently pass. Full page audit, full suites, installation and CI remain
  outstanding. No user Git branches or runtime projects have been changed.

### Follow-up progress

- The shared-scheduler-only `batch-b` sustained run also fails: **1.85%,
  1.85%, 1.70%** in its three five-minute windows; overall **1.80%**, median
  RSS **47.33 MiB**. Isolated CPU samples fell from 2.21% to 1.61%; those
  profiling samples are not the installed acceptance measurement.
- `batch-c` adds kernel-enforced pinned directory traversal on Darwin/Linux,
  fresh batched native run snapshots, and one current input resolution per
  combined health/resource read. Its first complete five-minute window is
  **0.93%**; final installed three-window acceptance remains outstanding.
  Focused race checks cover directory alias replacement, FIFO/symlink rejection,
  fresh run values, bounded subscribers, shutdown/restart, per-project history
  limits and combined-monitoring trust/cancellation paths.
- Desktop page screenshots were captured against empty real audit data for
  Projects, Ports, Resources, Logs, Discovery, Workspaces, Machines, Companion,
  Team, Agents, Plugins and Settings. Visual review found and corrected unstyled
  workspace buttons, absent diagnostics padding, a specificity bug in Discovery
  consent alignment, and misleading zero/disconnected storage messages.
- Lifecycle preview styling was scoped to Project Detail and therefore absent
  when opened elsewhere. Move those styles into the preview component and use
  a shared native dialog for lifecycle/branch reviews and Operations, retaining
  keyboard containment, Escape and focus restoration. Focused component tests
  pass; real-browser/narrow regression checks still remain.
- Live Git worktree inventory now includes unregistered checkouts as well as
  durable environment records. Branch switching must also invalidate/bind
  outstanding runtime reviews when the branch changes without a manifest change;
  complete that safety work before declaring the Git slice verified.

### Functional and quality verification

- The complete Go race suite passes with bounded package concurrency (`-p 2`);
  an earlier simultaneous heavy-check run timed out in fixture migration setup.
  Go vet/lint, architecture, repository checks, empty-database migrations,
  Linux/Windows adapter compilation and govulncheck pass (zero called vulnerable
  symbols). Frontend coverage passes: 33 files, 84 tests.
- Public site generation, type checks, lint, 21 unit tests, build and validation
  pass; 158 HTML pages validated. Hosted full site/browser CI is still pending.
- A disposable real Git repository displayed README in both staged and unstaged
  groups and an untracked file, and listed its separate worktree. A confirmed
  branch switch completed as a durable operation, opened a fresh start preview,
  and started the feature branch's distinct HTTP response. The fixture was then
  stopped and removed from the disposable catalog. User branches were untouched.
- Detailed Git inventory has no background poll. Checkout hooks, configured
  filters, dirty/conflicted/in-progress states, stale refs and active runtimes
  fail safely. Native worktree starts are explicitly unavailable pending
  environment-aware run ownership and port remapping; Compose remains supported.
- Full browser, visual, narrow-screen and installed performance acceptance are
  still pending. Candidate-C's short result is not final acceptance.

- All five real API/browser lifecycle flows pass, including Compose, native
  process, workspace coordination, terminal streaming, Settings and empty Team.
  The 16 existing macOS visual cases pass with reviewed updates to Ports and
  Agents; a new 390-pixel long-project-name case passes. The phone route audit
  exposed intrinsic project-card width overflow; corrected header wrapping and
  bounded card/path widths, followed by a full route rerun.
- The native app rebuilt successfully. Final installation, sustained resource
  measurements, Linux visual baselines and exact-commit hosted CI remain pending.

### Sustained installed follow-up

The installed `26f45a4` build's 900-second closed-app run averaged under 1%, but
its five-minute windows were **0.987%, 0.955%, 1.041%**. The sustained criterion
therefore fails; continue profiling rather than accepting the aggregate. A
further isolated CPU profile includes a full housekeeping cycle.

A final accuracy review also found native runtime declarations could appear
bound before sockets existed. Correct this in a separate on-demand port
observation, with current process ownership and actual socket addresses; keep
ordinary health/metrics independent of optional socket permissions.

The next profile (`profile-d`) sampled 2.54 CPU seconds over 330 seconds (0.77%);
its largest remaining costs are Docker HTTP and SQLite reads. Profiling rates
are not the installed acceptance result. Compose metrics previously duplicated
container list/inspection work between status and statistics. A new call-scoped
cycle shares one inventory and post-statistics inspections, without retaining
old restart values, changing intervals or weakening lifecycle reconciliation.
Focused race tests verify reuse, fresh restart counts, inactive containers,
partial failures, cancellation and current trust at the optimized boundary.

### Installed acceptance and native follow-up

The installed `a9cffe7478da` daemon meets sustained closed-app acceptance:
**0.781%, 0.772%, 0.834%** in three consecutive five-minute windows, overall
**0.795%**, median RSS **49.13 MiB**, median charged footprint **34.21 MiB**.
All sixteen real services retained available CPU and memory at every raw point
(90 points each, zero partial samples, worst gap 10.074 seconds). Their container
identities, start times and restart counts remained unchanged. CPU includes
attributable live/reaped children, not just the daemon's own process.

The initial installed native measurements were **4.868% visible** and
**1.382% hidden**, including daemon, app and WebKit helpers, over five minutes
each. These improve the earlier approximately 6.00% / 1.80% samples, but hidden
still misses this work's stricter closed-plus-0.5-point target. A native-only
follow-up narrows the tray presentation payload and confines a bounded Go runtime
policy to its short-lived snapshot child. In 25 reads against the live daemon,
child CPU fell from 45.65 ms to 34.11 ms, payload from 34,479 to 7,198 bytes,
and maximum child RSS from 18.84 to 18.35 MiB. Read sources, intervals, tray and
notification inputs remain unchanged. Installation and repeated native
acceptance measurements for this final adapter change are pending.

The `a9cffe7478da` hosted main CI, full site CI, security checks and site deploy
all passed. This includes Linux/Windows/macOS adapters, native packaging, all
five browser integration flows and seventeen visual cases. The final native-only
commit still requires its own hosted verification.

### Accepted installed results — 2026-10-18

This section supersedes the earlier pending acceptance entries. The installed
daemon remains `a9cffe7478da` (version 1.0.0, schema 18, PID 70733). The installed
CLI and native adapter are `c55b811f167b`. That follow-up changes only the native
CLI presentation, Rust child-process policy, tests and documentation; daemon
runtime/monitoring source is unchanged. Native installation kept the exact
measured daemon process running and retained rollback copies of both artifacts.

The local native bundle uses the optimized release profile with debug assertions
enabled, retaining the existing disabled signed-updater behavior. A normal signed
release intentionally requires updater trust/feed credentials held by release
CI; no placeholder credentials or relaxed updater verification were installed.
The ordinary debug build also passes. The installed app passes strict deep
ad-hoc signature verification, and installed CLI build metadata is verified.

| Native state | Measurement duration | Total CPU (% of one core) | Median summed RSS (MiB) | Median charged footprint (MiB) |
|---|---:|---:|---:|---:|
| Closed | 900 s | 0.795 | 49.13 | 34.21 |
| Dashboard visible | 300 s | 4.890 | 270.32 | 309.88 |
| Hidden, still running | 300 s | 1.207 | 271.82 | 152.36 |

CPU includes the daemon, native shell, coalition-attributed WebKit/Metal helpers,
and cumulative live/reaped child CPU. Docker VM and project workloads are
recorded separately. Memory includes native helpers; summed RSS can count shared
pages, so charged footprint is reported separately. Short-lived snapshot-child
peak RSS was separately measured at 18.35 MiB. Initial launch/hide settling and
compilation are outside the native windows; periodic work inside a window remains
included. These are window averages, not a guarantee against instantaneous CPU
spikes.

After the visible/hidden cycle, a further 300-second closed-app window averaged
**0.764% CPU**, **47.79 MiB RSS** and **35.63 MiB charged footprint**. The exact
native/WebKit cohort had exited, the same daemon process remained running, and
all sixteen services again had 30 complete raw samples (worst gap 10.117 s).
Container identities, start times and restart counts still matched. This checks
that using and quitting the app returns to the measured closed-app baseline.
The verified app was then reopened to leave it available hidden.

All targets pass: three consecutive closed five-minute windows are **0.781%,
0.772%, 0.834%**; visible is below 5%; hidden is below closed plus 0.5 percentage
points (limit 1.295%). Compared with the prior approximately 6.00% visible / 1.80%
hidden samples, native CPU is lower by about 18% / 33%. Earlier short samples and
failed candidates remain above as investigation evidence.

Raw continuity checks found all sixteen service series in every state, with CPU
and memory available at every point and zero partial samples: 90 points each
closed, 30 each visible, 30 each hidden. Worst gaps were 10.074 s, 10.061 s and
11.056 s respectively at the unchanged ten-second policy. Closed accuracy was
archived before the normal one-hour raw retention expired. All sixteen container
IDs, start times and restart counts stayed unchanged, with all eleven projects
retained. No user branch, trust record or project configuration was modified.

The installed dashboard was visually checked after launch. All thirteen routes
passed desktop/390-pixel overflow, serious/critical accessibility and browser
exception checks. Five real browser integration flows and seventeen visual cases
pass, alongside the complete bounded-concurrency Go race suite, Go vet/lint,
architecture/migration/generated-contract checks, frontend checks and 18 Rust
tests. Focused Go race and full lint/vet checks also pass for the final native
presentation change.

Hosted `c55b811f167b` checks are green: [main CI](https://github.com/dcuellar322/envarium/actions/runs/39103239426),
[full site CI](https://github.com/dcuellar322/envarium/actions/runs/39103239424),
[security](https://github.com/dcuellar322/envarium/actions/runs/39103239432), and
[site deployment](https://github.com/dcuellar322/envarium/actions/runs/39103239443).
Final evidence/documentation updates must regenerate the site and receive their
own exact-commit hosted verification before handoff.

The native-worktree limitation recorded above was subsequently addressed by
ADR-0027 and the Phase 13 native-worktree completion on 2026-10-18. Live external pairing, paid AI,
signed release installation and user-data deletion were not exercised on the
user's system. The existing unused-API Linux GLib advisory remains tracked in
`security-cleanup-2026-10-18.md`; passing audit CI does not mean zero dependency
advisories.

Final evidence-commit CI exposed a Windows cancellation-test race: Job Object
active count reached zero before the asynchronous command waiter released the
last process handle, leaving the exited PID briefly queryable. The integration
test now polls for the same PID-disappearance assertion with a three-second
deadline. It still fails if cleanup does not finish; no test is skipped and no
runtime termination, ownership or monitoring behavior changes.
This matches Microsoft's documented [process-object lifetime](https://learn.microsoft.com/en-us/windows/win32/procthread/terminating-a-process):
other open handles can keep an exited process object valid. Twenty-five local
race-enabled repetitions, the full native-process race suite, Go vet/lint and
Windows test compilation pass; hosted Windows execution remains the acceptance
check for this fixture-only repair.
