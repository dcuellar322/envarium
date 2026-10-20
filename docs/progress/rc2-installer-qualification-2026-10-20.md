---
title: "RC2 installer qualification: October 20, 2026"
description: Artifact-level installer evidence, migration and plugin CLI defects, and the remaining human release gates.
category: contributor
audience: [maintainer, contributor]
platforms: [macos, linux]
lastVerified: 2026-10-20
---

## Decision

**Hold publication.** Qualification found reproducible migration-reporting and
plugin-setup CLI defects. The clean-machine native desktop and signed update
matrix also remains open. Windows/WSL2 is outside this release gate under
ADR-0030.

Candidate: `v1.0.0-beta.rc.2`, commit
`a8a87a6ee8cb18db3cde8adaeb1209686f34ab7a`.
The tag and signed artifacts were not changed. Testing used the downloaded
candidate binaries, not locally rebuilt product binaries.

This evidence maps to Phase 18's primary-workflow and preserved-data criteria,
the [release checklist](../release.md#release-candidate-checklist),
[platform matrix](../platform-support.md#primary-workflow-matrix), and
[installation procedures](../desktop-installation.md).

## Verified artifacts and distribution

- [Exact-commit CI](https://github.com/dcuellar322/envarium/actions/runs/39440832848),
  [site CI](https://github.com/dcuellar322/envarium/actions/runs/39440833152), and
  [Security](https://github.com/dcuellar322/envarium/actions/runs/39440832810)
  passed before the tag was built.
- [Signed release workflow](https://github.com/dcuellar322/envarium/actions/runs/39442884629)
  passed for both macOS architectures and Linux, including native signatures,
  notarization, updater signatures, SBOMs, and attestations.
- Downloaded release assets passed all 29 aggregate SHA-256 entries and all
  three updater signatures against the committed production public key. The
  complete macOS/Linux artifact matrix is present; Windows is not required.
- GitHub attestation verification of `checksums.txt` matched the exact source
  commit, tag, release workflow, and GitHub-hosted runner identity.
- Independent Linux GPG verification accepted AppImage, deb, and rpm detached
  signatures against fingerprint
  `2380BA10D0E6322DA0B3CD4D985DD308E9EE0B65`.
- The production stable feed returned HTTP 204. RC2 remains a draft prerelease
  and is not offered to installed stable clients.

## macOS Intel: macOS 15.8.1, x86_64

- The actual Intel DMG passed `hdiutil verify`, signature verification, and
  Gatekeeper disk-image assessment as a notarized Developer ID artifact.
- Mounted the DMG read-only and copied its app to a test-owned installation
  directory. The copied app passed deep/strict `codesign` verification,
  stapled-ticket validation, and Gatekeeper executable assessment.
- The app reports bundle ID `dev.envarium.desktop`, candidate version
  `1.0.0-beta.rc.2`, macOS minimum 13.0, and the `envarium` URL scheme. Its
  bundled control plane reports the candidate version and correct commit.
- All six existing browser workflows passed against the copied signed bundle's
  control plane and embedded production web assets: discovery/approval,
  terminal Unicode output/reconnect/termination, ports, native process
  start/logs/stop, workspace lifecycle, real Docker Compose lifecycle,
  worktree isolation/checkout transitions, settings, and accessibility checks.
  The test harness substituted only the candidate binary and isolated data
  paths. This tests the browser adapter; it does not qualify Tauri's WebKit view.
- MCP Inspector passed tools/resources/prompts discovery using the bundled
  control plane's observe profile; destructive teardown was not exposed.
- A separate real stdio MCP exchange verified observe-profile mutation denial,
  project-scoped develop-profile start/stop, successful operation waits, and
  two persisted audit events identifying the qualification agent.
- Replaced the staged app from the original DMG and reattached to retained
  data. Four projects, five manifest snapshots, 29 audit events, the settings
  record, and two settings audit records were preserved exactly.
- The standalone Darwin x86_64 CLI archive also passed the data checks below.

The existing account's Applications folder, production Envarium database,
desktop preferences, launch agents, and notification permissions were not
modified. This was a staged bundle copy/reinstall, not a Finder installation
and first launch in a clean macOS account.

## macOS Apple Silicon: artifact inspection only

The arm64 DMG passed disk-image integrity and Gatekeeper checks. Its mounted
application passed deep/strict code-signature and stapled-ticket validation;
the executable is arm64. The first local ticket validation encountered an Apple
CloudKit connection reset; the retry passed, consistent with the release CI.
No Apple Silicon executable was run on this Intel host.

## Linux: isolated package environments

The Ubuntu 22.04 and Debian 12 containers installed the actual deb with its
declared system dependencies. The native shell's shared libraries resolved
without missing dependencies. The packaged Go sidecar ran under an ordinary
user account, passed the data checks below, and reattached after package
removal/reinstall while retaining its database. AppImage extraction and the
extracted sidecar's candidate identity also passed under an ordinary user.

The Ubuntu AppImage harness initially extracted as root and then tried to run
that private extraction as another account. That invalid ownership setup failed;
extracting and running as the same unprivileged user passed. No artifact modes
or signatures were altered to make it run.

The standalone Linux x86_64 CLI archive passed the same data checks in a
separate Linux container.

Fedora 44 installed the actual RPM and resolved all native shell shared
libraries. Its packaged sidecar passed the same ordinary-user data checks;
RPM removal/reinstall retained the database and the reinstalled sidecar
reattached successfully. The first Fedora harness lacked `runuser`; installing
that test utility and repeating the unchanged artifact checks passed.

These are package/dependency and control-plane checks, not graphical desktop
qualification. They do not exercise Linux display servers, FUSE/AppImage
mounting, desktop menus, tray implementations, Secret Service, notifications,
login autostart, or a native updater installation.

## Data safety checks on the actual candidate binaries

The macOS bundle, standalone macOS/Linux x86_64 CLIs, and installed Linux
package sidecars were exercised with test-owned databases:

- Candidate identity, fresh daemon startup, API v1/schema 19, and empty catalog.
- Offline version command without creating data; offline schema inspection.
- Verified private mode-0600 backups; refusal to overwrite an existing backup.
- Schema 3 automatic startup migration and schema 17 explicit write migration
  to schema 19, including private non-overwriting pre-migration backups.
- Read-only migration previews leave source database bytes unchanged.
- Projects, accepted manifests, and audit events survive both upgrades; schema
  17 settings and settings audit history survive unchanged as well.
- Explicit migration refuses a lock while an upgrade is required.
- Restored pre-migration backups inspect at their original schema in separate
  data directories, leaving upgraded data intact.
- Daemon startup refuses simulated newer schema 999; database bytes remain
  unchanged. The CLI reporting defect below remains separate from that safety
  protection.

These use synthetic historical-schema fixtures, not previously installed
alpha/beta applications or a stable updater installation.

## Finding: migration CLI falsely reports newer data as current

Reproduced with both the macOS bundle/standalone CLI and Linux binaries:

1. Copy a test database and add an applied migration version greater than the
   supported schema (the qualification fixture uses 999; RC2 supports 19).
2. Run `envarium --data-dir <fixture> data migrate --write`.
3. Observe exit status 0 and `Database is current at schema 999.`
4. JSON output likewise returns `currentVersion: 999`, `targetVersion: 19`,
   `migrationRequired: false`, without a compatibility error.

The CLI only invokes the migration service when `migrationRequired` is true,
then labels every false value "current". That includes unsupported newer
schemas. See `internal/transport/cli/data_command.go`, lines 64–79, and
`internal/platform/sqlite/database.go`'s inspection status calculation.

No database corruption was observed: starting the daemon correctly refuses
the same fixture and preserves the database. Nevertheless, this is misleading
success in an operator safety procedure. Correct the incompatible-schema CLI
result, add a regression test for both human and JSON output, and qualify a
new immutable candidate before publishing. RC2 itself was not patched.

## Finding: documented plugin CLI setup fails idempotency validation

Built only the repository's reviewed sample plugin (the Envarium host remained
the downloaded, signed RC2 binary), placed it in the isolated daemon's plugin
directory, and restarted that daemon. `plugin list` discovered the package as
untrusted and disabled with its exact executable fingerprint.

Both documented setup steps fail in RC2:

- `envarium plugin refresh` exits 6 with `IDEMPOTENCY_KEY_INVALID`.
- `envarium plugin trust fixture-inspector --fingerprint <reviewed-sha256> --yes`
  exits 6 with the same error: "Use an opaque key between 8 and 128 characters."

The thin HTTP client calls `RefreshPluginsWithResponse` and
`TrustPluginWithResponse` without an idempotency request editor, while the
server requires a key for those mutations. See
`internal/transport/httpclient/plugins.go`, lines 14 and 35. Other plugin
mutation methods have the same missing-editor pattern, but only refresh and
trust were directly exercised here. Trust was never granted and the plugin
was never enabled; no subsequent plugin success is claimed.

Correct the CLI/client mutation boundary and add transport-level regression
coverage, then rerun discovery, trust, least-privilege grants, health,
inspection, operation completion, disable/revocation, and fingerprint-change
refusal with the next candidate. Existing SDK/service tests do not establish
that the published CLI setup path works.

## Remaining manual release gates

Before the manual matrix, correct both CLI findings and rerun the affected
artifact-level checks on a new candidate. Plugin protocol/conformance smoke
could not proceed through the documented CLI setup; this is a software failure,
not evidence that a human-only check would pass.

| Gate | Required evidence before publication |
|---|---|
| Clean macOS installation | On Intel and Apple Silicon, download normally with quarantine intact, open the DMG, drag to Applications, and launch through Finder. Include the macOS 13 baseline, not only the current host. Verify first-run UI and no sidecar/version mismatch. |
| Clean Linux desktops | On Ubuntu 22.04+, Debian 12+, and a current Fedora desktop, install the native package and launch through the application menu. Exercise AppImage through FUSE with normal user permissions. Cover supported Wayland/X11 sessions. |
| Native shell integration | Tray and hide/quit behavior, daemon attachment/reconnection, notifications and denial behavior, Secret Service/Keychain integration, enable/disable launch-at-login followed by a real login, project/workspace deep links, and external terminal/editor handoff. |
| Primary workflows in native views | Repeat discovery/approval, Compose/native lifecycle, logs/metrics/ports/actions, and terminal resize/reconnect/termination in the installed desktop environments. Verify the corrected plugin flow and provider-client integration across platforms. Intel MCP observe denial/develop mutation/audit is already verified; browser tests do not qualify native views. |
| Historical install/update matrix | Upgrade actual supported alpha/beta installations with nonempty projects/settings/manifests/audit history, exercise the fixed newer-schema refusal, and test an old matching binary against a separate restored backup. |
| Signed updater | Arrange a reviewed staging strategy: RC2 pins the production stable endpoint, which currently returns 204, so this cannot be completed by clicking Check for Updates today. Exercise candidate-to-stable download, signature enforcement, installation, relaunch, retained data, and rejection of tampered/foreign signatures through the native updater. Do not publish an unqualified stable feed to perform the test. |
| Ordinary uninstall/reinstall | Remove installed apps/packages through normal OS/package tools while retaining nonempty test-user data; reinstall and confirm attachment, preferences, repositories, and volumes. Verify login-item cleanup choices. Staged-copy/container checks provide supporting evidence only. |

## Reproduction artifacts

Test-owned scripts and extracted installations are under
`.cache/rc2-qualification/`; sanitized execution logs are retained under its
`evidence/` directory. No signing keys or bootstrap tokens are stored there.
The browser harness reuses `web/tests/e2e` scenarios with explicit candidate
binary/data paths and disables traces to avoid retaining browser credentials.
Documentation generation, source-content validation, `git diff --check`, and
Astro checks pass (105 files, zero errors/warnings/hints). No runtime source,
immutable release tag, or signed artifact was changed during qualification.
