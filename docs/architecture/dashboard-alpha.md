---
title: Dashboard and product UI
description: Visual hierarchy, server-state ownership, routes, accessibility, and production UI states.
category: concept
audience: [contributor]
since: 1.0.0
lastVerified: 2026-10-19
---

The browser client is the visual adapter over the Envarium control plane. The dashboard does not own runtime policy: it reads
generated OpenAPI contracts, submits typed operations, and follows the same
durable event and log streams used by other clients.

## Route map

| Route | Responsibility |
|---|---|
| `/` and `/projects` | Project overview, filtering, sorting, tags, recent access, and aggregate status |
| `/projects/:id` | Runtime, health, logs, resources, Git, ports, trusted actions, and manifest provenance |
| `/ports` | Declared, reserved, and bound port evidence and conflicts |
| `/resources` | Current and retained project/service metrics, budgets, storage attribution, cleanup preview, and Envarium footprint |
| `/logs` | Bounded cross-project log search and filtering |
| `/discovery` | Deterministic scan, evidence review, validation, and trust approval |
| `/settings` | Revisioned daemon settings, approved roots, retention, tools, provider configuration, and appearance |
| `/workspaces` | Dependency graphs, coordinated lifecycle, and member progress |
| `/agents` | Diagnoses, reviewed actions, automation recipes, and agent sessions |
| `/plugins` | Package discovery, fingerprint trust, scopes, health, and supervision |
| `/fleet` and `/companion` | Optional peer inventory and a read-only companion view |
| `/team` | Trusted publishers, signed bundles, policy, registry metadata, and sync |

These routes share the same Go application services in the browser and the
Tauri desktop webview. Optional integrations show setup or unavailable states
until their prerequisites are configured.

## State ownership

TanStack Query owns all daemon-derived server state. Query keys are scoped by
resource and project, and the event stream invalidates operation state plus
project runtime and health observations at a bounded frequency. Ordinary
polling remains as a fallback when WebSocket events are delayed or unavailable.
The browser bounds dashboard fan-out, log history, and concurrent project
requests to avoid unbounded work in large catalogs.

Recent project access and browser-local fallback preferences use guarded local
storage and degrade to in-memory defaults when storage is denied. Durable
product settings use the revisioned [settings service](../settings.md). No
browser cache is treated as daemon authority.

## Command and operation boundary

Views may select only generated action enums or trusted action identifiers.
They never assemble an executable, argument list, Compose command, or shell
string. Mutations return durable operation records which appear in the global
operation toast and drawer. The drawer reports queued, running, terminal, and
cancellation-requested states and delegates cancellation to the operation API.

This preserves the architecture boundary:

```text
Vue intent -> generated HTTP client -> application use case
                                      -> durable operation
                                      -> trusted runtime/action adapter
```

## Degraded behavior

Each route distinguishes loading, empty, partial, stale, disconnected,
degraded, and failed observations. Docker failure is a capability failure, not
a daemon failure: process projects, discovery, Git, settings, logs already in
the archive, and other local workflows remain available. Container storage is
classified as exclusive, shared, estimated, or unknown with per-resource
evidence. Docker failure degrades only storage; persisted native-process
history remains available. Missing metric reads render as gaps or `—`, never as
synthetic zero consumption.

## Accessibility and visual evidence

The application provides a skip link, visible focus treatment, semantic
landmarks, named dialogs, keyboard-operable tabs and command palette, focus
restoration, reduced-motion support, and responsive layouts. Playwright runs an
axe WCAG 2/2.1 A/AA serious-and-critical gate on the live dashboard. Approved
visual baselines cover populated, empty, degraded, conflict, and narrow states.

The visual fixtures intercept API traffic only in visual tests. End-to-end
tests use the packaged daemon and exercise real deterministic onboarding,
native process lifecycle, and Docker Compose lifecycle behavior.
