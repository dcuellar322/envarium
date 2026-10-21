---
title: "ADR-0033: RPM prerelease metadata"
description: Preserve product SemVer while generating valid RPM prerelease ordering.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-21
---

- Status: Accepted
- Date: 2026-10-21
- Extends: ADR-0030, ADR-0031 and ADR-0032.

## Context

Tauri CLI 2.12.3 provides the required version-aware updater signer, but its RPM
builder rejects a SemVer prerelease hyphen in the RPM Version field. RPM uses
separate Version and Release fields; neither permits that separator. Historical
private candidates produced nonstandard headers. No stable release is published.

## Decision

Keep product/native/CLI/updater versions as their exact SemVer. Build the Linux
app, AppImage and deb with that version. Use the official bundle-only command
for RPM, setting only package Version to the core version and Release to
`0.<prerelease>` (internal prerelease hyphens become underscores). Stable RPMs
use Release `1`. For example, product `1.0.0-beta.rc.5` is packaged as Version
`1.0.0`, Release `0.beta.rc.5`; the stable `1.0.0-1` sorts after it.

RPM bundling does not compile the app again. Verify the native build output and bundled sidecar bytes remain equal; the
official bundler patches its per-package copy and restores the build output.
Qualification additionally checks the RPM payload's native format marker.
Disable automatic updater artifact signing only for this bundle-only step,
then explicitly sign the RPM with the original full product SemVer using the
official signer. Existing mandatory collection and final payload/version checks
must pass before uploads or a draft. AppImage and deb signatures remain bound
to the full SemVer. No private key or trust policy changes.

Stage this reviewed release-only tool from the packaging workflow revision
before checking out an immutable product tag, as for the Mac stapler wrapper.
Quality checks and provenance identify both revisions. Do not move the product
tag to repair packaging.

## Consequences

Native updater format markers and version binding remain correct. RPM package
metadata follows its own format while the running app reports the full candidate
version. Test numeric RC ordering and the RC-to-stable transition using RPM itself.
Historical unpublished RPM candidates may require removal/reinstallation because
their old nonstandard Version sorts differently; retain data and qualify that
transition explicitly. Do not add a permanent Epoch or force-downgrade switch to
the production updater to accommodate private test packages.
