---
title: "Phase 15: Desktop application"
description: Implementation evidence for Envarium product phase 15.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

## Implemented

### October 19 Apple credential setup

The release workflow now accepts App Store Connect API-key authentication,
using non-secret Key ID/Issuer ID/signing identity environment variables and
encrypted private-key/certificate secrets. A tested staging adapter validates
the P-256 key, refuses existing files/symlinks, creates an owner-readable private
file only in the ephemeral runner directory, exports only its path, and removes
it on setup failure or after the build. Tests cover valid staging, permissions,
missing/malformed input, overwrite refusal and failure cleanup. The existing
exact-commit quality and reviewed-tag gates remain unchanged.

The protected GitHub environment now requires the owner's review and accepts
only `v*` tags. The reused LeagueSaga API key authenticated with Apple's
notarization history endpoint; no private key contents were printed or committed.
The owner added the certificate-password secret directly in GitHub; its presence
was verified without reading its value. A real signed/notarized artifact remains
unverified. Updater feed/key configuration and other-platform signing are still
separate release gates; ordinary CI does not satisfy those external criteria.
Local verification passed the four release-quality/key-staging tests and
actionlint 1.7.12 for both changed workflows. The actual existing API key also
passed the staging adapter's parsing and owner-only-permissions checks in a
temporary directory that was removed afterward.

### Original desktop implementation

- Added a pinned Tauri 2/Rust native shell and generated platform icon assets
  from the Envarium mark.
- Bundled the trimmed Go binary as a target-triple sidecar with exact version,
  commit, and build-time identity.
- Added a typed, versioned `desktop snapshot` CLI boundary for daemon, host,
  project runtime/health, workspace, recent-operation, and port-conflict state.
- Added exact desktop/bundled-sidecar and compatible same-major daemon/API/
  minimum-schema preflight before browser credential issuance or native
  mutations.
- Added a migration rollback guard: a binary refuses a database newer than its
  embedded migrations before Goose attempts an upgrade.
- Added a live tray with daemon state and bounded project/workspace open,
  start, and stop actions; close-to-tray preference; optional launch-at-login;
  single-instance focus; and project/workspace deep links.
- Added transition-based native notifications for failed/partial operations,
  project health, daemon connectivity, Docker connectivity, port conflicts,
  and high host resource use.
- Added a release-only, public-key-verified updater and a protected GitHub
  workflow for signed, notarized, draft macOS releases.
- Added macOS desktop CI, source-build/headless installation guidance, and
  explicit uninstall/data-retention choices.

## Architecture decisions

No ADR changed. The implementation applies ADR-0003's bundled Go control plane,
ADR-0009's thin Tauri adapter, ADR-0013's authenticated loopback boundary, and
ADR-0015's macOS-first platform order. Rust contains native presentation and
fixed command mappings only; all product behavior remains behind Go application
services.

## Safety properties

- The Vue loopback origin has no Tauri capabilities. Native plugins are called
  only by Rust, and no arbitrary shell command is exposed.
- Deep links, UI routes, resource identifiers, sidecar envelopes, response
  size, version, API, and database schema are validated before use.
- Raw sidecar stderr and repository/log payloads are never copied into desktop
  notifications.
- First observation does not replay old failures. Health, daemon, and resource
  notices are edge-triggered rather than emitted every poll.
- Closing or uninstalling the adapter never deletes state, repositories,
  containers, volumes, or logs. Newer databases fail closed under older code.
- Release builds require updater and Apple signing material from a protected CI
  environment; local debug builds have no update authority.

## Tests added

- Database-newer-than-binary rejection before migration.
- Local UI path validation for remote origins, network paths, dot/encoded-dot
  traversal, fragments, backslashes, and injected bootstrap credentials.
- Exact desktop/sidecar plus same-major daemon/API/minimum-schema compatibility
  acceptance, additive-schema acceptance, and fail-closed mismatch cases.
- Deep-link route parsing and rejection of traversal, foreign origins, query
  injection, and malformed identifiers.
- Sidecar resource-identifier containment and versioned command response
  validation.
- Notification initialization, operation/health transitions, warning
  deduplication, and recovery.

## Acceptance criteria status

- [x] Desktop behavior is the same Go application behavior used by browser,
  CLI, and MCP clients.
- [x] Closing the window follows the persisted user preference and never
  fabricates daemon shutdown.
- [x] The tray reflects daemon/project/workspace state and uses fixed typed
  actions.
- [x] Incompatible sidecar/daemon/API/minimum-schema combinations are detected
  before native mutation; newer databases are also rejected by Go before
  migration.
- [x] Installation supports signed/notarized macOS releases and a CLI-only
  headless path.
- [x] Uninstall guidance offers preserve, backup, and reviewed-delete data
  choices without implying repository or Docker-volume deletion.

## Verification evidence

Focused verification on 2026-07-16 passed:

- the Go SQLite, CLI, and HTTP-client packages passed their focused tests;
- six Rust unit tests passed and Clippy with `-D warnings` was clean;
- the bundled sidecar reported version `0.1.0-alpha.0` and the current commit;
- Tauri built a native x64 `Envarium.app` and 24 MiB DMG; `hdiutil verify`
  reported a valid checksum;
- an ad-hoc signature passed `codesign --verify --deep --strict`;
- the packaged app launched a visible 1440×960 `Envarium` window and started
  its bundled detached daemon under the expected per-user data directory;
- clicking the standard close control removed the window while the desktop
  process remained alive, and a second launch restored the same instance;
- the app's Info.plist contains the `envarium` URL scheme; deep-link parsing
  and route selection passed unit tests; and
- quitting the adapter left the independent daemon running as documented.

The complete repository `make quality` run passed: generated artifacts were
clean; Go formatting, vet, GolangCI-Lint, architecture checks, the full suite,
race detection, migration checks, and Govulncheck passed; all 20 Vue unit tests,
four browser E2E tests, and nine visual tests passed; all six Rust tests,
rustfmt, and Clippy with `-D warnings` passed; and the production Vue/Go build,
native application, and DMG were produced. Apple notarization itself is
credential-gated and is verified by the protected release workflow rather than
a local debug build.

## Known limitations and deferred work

- Phase 15 packaged macOS first. Phase 18 now supplies the Linux/Windows bundle
  matrix, native Windows adapters, and explicit WSL behavior.
- Signing, notarization, and production updater-feed access require external
  Apple and Minisign credentials and cannot be truthfully exercised by an
  unsigned local build.
- The daemon intentionally outlives the desktop adapter so CLI, browser, MCP,
  and detached terminal sessions remain available.

## October 19, 2026 — updater trust and release signing

This vertical slice addresses Phase 15's signed updater and notarized packaging
deliverables, plus Phase 18's cross-platform artifact trust. ADR-0028 records
the narrow Pages response adapter needed for HTTP 204 before first publication.
It does not introduce a daemon cloud dependency or move lifecycle policy to Rust.

- Generated an encrypted, Envarium-specific updater key; verified the pinned
  Tauri signer's output independently and rejected modified payload bytes.
  Private key/password backups use mode 0600 in a mode 0700 directory outside
  the checkout; GitHub stores them in the reviewed `desktop-release` environment.
- Committed the public updater configuration, shared by native release builds,
  release verification, and feed generation. The canonical stable feed excludes
  drafts/prereleases and fails closed on missing/duplicate platforms, foreign
  URLs, malformed signatures, or another signing key.
- Added the bounded immutable feed Worker and publication/withdrawal rebuilds.
  The empty catalog returns HTTP 204 without inventing a version or test release.
- Prevented Mac architecture archive collisions, excluded raw application trees,
  and added complete updater payload verification before upload and attestation.
- Generated a dedicated RSA-3072 Linux identity, verified a live GPG signature,
  configured release secrets, and published its public key/fingerprint in site
  sources. Linux jobs require embedded AppImage signing and verify detached
  signatures for each installer against the full expected fingerprint.
- Confirmed LeagueSaga has no Windows signing secrets configured. Its Store
  route relies on Microsoft's per-app certification/signing. No reusable Windows
  private certificate was found. Direct MSI/NSIS release signing remains gated
  on an appropriate publisher identity; self-signed test keys are not substituted.

Focused site tests cover valid/tampered/foreign-key signatures, platform
completeness, draft exclusion, HTTP 204/200/HEAD/method handling, static asset
fallback, and architecture-specific collection names. Site lint/type/build/content
validation and workflow lint passed locally. The production Mac app was accepted by Apple (submission
`03b57029-75b8-4daa-a87a-1df609a37a56`); strict/deep code-signature verification,
stapled-ticket validation, and Gatekeeper assessment passed with the expected
Cuellar Labs LLC Developer ID and hardened runtime. The separately signed DMG was also accepted (submission
`5f431bd9-5afb-44c2-a6f0-98bf088291ba`), stapled, accepted by Gatekeeper as
Notarized Developer ID, and verified by `hdiutil`. Apple ticket-service checks
retry transient failures at most three times and still fail on persistent errors.
All 21 Rust tests, strict Clippy, 28 site tests, five release-tool tests, and
Worker compile validation passed locally. Exact-commit hosted quality results
remain a release gate; ordinary CI does not establish that protected
Windows/Linux release jobs used credentials.
