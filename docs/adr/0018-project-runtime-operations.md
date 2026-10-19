---
title: 'ADR-0018: Project-first runtime operations'
description: Add deploy, force-stop, service controls, plan confirmation, and operational log tooling without becoming a raw Docker administrator.
category: concept
audience: [contributor, maintainer, integrator]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-09-22

## Context

Envarium already owns trusted project lifecycle, runtime observation, durable
operations, health, logs, resources, and activity. Container-management products
such as Arcane demonstrate a useful operational experience: visible service
controls, a redeploy path, detailed inspection, live activity, and capable log
tools. Envarium must provide those workflows for both Compose and native-process
projects without weakening its project-oriented model or exposing Docker and
the host as unrestricted infrastructure APIs.

## Decision

Keep the runtime action vocabulary driver-neutral and add two explicit actions:

- `deploy` reconciles a trusted current definition. Compose runs `up --detach
--pull always --force-recreate --remove-orphans`; native processes perform a
  verified stop and start with the current manifest definition.
- `kill` force-stops selected services. Compose uses `compose kill`; native
  processes verify durable identity fingerprints before force-signalling the
  owned process group.

All project and service lifecycle requests use the same application service,
durable operation coordinator, permission checks, audit path, and typed API.
The UI requests a side-effect-free plan before queueing any lifecycle action,
shows risk, effects, working directories, and exact argument arrays, and sends
an explicit confirmation for execution. Destructive plans are also rejected by
the HTTP boundary when that confirmation is absent.

Service inspection and logs remain read models over existing runtime and
observability domains. UI adapters may filter, pause rendering, expose structured
fields, correlate resource samples, and download existing redacted exports, but
they may not call Docker, SQL, or process execution directly.

Envarium will not add raw container, image, volume, or network CRUD merely to
mirror a Docker administration product. Project definitions, trusted manifests,
storage inventory, cleanup preview, port evidence, and driver adapters remain
the authoritative boundaries. A future raw infrastructure-management mode would
require a separate product decision and threat model.

## Consequences

Users get the operational controls and evidence needed during local development
across both supported runtime drivers. CLI, HTTP, UI, MCP, and remote-machine
adapters share one action contract. Every command remains previewable, bounded,
and auditable. Service controls remain project-oriented; they do not expose
arbitrary Docker resource administration.
