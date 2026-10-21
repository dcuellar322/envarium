---
title: Release engineering
description: Channels, signing, artifacts, attestations, SBOMs, and the reviewed release process.
category: contributor
audience: [maintainer, contributor]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-20
---

Tagged releases are built only from validated `v*` tags by
`.github/workflows/release.yml`.
The first release targets **macOS and Linux only**, following
[ADR-0030](adr/0030-first-release-platform-scope.md). Windows and WSL2 remain
coming soon and do not gate release qualification.

The workflow cross-compiles macOS/Linux CLI archives with GoReleaser, emits
SHA-256 checksums and per-archive SBOMs, builds native Tauri bundles for both
macOS architectures and Linux x86_64, signs updater payloads,
applies Apple notarization and Linux GPG signing, generates desktop CycloneDX
SBOMs, attests the combined artifact set, and creates a draft GitHub release
for final human review.
Linux desktop packaging uses Ubuntu 22.04, the oldest supported Ubuntu base,
to avoid requiring a newer host glibc. Source CI can use current runners.
See [Tauri's Linux packaging guidance](https://v2.tauri.app/distribute/debian/).
Before that draft is created, the public-site release contract validates the
complete stable platform matrix and replaces build-specific digest files with
one aggregate `checksums.txt` covering every uploaded artifact.

The workflow defines release requirements; it is not evidence that a signed
release has been published. No public stable release is available yet. Check
[downloads](/download/) for current availability. **Windows and WSL2 support is
coming soon.** A successful development build or ordinary CI run does not
establish installer, signing, or clean-machine support.

## Release account setup

Before tagging the first release, a repository administrator needs to:

1. Create the `desktop-release` GitHub environment, configure release approval
   and allowed-tag rules, and add the secrets below using GitHub's secure UI.
   Never paste private keys or passwords into chat, source, issues, or logs.
2. Supply Apple Developer signing/notarization credentials for
   `dev.envarium.desktop` and a Developer ID Application identity; supply a
   Linux artifact-signing GPG key with a verified fingerprint. Windows signing
   credentials are deferred until Windows enters release scope.
3. Retain the Envarium updater key and its password in secure
   backups. The reviewed public key and canonical stable endpoint live in
   `desktop/src-tauri/tauri.release.conf.json`; native release builds and feed
   verification share that configuration. Do not regenerate the key for a new
   build: installed applications trust the existing key.
4. Choose the release version/channel and exact reviewed commit. Build a draft
   from its tag, then complete the signed install, upgrade, retained-data
   reinstall, deep-link, and uninstall matrix before publishing it.
5. Select a Homebrew tap and create the `package-distribution-review`
   environment with reviewers. After publishing a
   stable release, the package workflow uploads checksummed Homebrew
   **drafts** for review; it does not submit or publish the cask. Tap
   write access and submission/review are separate final steps. WinGet drafts,
   publisher ownership, and submission remain deferred with Windows support.

The release workflow consumes these environment secret names:

| Purpose | Secrets |
|---|---|
| Apple signing | `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD` |
| Apple notarization private key | `APPLE_API_KEY_BASE64` |
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
same Apple Developer team can sign multiple applications, including Envarium.
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
Intel and Linux x86_64 AppImage, deb, and rpm payloads, signatures from
the committed Envarium updater key, and exact asset URLs in this repository.
Missing, duplicate, foreign-key, or incomplete targets fail the deployment.
Windows assets are ignored by the updater catalog and are not required for
release completeness. The build authenticates each supported signature's
signed metadata; the release workflow and installed native updater additionally
verify the complete payload bytes.

Linux clients select their installed package format: `linux-x86_64-deb` and
`linux-x86_64-rpm` point to the matching packages, while `linux-x86_64` remains
the AppImage target. All five macOS/Linux entries are required. Package clients
must not fall back to downloading an AppImage that their installer cannot use.
See [ADR-0031](adr/0031-linux-updater-installer-formats.md).

A small, immutable Pages Worker serves `/updates/stable.json`: HTTP 204 before
the first published stable release, otherwise Tauri's static JSON catalog.
It has no runtime credentials, database, user mutations, or GitHub requests.
The Worker also returns explicit 404 responses for retired documentation
paths ([ADR-0029](adr/0029-retired-documentation-routes.md)). Other pages use
static asset delivery. Preview feeds always return 204. Publishing, editing, withdrawing, or deleting a GitHub release
rebuilds the production site from master. A failed build retains the previously
deployed feed. Binaries remain hosted by GitHub, and drafts are never offered.
See [ADR-0028](adr/0028-published-release-updater-feed.md).

Every updater signature must bind its advertised release version in Minisign's
authenticated trusted comment. Production sets `requireSignedVersion`; artifact
collection, final release validation and feed generation reject unversioned,
duplicate or mismatched versions. Tauri CLI 2.12.3 writes the app version during
build signing. Historical RC signatures without a version remain unpublished.
See [ADR-0032](adr/0032-version-bound-updater-signatures.md).

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

**Windows and WSL2 support is coming soon.** Neither is built, published, or
required by the first-release workflow, updater feed, or package drafts.
Direct MSI/NSIS downloads will require a trusted publisher signing identity
and verified installers. Microsoft Store
certification does not supply a reusable private certificate for direct
downloads; a Store distribution would need its own package identity and review.

Before adding Windows to the release workflow, provide a trusted Authenticode
certificate/private key and password through `WINDOWS_CERTIFICATE` (base64 PFX)
and `WINDOWS_CERTIFICATE_PASSWORD`. A hardware or cloud signing identity instead
requires the provider's supported signing integration; do not export a
non-exportable key or substitute a self-signed test certificate. The workflow
must check the signing EKU, expiry, private-key presence, and final installer
Authenticode status/expected thumbprint. Those checks and clean-machine
installer validation are prerequisites for the future Windows launch.

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

Release candidates use the beta channel, for example `v1.0.0-beta.rc.1`.
The validated tag supplies the version for the native app, bundled Go sidecar,
desktop compatibility check, and architecture-specific updater archives, so a
candidate retains its prerelease identity and can upgrade to the later stable
version. Draft candidates never enter the stable updater feed.

Mac release builds use a reviewed, release-only `xcrun` wrapper to retry
stapling and validate the application ticket before Tauri finishes packaging
the app into disk images and updater archives. Other `xcrun` operations pass
through unchanged. Persistent ticket failures still fail the mandatory
post-build code-signature, ticket, and Gatekeeper checks; they never produce
a draft. Stapler-only diagnostics are retained on the ephemeral runner for
failed-build diagnosis, without logging notarization credentials.

For a manually dispatched rebuild of an existing immutable tag, product code
is checked out from that tag. Mac packaging tools are first staged from the
exact workflow commit into the runner's temporary directory. Required quality
workflows must pass for both revisions when they differ. `metadata.json`
records the tagged product commit; GitHub's attestation authenticates the
packaging workflow revision and ref. Verify both explicitly rather than
assuming a manual build's workflow ref is its product tag. Rebuilding through
a reviewed packaging fix does not move the existing product tag.

The protected `desktop-release` environment supplies Apple credentials,
Linux GPG key material, and updater signing keys. Reviewers confirm that no
secret appears in logs or artifacts. The AppImage receives an
embedded GPG signature; signed updater payloads and GitHub attestations cover
every Linux bundle. Repository-specific deb/rpm metadata signing may be added
by a downstream package repository without changing application trust.

## Release candidate checklist

1. Run `make quality` from a clean checkout and confirm generated files stay clean.
2. Confirm required `CI`, public-site CI, and Security workflows pass for the
   exact commit. Native macOS/Linux checks are required; Windows adapter and
   sidecar jobs remain nonblocking coverage for the upcoming platform.
3. Exercise the macOS/Linux primary workflow matrix in [platform support](platform-support.md).
4. Test fresh install, alpha/beta upgrade, current-version reinstall, old-binary
   newer-schema refusal, restore-based downgrade, uninstall retaining data, and
   reinstall attaching to retained data.
5. Review dependency, CodeQL, secret-scan, race, vulnerability, and license results.
6. Tag the exact reviewed commit, let the protected workflow build it, and
   inspect the draft release before publishing.

Verify downloads with `checksums.txt`, the published Sigstore bundle, platform
signature tools (`codesign`/`spctl` on macOS and GPG on Linux), and GitHub's
artifact attestation verification. Compare the included SBOM to the artifact
name and release tag. Reject any updater artifact without its Tauri signature.

The release workflow never publishes partial success: the four macOS/Linux CLI
archives and all three native desktop builds (macOS arm64, macOS Intel, and
Linux x86_64) must complete before the draft release job runs. Stable validation
requires both Mac DMGs, Linux AppImage/deb/rpm, and every CLI archive with
checksums, plus all five authenticated updater payloads. Windows artifacts
and credentials are not part of this gate.
