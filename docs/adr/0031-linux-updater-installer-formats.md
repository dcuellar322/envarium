---
title: "ADR-0031: Linux updater payloads match the installed package"
description: Publish authenticated AppImage, deb, and rpm updater entries for Linux x86_64.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-21
---

- Status: Accepted
- Date: 2026-10-21
- Supersedes: ADR-0028's generic Linux updater payload selection only.
- Related: ADR-0030's macOS/Linux platform scope remains unchanged.

## Context

RC3 qualification reproduces a format mismatch in the proposed stable catalog.
The signed Linux binaries identify their installed bundle as AppImage, deb, or
rpm. The pinned Tauri updater first searches for an installer-specific entry,
then falls back to the generic OS/architecture entry. Offering only an AppImage
at `linux-x86_64` makes deb and rpm clients download that AppImage and reject it
as an invalid package before installation.

## Decision

Keep `linux-x86_64` for the signed AppImage and add required
`linux-x86_64-deb` and `linux-x86_64-rpm` entries pointing to their respective
signed packages. Retain the two macOS architecture entries. Require all five
payload signatures and exact repository asset URLs before creating a stable
catalog or validating stable release files. Missing Linux installer entries fail
closed. Windows remains coming soon.

The native updater continues to select its own installer format and verify
downloaded bytes with the existing public key. No production endpoint or key
changes, test feeds, package conversion, or binary rebuilding are required.

## Consequences

AppImage clients retain their existing generic target. Package clients receive
their native installer instead of an incompatible fallback. Linux package
updates still require normal OS authorization; isolated signature and format
selection checks do not certify interactive authorization or app restart.
Clean-machine native desktop, signed update/relaunch, and retained-data manual
acceptance remain release gates.
