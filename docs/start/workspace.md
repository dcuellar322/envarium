---
title: Start your first workspace
description: Coordinate related projects in dependency order with visible health gates and failure policy.
category: tutorial
audience: [user]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-19
sidebar:
  order: 7
---

Create a workspace in the browser after each member project is independently
trusted and can reach its own success state.

## Review the graph

Add projects and dependencies, then choose the partial-failure policy. Cycles
are rejected. Independent branches may start in parallel; dependents wait for
required health gates.

## Start from the CLI

```bash
envarium workspace list
envarium workspace get <workspace-id>
envarium workspace start <workspace-id>
```

The start command returns a durable operation. Open the workspace progress
view to see queued, running, healthy, failed, skipped, and rollback states per
project.

## Observable success

Every dependency reaches its configured health gate before its dependents
start. A partial failure follows the selected continue or rollback
policy without tearing down project data. Bulk stop preserves volumes unless a
separate destructive action is explicitly authorized.
