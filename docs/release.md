---
title: Release engineering
description: Channels, signing, artifacts, attestations, SBOMs, and the reviewed release process.
category: contributor
audience: [maintainer, contributor]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-17
---

Tagged releases are built only from validated `v*` tags by
`.github/workflows/release.yml`.
The workflow cross-compiles CLI archives with GoReleaser, emits SHA-256
checksums and per-archive SBOMs, builds native Tauri bundles on each operating
system and both macOS architectures, signs updater payloads, applies Apple
notarization and Windows Authenticode signing, generates desktop CycloneDX
SBOMs, attests the combined artifact set, and creates a draft GitHub release
for final human review.
Before that draft is created, the public-site release contract validates the
complete stable platform matrix and replaces build-specific digest files with
one aggregate `checksums.txt` covering every uploaded artifact.

These are workflow capabilities, not evidence of a published signed release.
On October 17, 2026, the repository had no published releases and no configured
`desktop-release` environment. Only the two Cloudflare repository secrets were
present. Passing ordinary CI does not exercise protected release credentials.

## Release account setup

Before tagging the first release, a repository administrator needs to:

1. Create the `desktop-release` GitHub environment, configure release approval
   and allowed-tag rules, and add the secrets below using GitHub's secure UI.
   Never paste private keys or passwords into chat, source, issues, or logs.
2. Supply Apple Developer signing/notarization credentials for
   `dev.envarium.desktop` and a Developer ID Application identity; supply a
   Windows Authenticode certificate matching the intended publisher; supply a
   Linux artifact-signing GPG key with a verified fingerprint.
3. Generate and securely retain the Tauri updater signing key pair. Select a
   maintained HTTPS update-feed URL and implement the signed response/artifact
   hosting for each target, architecture, and release channel. The matching
   public key is embedded in release builds. A URL and key alone do not publish
   or maintain an update feed.
4. Choose the release version/channel and exact reviewed commit. Build a draft
   from its tag, then complete the signed install, upgrade, retained-data
   reinstall, deep-link, and uninstall matrix before publishing it.
5. Select a Homebrew tap and WinGet publisher/package ownership. Create the
   `package-distribution-review` environment with reviewers. After publishing a
   stable release, the package workflow uploads checksummed Homebrew/WinGet
   **drafts** for review; it does not submit or publish either package. Tap
   write access and WinGet submission/review are separate final steps.

The existing release workflow consumes these environment secret names:

| Purpose | Secrets |
|---|---|
| Apple signing and notarization | `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_SIGNING_IDENTITY`, `APPLE_ID`, `APPLE_PASSWORD`, `APPLE_TEAM_ID` |
| Windows signing | `WINDOWS_CERTIFICATE`, `WINDOWS_CERTIFICATE_PASSWORD` |
| Linux signing | `LINUX_GPG_PRIVATE_KEY`, `LINUX_GPG_PASSPHRASE`, `LINUX_GPG_KEY_ID` |
| Updater signing | `TAURI_SIGNING_PRIVATE_KEY`, `TAURI_SIGNING_PRIVATE_KEY_PASSWORD` |
| Embedded updater trust and feed | `ENVARIUM_UPDATER_PUBLIC_KEY`, `ENVARIUM_UPDATER_ENDPOINT` |

Public-key and endpoint values are not themselves secrets, but the current
workflow reads them from environment secrets. Do not change trust configuration
after producing a reviewed artifact without rebuilding and revalidating it.

## Reusing an existing Apple signing identity

A valid **Developer ID Application** certificate and its private key from the
same Apple Developer team can sign both LeagueSaga Import Helper and Envarium.
The certificate identifies the publisher, rather than one bundle ID. An
**Apple Development** or Mac App Store distribution identity is not a substitute
for Developer ID distribution. Envarium retains its own bundle identifier
`dev.envarium.desktop`, entitlements, and notarized build artifacts. Every new
distributed build still goes through notarization.

Configure the matching Apple credentials in Envarium's protected
`desktop-release` environment; existing credentials in another repository are
not automatically available here. Reuse the identity through secure secret
configuration, without copying private material into this repository or chat.
Keep Envarium's Tauri updater key pair and feed separate from other applications:
updater signatures are a separate trust mechanism from Apple's code signature.

See [Apple's Developer ID certificate guidance](https://developer.apple.com/help/account/certificates/create-developer-id-certificates/)
and [Tauri's macOS signing documentation](https://v2.tauri.app/distribute/sign/macos/).

## Release channels

| Channel | Tag form | Distribution behavior |
|---|---|---|
| Nightly | `v1.2.0-nightly.20260716` | Unsigned cross-platform CLI snapshots are also built daily from the default branch and retained as short-lived workflow artifacts. A tagged nightly uses the protected release workflow and is marked prerelease. |
| Alpha | `v1.2.0-alpha.1` | Signed draft prerelease for early contract and migration testing. |
| Beta | `v1.2.0-beta.1` | Signed draft prerelease after the supported-platform workflow matrix passes. |
| Stable | `v1.2.0` | Signed draft release eligible to become the latest published version after human review. |

`scripts/release-channel.sh` is the shared fail-closed classifier. The release
workflow checks out the exact tag for both tag pushes and manual dispatches;
manual dispatch cannot build an arbitrary branch under a release name. Tags
with unknown prerelease labels are rejected. Nightly workflow artifacts are
deliberately unsigned and are never updater-authoritative.

The protected `desktop-release` environment supplies Apple credentials,
`WINDOWS_CERTIFICATE`, `WINDOWS_CERTIFICATE_PASSWORD`, Linux GPG key material,
updater signing keys, and updater endpoint/public-key configuration. Reviewers
confirm that no secret appears in logs or artifacts. The AppImage receives an
embedded GPG signature; signed updater payloads and GitHub attestations cover
every Linux bundle. Repository-specific deb/rpm metadata signing may be added
by a downstream package repository without changing application trust.

## Release candidate checklist

1. Run `make quality` from a clean checkout and confirm generated files stay clean.
2. Confirm native `CI` jobs pass on macOS, Linux, and Windows.
3. Exercise the primary workflow matrix in [platform support](platform-support.md).
4. Test fresh install, alpha/beta upgrade, current-version reinstall, old-binary
   newer-schema refusal, restore-based downgrade, uninstall retaining data, and
   reinstall attaching to retained data.
5. Review dependency, CodeQL, secret-scan, race, vulnerability, and license results.
6. Tag the exact reviewed commit, let the protected workflow build it, and
   inspect the draft release before publishing.

Verify downloads with `checksums.txt`, the published Sigstore bundle, platform
signature tools (`codesign`/`spctl` or `Get-AuthenticodeSignature`), and GitHub's
artifact attestation verification. Compare the included SBOM to the artifact
name and release tag. Reject any updater artifact without its Tauri signature.

The release workflow never publishes partial success: the CLI build and all
four native desktop matrix builds must complete before the draft release job
runs.
