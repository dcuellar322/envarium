---
title: Operational experience and resource refresh
description: Execution plan and verification ledger for performance, source control, ports and page usability.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

# Operational experience refresh

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
