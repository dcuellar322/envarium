---
title: "ADR-0021: Direct macOS socket inventory with compatibility fallback"
description: Avoid recurring scan subprocesses while retaining kernel port and PID evidence.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-10-16
- Refines: ADR-0007 and ADR-0015

## Context

Repeated socket-inventory subprocesses add process creation and output-parsing
overhead. The adapter must reduce that work without weakening socket evidence
or reducing monitoring accuracy.

## Decision

Read macOS TCP/UDP PCB inventories directly using the sysctl API used by Apple's
netstat. Decode length/kind tagged records, generation filtering, network-order
ports, IPv4/IPv6 addresses, TCP state and last-owner PID. Keep the existing
observation coalescing window and fresh lifecycle/conflict reads. Revalidate
process fingerprints as before; a socket owner PID is never ownership authority.

These exported records use private ABI definitions. Check every supported record
layout, group completeness, protocol and framing. Fall back to the existing
numeric netstat reader on unavailable APIs, permission errors or unknown layouts;
never turn a decode error into an empty successful inventory. Cancellation stays
visible. Keep input size bounded, and test live IPv4/IPv6 TCP and UDP listeners,
PID attribution, malformed data and comparison with netstat. No cgo or extra
runtime dependency is introduced. Other platforms retain their existing adapter.

## Consequences

Supported macOS versions avoid recurring process creation for monitoring.
Future incompatible kernels retain observation through the compatibility path,
with its original cost. ABI changes require fixture and live-parity verification.
This changes the platform adapter only, not monitoring cadence or product policy.

ABI references: Apple's [netstat reader](https://github.com/apple-oss-distributions/network_cmds/blob/main/netstat.tproj/inet.c),
[PCB records](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/netinet/in_pcb.h),
[socket records](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/socketvar.h), and
[TCP records](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/netinet/tcp_var.h).
