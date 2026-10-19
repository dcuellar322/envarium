---
title: v1 compatibility and deprecation policy
description: Stability promises for manifests, APIs, CLI output, MCP, plugins, and local data.
category: reference
audience: [user, integrator, contributor]
since: 1.0.0
lastVerified: 2026-10-19
---

This is the compatibility policy for stable releases. No public stable release
has been published yet; development builds may change before that release.
Envarium uses semantic versioning for the product and marks every
machine-consumed contract with an independent version marker. SQLite remains
an internal storage format and is not an integration API.

## Compatibility table

| Surface | v1 identifier | Compatibility promise |
|---|---|---|
| Product | `1.x.y` | Patch releases fix defects. Minor releases may add opt-in capabilities without breaking v1 workflows. |
| HTTP and WebSocket | `/api/v1`, `apiVersion: v1` | Existing methods, paths, required request fields, meanings, and error codes remain compatible for v1. Additive response fields are allowed. |
| CLI automation | `envarium.cli/v1` | Commands, semantic exit classes, JSON/JSONL envelope fields, and published schemas remain compatible. Human text is not an automation contract. |
| Project manifest | `envarium.dev/v1` | Unknown fields remain errors. Existing fields and defaults are not repurposed. Additions are optional. |
| MCP | stable tool names under server major `1` | Existing inputs, bounded outputs, risk annotations, and permission requirements are not weakened or repurposed. |
| Plugin JSON-RPC | `envarium.plugin/v1` | Exact protocol negotiation. Additions must be safely ignorable or optional; capabilities and scopes retain their meaning. |
| Plugin manifest | `envarium.plugin-manifest/v1` | Exact schema identifier, strict validation, and executable fingerprint review. |

The desktop bundles an exact-version sidecar. It may attach to another
Envarium daemon with the same product major and `apiVersion: v1`; additive
database migrations do not make a compatible v1 daemon unusable by the shell.

## Deprecation

A stable capability is deprecated in release notes, reference documentation,
and the relevant CLI/UI discovery surface before removal. The normal support
window is at least two minor releases and 90 days, whichever is longer. A
security issue may require faster removal; that release must publish the risk,
replacement, and migration procedure.

Envarium does not silently guess across major versions. Unsupported API,
plugin, manifest, or database versions fail before mutation with an actionable
error. A major-version migration must be previewable and create a verified
backup before changing durable data.

## Unsupported manifest versions

Only `envarium.dev/v1` manifests are accepted. There is no manifest-version
conversion command. Use deterministic onboarding to review a new manifest
when an older document uses an unsupported schema. Incompatible plugin executables must be
rebuilt against the stable SDK; the host never runs an incompatible plugin.

Ordered embedded database migrations remain supported for the active local
installation, including verified pre-migration backups. See
[database backup and migration](migration-v1.md).
