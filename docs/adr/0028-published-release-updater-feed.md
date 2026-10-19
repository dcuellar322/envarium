---
title: "ADR-0028: Published release updater feed"
description: Serve a bounded, signed-artifact update catalog without a runtime release proxy.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-10-19

## Context

ADR-0017 keeps the public site static and GitHub authoritative for binaries.
Tauri requires HTTP 204 when no release is available; a static empty document
cannot express that response. Updater signatures and platform code signatures
are separate trust boundaries. Mac updater archives also need architecture in
their names before matrix artifacts are merged.

## Decision

Keep release discovery and signature validation at build time. Embed the latest
published stable release's bounded, complete updater catalog in a small Pages
Worker handling only `/updates/stable.json`. It returns HTTP 204 before the
first stable release and HTTP 200 for a complete catalog. All other requests
continue through static Pages assets. Limit Worker invocation with `_routes.json`.
The Worker has no secrets, storage, user input mutations, binary proxy, or live
GitHub calls. This is a narrow exception to ADR-0017's no-runtime-backend rule.

The public updater key is committed separately from the private key. Production
native builds embed that reviewed key and canonical HTTPS feed. Build-time feed
validation checks signature provenance and platform completeness; the native
updater verifies downloaded bytes before installation. Release jobs additionally
verify the complete payload's updater signature before uploading it. Drafts,
prereleases, nightly artifacts, and test signatures cannot become stable updates.
Release publication and withdrawal rebuild the site/feed atomically.

Private updater and Linux signing keys belong in the reviewed release environment
and a restricted local backup, never in the checkout. AppImage embedded GPG
signatures and detached signatures for Linux installers supplement updater
signatures and release attestations. Windows Store signing cannot be reused as
an Authenticode credential for direct downloads.

## Consequences

The feed remains available without GitHub runtime availability or credentials.
A failed build/deployment retains the previous catalog. Revoking a published
release requires a successful site rebuild, rather than editing a live database.
No update is offered until a complete signed stable release is published.
