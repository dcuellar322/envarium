---
title: Resolve a local port conflict
description: Identify declaration, reservation, and live-binding evidence before choosing a safe port.
category: tutorial
audience: [user]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-19
sidebar:
  order: 6
---

## Inspect the conflict

```bash
envarium ports list
```

The registry keeps three facts distinct:

- a manifest or Compose **declaration**;
- an Envarium **reservation or lease**, including stopped projects;
- a live operating-system or Docker **binding**.

Open the conflict drawer in the port registry to see the project, service,
protocol, host address, and evidence source for both sides.

## Ask for a bounded suggestion

```bash
envarium ports next --range 15000-19999 --project <project> --json
```

The suggestion considers declarations, leases, current listeners, excluded
ports, and worktree environments. It does not edit project source.

## Apply a reviewed configuration change

Update the portable manifest or machine-local overlay, then validate and
preview start again:

```bash
envarium manifest validate <project>
envarium plan start <project>
```

Success is a clean port registry and a plan whose exact host port matches the
reviewed source.
