---
title: "ADR-0020: Background configuration resolution without CLI polling"
description: Resolve observation metadata in-process while preserving installed-CLI lifecycle validation.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-10-16
- Refines: ADR-0006 and ADR-0008

## Context

Repeated Compose CLI normalization creates child-process overhead while
independent observers duplicate decoding, merging, and validation of unchanged
manifests. Longer polling intervals would
trade freshness for overhead rather than remove the repeated work.

## Decision

Use the official compose-go reference loader in the Compose adapter to resolve
read-only project names and service membership in-process. Continue using the
installed Compose CLI, with fresh normalization, for every lifecycle plan and
execution. Preserve the existing bounded observation refresh window; parsing
errors remain visible and never become a successful empty runtime. Observation
uses local files only, never downloads remote includes or invokes repository
commands. Compare loader metadata with the installed CLI on supported fixtures
and representative local projects before changing the observation loader.

Read Docker's active context selection and non-secret endpoint metadata from
its local configuration/context store for SDK observation. Preserve context and
environment precedence, verified TLS, explicit errors, and fresh lifecycle CLI
resolution. Read only bounded files; do not read credential helper results or
persist configuration contents.

Memoize effective manifest resolution by exact source contents, including all
precedence layers and runtime overrides. Read current sources before every
lookup, preserve provenance, return independent results, bound retained memory,
and never cache read or validation errors. For persisted accepted snapshots,
a current catalog revision read may identify their immutable contents without
rereading the snapshot body; legacy/unknown revisions use exact content reads.
Local files are always reread before lookup. This memoization has no staleness
window and does not bypass trust, containment, or lifecycle authorization.

Read-only Compose metadata may reuse a previously validated result when exact
contents of all known local inputs and the process environment are unchanged.
Track top-level Compose files, the default dotenv file (including its absence),
and literal service env/label files; reread these inputs within the existing
observation refresh window. Validate before and after loading to avoid retaining
a result across an input race. Includes, extends, dynamically named dependency
files, explicit dotenv modes, and unrecognized dependency syntax fall back to
the reference loader on every refresh. Errors and unsupported graphs are never
memoized. Context selection remains independently fresh. Bound both entries and
tracked input count; retain digests/metadata rather than environment contents.

## Consequences

Background observation no longer requires process creation for unchanged local
Compose projects. The Compose parser is an explicit pinned dependency which
must be updated and tested for compatibility. Installed Compose remains the
lifecycle authority; unsupported remote observation inputs fail explicitly.
No product logic moves to Rust or the UI, and no global mutable cache is added.

Implementation references: [Compose reference loader](https://github.com/compose-spec/compose-go/tree/v2.16.1) and
[Docker context store](https://github.com/docker/cli/tree/v29.4.3/cli/context/store).
