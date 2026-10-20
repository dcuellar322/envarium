---
title: "ADR-0030: First release platform scope"
description: Qualify and publish macOS and Linux before Windows and WSL2.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-20
---

- Status: Accepted
- Date: 2026-10-20
- Supersedes: The all-platform release completeness requirement in ADR-0028.
- Related: ADR-0015 platform support order.

## Context

Windows and WSL2 are advertised as coming soon, but the release workflow,
stable artifact contract, updater catalog, and package drafts require Windows
artifacts and signing credentials. That makes Windows a prerequisite for the
first macOS and Linux release despite the accepted platform support order.

## Decision

The first release distributes macOS and Linux only. Require signed/notarized
macOS arm64 and Intel desktop bundles, signed Linux x86_64 AppImage/deb/rpm
bundles, and macOS/Linux arm64 and x86_64 CLI archives. The stable updater
catalog requires exactly the two macOS architectures and Linux x86_64;
Windows assets do not enter the catalog or its required artifact matrix.

Remove Windows builds and signing from the release workflow and Windows CLI
archives from GoReleaser. Keep Windows native adapter and sidecar CI as
nonblocking coverage for the upcoming platform. Local Windows cross-compilation
is optional through `make platform-check-windows`. macOS/Linux CI, security,
artifact integrity, signatures, attestations, and human release review remain
required. Package distribution prepares Homebrew drafts only; WinGet remains
deferred. Windows source adapters and reusable package-format code remain in
the repository without a public support claim.

## Consequences

Missing Windows credentials, artifacts, or failing Windows CI cannot block
macOS/Linux release qualification. Missing macOS/Linux artifacts, checksums,
or updater signatures still fail closed. Clean-machine install, upgrade,
retained-data reinstall, and uninstall acceptance remains required on macOS
and Linux before publishing; this decision does not certify those gates.

Windows/WSL2 launch requires a separately reviewed expansion of the release
matrix, signing configuration, updater targets, package distribution, and
clean-machine acceptance evidence. Until then both remain coming soon.
