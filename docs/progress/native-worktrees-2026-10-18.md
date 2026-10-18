---
title: Native worktree completion and installed verification
description: Evidence for isolated native checkout runtimes and reviewed service transitions.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-18
---

## Delivered slice

ADR-0027 completes native worktree lifecycle support for existing registered
checkouts. CLI status, plan, lifecycle, logs and metrics accept exact environment
IDs while preserving catalog project selection. Runtime scope IDs isolate durable ownership beneath the trusted
catalog owner. Selected-checkout manifests and local overlays determine commands;
explicit `${ports.ID}` references consume that checkout's exact leases. Native
lease changes use the same root lock as lifecycle execution and inspect fresh
owned process evidence before writing. Metadata reconciliation preserves newer routes and leases.
Retired owned services remain observable
and are included in whole-scope stop/restart/kill cleanup. Pending crash restarts
remain active ownership; stop/kill cancel and join them before reporting a
quiescent checkout. An exited failure with no active service can still switch
branches without falsifying the recorded failure.

The Git selector inspects each registered checkout. Branch switching guides
separate stop, checkout, and fresh-start reviews. Moving services between
worktrees reviews the source stop and destination restart, then selects the
destination. Failed/cancelled stages end the sequence; queued operations remain
visible in Operations. Native worktree bindings appear in the port registry and
HTTP routes require exact owned socket evidence.

Schema 19 preserves existing run fingerprints and health scopes. New worktree
runs, archived log queries and health history remain isolated. A lossy downgrade
fails transactionally. No additional closed-app Git or socket poll was added.

## Local verification

- Go vet, full tests, full race tests, lint, architecture and repository checks pass.
- Migration tests cover schema 18 upgrades, preserving ownership, a lossless
  downgrade, and refusal to collapse isolated ownership on downgrade.
- UI lint/type checks and 89 tests in 35 files pass; 17 macOS visual checks pass.
- All six real browser flows pass. The native-worktree acceptance also passes
  separately with explicit port-registry attribution checks. It runs two owned
  Git worktrees concurrently, verifies separate ports, logs and live metrics,
  switches one to a distinct branch response and different port declaration while
  the other's run stays intact,
  and reviews moving services to the other checkout. Fixture services are stopped
  before its repositories are removed.
- Site quality passes: 125 documentation sources, 161 built pages, no content or
  build errors. Site browser checks pass 60 cases with 16 configured skips; all
  six site visual cases pass.
- Govulncheck reports no called vulnerabilities. Existing uncalled module and
  Linux desktop dependency advisories remain subject to the recorded security policy.

## Installed and hosted acceptance

The optimized local macOS app, installed CLI and development bundle helper use
runtime commit `09e0c03fc375` (version 1.0.0). The app and helpers pass deep code
signature verification. The daemon stopped gracefully, an offline schema 18
backup passed integrity verification, and the new daemon became ready on schema
19. The installed native dashboard was visually checked with the real catalog.
The user's 11 projects and 16 running containers retained their identities,
start times and restart counts.

CPU below is a percentage of one logical core, summed across the daemon, native
app and native helpers, including reaped helper CPU. Memory values are medians of
process totals; RSS and macOS physical footprint use different accounting. Docker
VM and project workload usage are recorded separately from Envarium overhead.

| App state | Duration | CPU | RSS | Physical footprint |
| --- | --- | --- | --- | --- |
| Closed | 15 minutes | 0.84% | 59.2 MiB | 41.1 MiB |
| Visible native dashboard | 5 minutes | 4.84% | 284.9 MiB | 343.5 MiB |
| Hidden, still running | 5 minutes | 1.26% | 287.1 MiB | 168.4 MiB |
| Closed after native use | 5 minutes | 0.81% | 51.0 MiB | 42.3 MiB |

The three separate closed-app five-minute windows were 0.804%, 0.822% and
0.881%, satisfying the under-1% target. Native helper cohorts and app visibility
were checked throughout each sample. No product monitoring interval was reduced.

All 16 service series had CPU and memory available and zero partial samples.
Each service had at least 90 points during the closed sample and 30 during every
five-minute sample. Worst raw-sample gaps were 11.04 seconds closed, 10.15 visible,
10.06 hidden, and 10.05 after quitting, against the configured ten-second cadence.
The accuracy evidence was captured before raw retention could remove it.

Private raw measurements and verification records live under the ignored
`.cache/resource-live-20261016/native-worktree-*.json` and
`.cache/experience-20261017/native-worktree-verification-*.json` paths. The
installation record includes the offline backup and rollback paths.

### Hosted acceptance and runner isolation

The first hosted browser failure was the expected production refusal of inherited
Git LFS checkout filters configured on runner images. The test daemon and its
owned Git fixtures now isolate system/global Git configuration; the production
filter refusal remains intact. Fixture project IDs are unique per attempt so a
retry cannot collide with an earlier accepted catalog entry. Two consecutive
native-worktree browser runs passed with a temporary runner-style Git LFS config.

All workflows for fixture/evidence baseline `e0d3109` passed:
[main CI](https://github.com/dcuellar322/envarium/actions/runs/39134552841),
[full site CI](https://github.com/dcuellar322/envarium/actions/runs/39134552700),
[security](https://github.com/dcuellar322/envarium/actions/runs/39134552791), and
[site deployment](https://github.com/dcuellar322/envarium/actions/runs/39134552751).
Main CI passed every job, including Linux/Windows/macOS adapters, native packaging,
Go/race/lint/generated/migration checks, frontend checks, real browser/visual
flows, and embedded release/MCP packaging. Site CI validated 125 documentation
sources and 161 HTML pages, passed 60 browser cases with 16 configured skips, and
passed all six visual cases.

Production source paths (`cmd`, `internal`, `api`, `sdk`, `web/src`, `desktop`)
remain identical to installed runtime commit `09e0c03fc375`; subsequent changes
contain test-fixture isolation and verification evidence. The final evidence-only
master commit also receives hosted verification before handoff.
