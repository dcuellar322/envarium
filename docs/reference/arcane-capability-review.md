---
title: Arcane capability review
description: A complete mapping of Arcane's documented capabilities to Envarium's project-first local development model.
category: reference
audience: [user, contributor, integrator]
lastVerified: 2026-09-22
---

This review covers the capabilities documented by Arcane as of 2026-09-22. It
is a product comparison, not a source-code port. Envarium adopts the workflows
that improve local project operation while preserving its runtime-neutral,
manifest-driven, least-privilege architecture.

Status meanings:

- **Delivered**: added or materially improved by the operational-experience slice.
- **Existing**: already implemented in Envarium and retained.
- **Equivalent**: the user outcome exists through a safer project-oriented model.
- **Boundary**: intentionally not cloned because it is raw Docker or server-admin
  functionality outside Envarium's accepted product architecture.

## Project and service operations

| Arcane capability                           | Envarium result                                                                                                                                                     | Status    |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| Project up/down/restart                     | Start, stop, and restart from UI, CLI, API, MCP, and remote operations                                                                                              | Existing  |
| Redeploy with image pull and force recreate | `deploy` pulls newer Compose images, force-recreates declared services, removes orphans, and preserves volumes; native services restart from the trusted definition | Delivered |
| Build and deploy                            | `rebuild` builds and force-recreates Compose services                                                                                                               | Existing  |
| Destroy project                             | `teardown` removes Compose containers and networks; volume removal is separate and destructive                                                                      | Existing  |
| Container start/stop/restart/pause/resume   | The same controls are available per declared service without exposing arbitrary containers                                                                          | Delivered |
| Container kill                              | `kill` force-stops a selected Compose service or a verified Envarium-owned native process group                                                                     | Delivered |
| Select services and Compose profiles        | Plans accept validated service IDs and trusted declared profiles                                                                                                    | Existing  |
| Preview and watch command output            | Every lifecycle action shows an exact plan; durable operations stream progress into the activity center                                                             | Delivered |
| Project and container inspection            | Service state, health, ports, runtime identity, container image/exit/restart data, native PID/executable/workdir, CPU, memory, and observation time                 | Delivered |
| Git-backed project workflow                 | Git state, worktrees/environments, trusted actions, manifests, and deterministic discovery                                                                          | Existing  |
| Project templates and variables             | Signed team templates plus manifest environment overlays and credential-store references                                                                            | Existing  |

## Logs, activity, and observability

| Arcane capability                           | Envarium result                                                                                                | Status    |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | --------- |
| Live and historical logs                    | Bounded persisted snapshots plus resumable WebSocket streaming                                                 | Existing  |
| Search and service/level/stream filters     | Client-side filters over the bounded redacted log set                                                          | Delivered |
| Pause display while collection continues    | Rendering can pause without stopping ingestion                                                                 | Delivered |
| Structured JSON fields and multiline output | JSON object fields are inspectable; original messages remain authoritative and preserve whitespace             | Delivered |
| CPU and memory beside logs                  | Current project resource totals are visible in the log header                                                  | Delivered |
| Download logs                               | Existing redacted plain-text and NDJSON exports are now exposed in the project log UI                          | Delivered |
| Activity and events                         | Durable operation state, progress, cancellation, audit, runtime events, diagnostics, and desktop notifications | Existing  |
| Health and degraded states                  | Required health checks, stale/disconnected states, engine availability, and partial-data UI                    | Existing  |

## Images, storage, networking, and remote operation

| Arcane capability                       | Envarium result                                                                                                                       | Status     |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| Image pull/update and image builds      | Deploy pulls; rebuild builds; operation output records progress                                                                       | Equivalent |
| Image/container inventory and deletion  | Envarium inventories project-attributed and shared runtime storage and provides non-executable cleanup previews                       | Equivalent |
| Volume browsing and backup/restore      | Envarium preserves volumes by default and supports application-data backup; arbitrary Docker-volume file browsing is not exposed      | Boundary   |
| Networks                                | Published ports, listener attribution, conflict detection, `.localhost` routing, and project topology replace raw Docker network CRUD | Equivalent |
| Vulnerability scans of arbitrary images | Envarium scans and governs its own dependencies but does not currently act as an image-registry vulnerability platform                | Boundary   |
| Remote environments                     | mTLS federation, certificate pins, capability grants, machine inventory, remote snapshots, and audited typed operations               | Existing   |
| Swarm management                        | Swarm is an infrastructure orchestrator, not a local project runtime supported by Envarium                                            | Boundary   |
| Registries                              | Compose and Docker retain registry and credential authority; Envarium never persists registry secrets                                 | Boundary   |

## Administration and product settings

| Arcane capability                         | Envarium result                                                                                                                                            | Status     |
| ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| Authentication, OIDC, MFA, access control | Local IPC/browser-session security protects the single-user app; MCP profiles, plugin scopes, and federation grants authorize automation and remote actors | Equivalent |
| Notifications                             | Diagnostics and desktop notifications report actionable state without a required hosted service                                                            | Existing   |
| Appearance                                | Dark/light/system themes and density/time preferences                                                                                                      | Existing   |
| Backups                                   | Verified SQLite backup and support-bundle workflows cover Envarium-owned data                                                                              | Existing   |
| Scheduled updates                         | Signed desktop update checks and release-channel controls exist; unattended mutation remains policy-bound                                                  | Equivalent |

## Sources

- [Arcane documentation](https://getarcane.app/)
- [Projects](https://getarcane.app/docs/features/projects)
- [Containers](https://getarcane.app/docs/features/containers)
- [Images](https://getarcane.app/docs/features/images)
- [Volumes](https://getarcane.app/docs/features/volumes)
- [Networks](https://getarcane.app/docs/features/networks)
- [Remote environments](https://getarcane.app/docs/features/environments)
- [Activity and events](https://getarcane.app/docs/features/activity-events)

Arcane evolves independently. Recheck its current documentation before using
this table for future product decisions.
