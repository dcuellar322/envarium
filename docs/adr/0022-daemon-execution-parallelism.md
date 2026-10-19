---
title: "ADR-0022: Bound default daemon CPU execution parallelism"
description: Reduce Go scheduler wakeups for the local I/O-oriented control plane.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-10-17
- Refines: ADR-0002 and ADR-0003

## Context

An I/O-oriented local daemon can spend unnecessary CPU waking multiple Go
execution threads during small periodic reads. The control plane primarily
performs concurrent I/O and short in-memory computations rather than sustained
parallel CPU computation.

## Decision

At the CLI daemon composition boundary, default Go CPU execution parallelism to
one slot when GOMAXPROCS has not been explicitly configured. Respect a nonempty
GOMAXPROCS environment override. Restore the previous setting when the daemon
command returns, including startup failure, to isolate embedded CLI/test use.
Current refinement: [ADR-0024](0024-shared-observation-and-sampling-isolation.md)
also applies the entry-point execution budget to CLI reads and defines separate,
bounded desktop snapshot bridge settings.

This limits simultaneous Go CPU execution, not goroutine concurrency or operating
system threads. Network waits, native process supervision, probe worker limits,
configured probe intervals, runtime/socket reconciliation, metric sampling and
cancellation stay unchanged. Project processes, containers, builds and external
AI/CLI commands remain independent. No CPU percentage limiter or sleep loop is
introduced, and the setting is not changed based on window visibility.

Verify concurrent probe completion, cancellation, actual snapshot latency,
raw sample cadence and workload health before accepting the default. Measure
normal installed binaries in repeated clean windows, including child CPU.

## Consequences

The local daemon avoids excess CPU-worker wakeups in a small I/O workload.
CPU-heavy control-plane operations may take longer; operators needing parallel
Go computation can explicitly configure GOMAXPROCS. This is process runtime
configuration at the composition root, not mutable global product state.
Revisit this default if representative load tests show delayed monitoring or
unacceptable command latency.
