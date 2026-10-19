---
title: "ADR-0027: Native worktree ownership and reviewed transitions"
description: Isolate native runtime ownership and exact ports across trusted checkouts.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-10-18
- Refines: ADR-0007, ADR-0008, ADR-0026

## Decision

Keep the catalog project as the durable foreign-key owner of native runs, while
persisting a separate runtime scope identity. A base checkout and each registered
worktree have separate process groups, runs, logs, metrics and lifecycle gates.
Keep health history scoped and log segment metadata catalog-owned as well.
Migrate existing runs without changing their scope or fingerprints. Never create
fake catalog projects to satisfy a run foreign key.

Resolve portable/local manifest inputs from the selected registered checkout,
using the trusted base project's accepted inference as fallback. Bind previews to
that checkout's current configuration and exact allocated ports. Cache projections
per project and root with the existing memory bounds; no additional background
Git or socket polling is introduced. Worktree discovery updates metadata without
overwriting newer runtime routing state or exact leases; unchanged allocations
need no runtime write.

Native commands and environment values may explicitly reference `${ports.ID}`.
Resolve only known declarations to decimal host-port values; never infer and
rewrite arbitrary numeric arguments. A native worktree's declared service ports
must be consumed by that service's effective command/environment, or its start
preview fails with an actionable error. Keep executable and working-directory
selection explicit; reject worktree working directories outside the selected
checkout. Preserve argument arrays and existing shell opt-in.

Use the same local-user-only branch review for registered worktrees. Checkout
changes remain stopped, clean, ref-bound, hook/filter-free and serialized with
runtime actions. The interface guides stop, checkout, and fresh start reviews;
approval of old-branch commands never authorizes newly checked-out commands.
Switching services between existing worktrees stops only the selected source and
reviews a fresh target restart. Other worktrees remain independent. Cancellation
or failure stops the workflow and is visible; do not silently roll back checkout
changes or retry commands from another branch. Pending crash restart backoff
counts as active ownership; stop/kill cancel and join the supervised restart
before checkout mutation. A fully exited failed runtime is quiescent, while
starting, restarting and unknown states remain blocked.

Native friendly HTTP routes require a currently observed, ownership-verified TCP
binding at an exact leased port. Declarations and a live PID alone cannot publish
a route. Existing Compose isolation remains unchanged.

## Verification requirements

Tests must cover concurrent worktrees with independent persisted ownership,
ports, logs, metrics, health, and restart reconciliation. Stale reviews, revoked
trust, fixed or occupied ports, cancellation, and failed checkout transitions
must remain explicit failures. Browser tests use owned fixture repositories;
verification must preserve unrelated running projects.
