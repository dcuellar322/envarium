---
title: Release engineering
description: Channels, signing, artifacts, attestations, SBOMs, and the reviewed release process.
category: contributor
audience: [maintainer, contributor]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-19
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

On October 19, the `desktop-release` environment was configured with the
repository owner's required review and a `v*`-tags-only deployment policy. The
existing Apple API key authenticated successfully with the notarization service;
the private API key and exported certificate were installed as encrypted
environment secrets, and the owner added the certificate export password in
GitHub. Updater signing and Linux GPG signing were subsequently configured with
Envarium-specific keys and restricted local backups. The stable feed is
`https://envarium.davidcuellar.tech/updates/stable.json`. No public stable
release has been published. Windows direct-download signing remains a separate
credential gate: LeagueSaga has no Windows signing secrets configured.

## Release account setup

Before tagging the first release, a repository administrator needs to:

1. Create the `desktop-release` GitHub environment, configure release approval
   and allowed-tag rules, and add the secrets below using GitHub's secure UI.
   Never paste private keys or passwords into chat, source, issues, or logs.
2. Supply Apple Developer signing/notarization credentials for
   `dev.envarium.desktop` and a Developer ID Application identity; supply a
   Windows Authenticode certificate matching the intended publisher; supply a
   Linux artifact-signing GPG key with a verified fingerprint.
3. Retain the existing encrypted Envarium updater key and its password in secure
   backups. The reviewed public key and canonical stable endpoint live in
   `desktop/src-tauri/tauri.release.conf.json`; native release builds and feed
   verification share that configuration. Do not regenerate the key for a new
   build: installed applications trust the existing key.
4. Choose the release version/channel and exact reviewed commit. Build a draft
   from its tag, then complete the signed install, upgrade, retained-data
   reinstall, deep-link, and uninstall matrix before publishing it.
5. Select a Homebrew tap and WinGet publisher/package ownership. Create the
   `package-distribution-review` environment with reviewers. After publishing a
   stable release, the package workflow uploads checksummed Homebrew/WinGet
   **drafts** for review; it does not submit or publish either package. Tap
   write access and WinGet submission/review are separate final steps.

The release workflow consumes these environment secret names:

| Purpose | Secrets |
|---|---|
| Apple signing | `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD` |
| Apple notarization private key | `APPLE_API_KEY_BASE64` |
| Windows signing | `WINDOWS_CERTIFICATE`, `WINDOWS_CERTIFICATE_PASSWORD` |
| Linux signing | `LINUX_GPG_PRIVATE_KEY`, `LINUX_GPG_PASSPHRASE`, `LINUX_GPG_KEY_ID` |
| Updater signing | `TAURI_SIGNING_PRIVATE_KEY`, `TAURI_SIGNING_PRIVATE_KEY_PASSWORD` |

Public-key and endpoint values are committed public configuration, not secrets.
Do not change trust configuration after producing a reviewed artifact without
rebuilding and revalidating it.

### Apple API key configuration

The **Key ID** and **Issuer ID** identify the credential and team; they are not
secret authenticators. The downloaded `.p8` private key is secret. Signing is a
separate operation: it also requires the Developer ID Application certificate
and its private key, usually exported together as a password-protected `.p12`.
An API key cannot replace that signing identity.

Configure these **environment variables** in `desktop-release`:

| Variable | Value |
|---|---|
| `APPLE_API_KEY` | The Apple Key ID, not private-key contents |
| `APPLE_API_ISSUER` | The App Store Connect Issuer ID |
| `APPLE_SIGNING_IDENTITY` | The full Developer ID Application identity name |

Configure these **environment secrets**, using local files or GitHub's secret
entry form without pasting private material into chat:

| Secret | Value |
|---|---|
| `APPLE_API_KEY_BASE64` | Base64-encoded contents of `AuthKey_<KeyID>.p8` |
| `APPLE_CERTIFICATE` | Base64-encoded exported `.p12`, including its private key |
| `APPLE_CERTIFICATE_PASSWORD` | The password chosen when exporting that `.p12` |

Base64 is an encoding, not encryption; GitHub Secrets supplies encrypted storage.
The release job validates the API identifiers and P-256 key, stages the decoded
key exclusively in the ephemeral macOS runner's temporary directory with mode
`0600`, and passes only its path as `APPLE_API_KEY_PATH`. An always-run cleanup
step removes that file before artifact collection. Missing/malformed credentials
fail explicitly. This workflow uses API authentication rather than Apple ID and
app-specific-password authentication.

The same team API key can be reused for notarization if its permissions allow
it; a different API key per application is not required. A matching filename
alone does not prove access: validate with the notarization service before using
the key. If the `.p8` file has been lost, Apple permits only one download of a
new key; create a replacement rather than assuming the IDs recover its contents.
See [Tauri's notarization setup](https://v2.tauri.app/distribute/sign/macos/#notarization)
and [Apple's Notary API authentication guidance](https://developer.apple.com/documentation/NotaryAPI/submitting-software-for-notarization-over-the-web).

## Reusing an existing Apple signing identity

A valid **Developer ID Application** certificate and its private key from the
same Apple Developer team can sign both LeagueSaga Import Helper and Envarium.
The certificate identifies the publisher, rather than one bundle ID. An
**Apple Development** or Mac App Store distribution identity is not a substitute
for Developer ID distribution. Envarium retains its own bundle identifier
`dev.envarium.desktop`, entitlements, and notarized build artifacts. Every new
distributed build still goes through notarization. Tauri notarizes the app; the
release workflow separately notarizes and staples the signed DMG, then checks
Gatekeeper for both the executable application and the downloadable disk image.

Configure the matching Apple credentials in Envarium's protected
`desktop-release` environment; existing credentials in another repository are
not automatically available here. Reuse the identity through secure secret
configuration, without copying private material into this repository or chat.
Keep Envarium's Tauri updater key pair and feed separate from other applications:
updater signatures are a separate trust mechanism from Apple's code signature.

See [Apple's Developer ID certificate guidance](https://developer.apple.com/help/account/certificates/create-developer-id-certificates/)
and [Tauri's macOS signing documentation](https://v2.tauri.app/distribute/sign/macos/).

## Maintained updater feed

The site build fetches GitHub release metadata and updater signatures. Only a
published, non-prerelease stable tag is eligible. It requires Mac arm64, Mac
Intel, Linux x86_64 AppImage, and Windows x86_64 NSIS payloads, signatures from
the committed Envarium updater key, and exact asset URLs in this repository.
Missing, duplicate, foreign-key, or incomplete targets fail the deployment.
The build authenticates each signature's signed metadata; the release workflow
and installed native updater additionally verify the complete payload bytes.

A small, immutable Pages Worker serves `/updates/stable.json`: HTTP 204 before
the first published stable release, otherwise Tauri's static JSON catalog.
It has no runtime credentials, database, user mutations, or GitHub requests.
Only this route invokes the Worker; the website remains static. Preview feeds
always return 204. Publishing, editing, withdrawing, or deleting a GitHub release
rebuilds the production site from master. A failed build retains the previously
deployed feed. Binaries remain hosted by GitHub, and drafts are never offered.
See [ADR-0028](adr/0028-published-release-updater-feed.md).

The release workflow verifies every updater payload before uploading it and
again before creating the draft. Mac archives gain architecture-specific names
before matrix outputs are merged, preventing the two `Envarium.app.tar.gz`
files from overwriting one another. Raw application directories are excluded
from the release upload.

## Linux artifact trust

The dedicated RSA-3072 signing key expires on October 18, 2029. Its fingerprint is:

```text
2380BA10D0E6322DA0B3CD4D985DD308E9EE0B65
```

Download the [public Linux signing key](https://envarium.davidcuellar.tech/signing/envarium-linux.asc)
and compare the fingerprint over an authenticated channel before importing it.
The release job requires AppImage embedded signing to succeed, creates detached
ASCII GPG signatures for AppImage, deb, and rpm files, and verifies each detached
signature against the expected full fingerprint. For example:

```bash
gpg --import envarium-linux.asc
gpg --fingerprint 2380BA10D0E6322DA0B3CD4D985DD308E9EE0B65
gpg --verify Envarium_1.0.0_amd64.AppImage.asc Envarium_1.0.0_amd64.AppImage
```

Detached signatures authenticate the downloaded files. They do not configure an
APT/RPM repository's metadata trust. Tauri updater signatures remain separate
and mandatory for automated updates.

## Windows signing availability

A trusted Windows code-signing identity for the same publisher may sign multiple
applications if its provider permits it. LeagueSaga currently has no
`WINDOWS_CSC_LINK` or `WINDOWS_CSC_KEY_PASSWORD` secret, and its documented Store
route receives Microsoft signing after app certification. That service does not
supply a reusable private certificate for Envarium's direct MSI/NSIS downloads.
A new Envarium Store listing would also need its own package identity and review.

For the existing direct-download workflow, provide a trusted Authenticode
certificate/private key and password through `WINDOWS_CERTIFICATE` (base64 PFX)
and `WINDOWS_CERTIFICATE_PASSWORD`. A hardware or cloud signing identity instead
requires the provider's supported signing integration; do not export a
non-exportable key or substitute a self-signed test certificate. The workflow
checks the signing EKU, expiry, private-key presence, and final installer
Authenticode status/expected thumbprint. Missing credentials fail explicitly.

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
and updater signing keys. Reviewers
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
