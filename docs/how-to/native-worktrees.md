---
title: Run native services in Git worktrees
description: Give each checkout independent native processes and ports, and review safe branch and worktree transitions.
category: how-to
audience: [user, contributor]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-18
---

Envarium can run native services from multiple existing Git worktrees at the
same time. Each registered checkout gets independent process ownership, logs,
health checks, live metrics, and exact host-port leases.

## Prepare a portable command

Use an explicit port reference in your `.envarium/project.yml` command or
environment. Envarium replaces `${ports.http}` with the selected checkout's
allocated host port. The main project uses its declared host port.

```yaml
schemaVersion: envarium.dev/v1
kind: Project
metadata: { id: example, name: Example }
repository: { root: . }
runtime:
  driver: process
  process:
    processes:
      - id: web
        command: [node, server.mjs, "${ports.http}"]
        workingDirectory: .
services:
  - id: web
    source: { process: web }
    healthChecks:
      - id: ready
        type: http
        url: "http://127.0.0.1:${ports.http}/"
        expectedStatus: 200
        required: true
ports:
  - { id: http, service: web, host: 8080, target: 8080, protocol: tcp }
```

Here `server.mjs` must read its port from the first command argument. If your
framework reads `PORT`, put `PORT: "${ports.http}"` under the process's
`environment` instead. Every declared native service port must be consumed by
that service's effective command or environment. Hard-coded port numbers are
never rewritten automatically. Working directories must stay inside the selected
checkout, including after resolving symlinks.

## Register and run

1. Create your Git worktrees using your usual Git tools, then approve the base
   project in Envarium.
2. Open the project's **Git** tab and choose **Reconcile worktrees**.
3. Use the checkout card's **Review start** and confirm its exact commands,
   directory, allocated ports, and effects.
4. Start another worktree the same way. Starting, stopping, or restarting one
   checkout preserves independently running worktrees.

Portable and local manifest files are read from the selected checkout. Reconcile
after changing port declarations. Stop an active native checkout before changing
its leases. An occupied lease fails with an error; Envarium does not adopt or
terminate the unrelated listener. Friendly HTTP routing requires an observed,
ownership-verified socket at the allocated TCP port.

## Switch safely

Choose **Inspect checkout** to see that checkout's staged, unstaged, untracked
and conflicted files, branches, and HEAD. **Review switch** guides you through
stopping the selected root's services, reviewing an existing local branch, and
reconciling its declared ports, and reviewing a fresh start plan from the new
branch. The main checkout also checks
its registered primary environment before changing files.

To move running services to another existing checkout, use **Switch services
here…** on the destination card. Review stopping the selected source, then review
the destination restart. Other worktrees continue running. Use worktree card
controls for worktree lifecycle actions.

Each step must succeed before the next review appears. Cancellation or failure
ends the transition; services may remain stopped. Operations retains the outcome
for inspection. Dirty or conflicted checkouts, an in-progress Git operation,
changed refs or manifests, unsafe hooks/filters, and revoked trust fail safely.
Envarium does not stash, reset, fetch, create, or remove worktrees in this flow.

## CLI runtime controls

Use the exact environment ID from `envarium environment list <project>` with
shared runtime commands, for example:

```bash
envarium plan start <environment-id>
envarium start <environment-id>
envarium status <environment-id>
envarium logs <environment-id> --tail 50
envarium metrics <environment-id>
envarium stop <environment-id>
```

The daemon checks current catalog trust and the selected checkout for every
runtime request. The guided branch/worktree transition is available in the Git UI.
