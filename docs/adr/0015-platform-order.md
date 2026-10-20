---
title: "ADR-0015: Platform support order"
description: Implement OS adapters in an explicit order behind platform-neutral contracts.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-07-15

## Context

Process, port, IPC, PTY, notification, launcher, and packaging behavior differs
substantially by operating system. Attempting simultaneous parity would slow
the product and spread conditionals through domains.

## Decision

Implement and validate macOS first, Linux second, and Windows/WSL third.
Platform-specific behavior lives behind capability interfaces and OS adapter
packages. Domain and application packages remain platform-neutral.

## Consequences

Source builds currently target macOS and Linux. Windows and WSL2 support is
coming soon; source adapters and passing CI do not establish installer or
clean-machine support. See [platform support](../platform-support.md) for
availability and platform boundaries. An unsupported capability returns a
typed, explainable state rather than silent failure.
