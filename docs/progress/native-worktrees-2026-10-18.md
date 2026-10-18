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
and are included in whole-scope stop/restart/kill cleanup.

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
- Site quality passes: 124 documentation sources, 160 built pages, no content or
  build errors. Site browser checks pass 60 cases with 16 configured skips; all
  six site visual cases pass.
- Govulncheck reports no called vulnerabilities. Existing uncalled module and
  Linux desktop dependency advisories remain subject to the recorded security policy.

## Installed and hosted acceptance

Build/install, sustained resource measurements with the user's projects, and
exact-master hosted workflow results will be appended after completion. Before
upgrade, the user's 11 catalog projects and 16 running containers were recorded;
installation must preserve their identities, start times and restart counts.
