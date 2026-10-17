---
title: "ADR-0022: Bound default daemon CPU execution parallelism"
description: Reduce Go scheduler wakeups for the local I/O-oriented control plane.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

- Status: Accepted
- Date: 2026-10-17
- Refines: ADR-0002 and ADR-0003

## Context

After eliminating recurring scan subprocesses and redundant configuration/query
work, the actual eleven-project workload still spends a substantial part of CPU
samples waking Go execution threads. The host has twelve logical CPUs; the Go
default chooses twelve execution slots. The user requires under 1% combined
background CPU without weakening monitoring accuracy. A two-slot candidate still
measured 1.09%. The control plane's work is primarily concurrent I/O and short
in-memory computations, not sustained parallel CPU computation.

## Decision

At the CLI daemon composition boundary, default Go CPU execution parallelism to
one slot when GOMAXPROCS has not been explicitly configured. Respect a nonempty
GOMAXPROCS environment override. Restore the previous setting when the daemon
command returns, including startup failure, to isolate embedded CLI/test use.
No other CLI command changes its runtime setting.

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
