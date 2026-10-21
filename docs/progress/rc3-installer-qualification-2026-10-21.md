---
title: RC3 installer qualification — October 21, 2026
description: Native Linux GUI, isolated signed updater, data preservation, and remaining manual release gates.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-21
---

Candidate: `v1.0.0-beta.rc.3`, product commit
`7ffbbd5eaaa7fd5c206efe6fe4dc4dc7672ed53c`. The
[signed build](https://github.com/dcuellar322/envarium/actions/runs/39499086924)
uses packaging revision `68e38ed70929070c79e9260dc7c521d99b2c9512` at
`refs/tags/v1.0.0-beta.rc.3-packaging.1`. Both revisions passed required CI.
The candidate remains a draft prerelease. Windows/WSL2 remains coming soon.

## Results

| Target | Verified | Remaining limit |
| --- | --- | --- |
| macOS Intel, macOS 15.8.1 | Signed/notarized DMG/app; bundled daemon/data/plugin/MCP checks; six browser workflows; staged reinstall; native updater library download/verify/install | Clean-machine Tauri GUI, macOS 13, native integrations and updater UI/restart |
| macOS Apple Silicon | Signature/ticket/Gatekeeper/version inspection; owner reports RC3 works | Owner's exact OS/version and detailed native integration/update coverage not supplied |
| Ubuntu 22.04 | Signed deb install/remove/reinstall; actual deb and extracted AppImage native GUI smoke | Real desktop/GPU/session and FUSE launch |
| Debian 12 | Same deb/AppImage checks; actual RC2 AppImage updated to RC3 and its extracted native GUI relaunched with populated data | Real desktop/GPU/session, FUSE and app-controlled restart |
| Fedora 44 rpm | Signed package install/remove/reinstall and native GUI smoke | Real desktop and interactive authorization |
| Fedora 44 AppImage | Signatures, extraction and CLI checks | Extracted native GUI fails EGL initialization under Xvfb; real Fedora desktop acceptance remains open |
| macOS Intel / Linux x86_64 CLI | Actual release binaries pass migration, plugin and MCP/audit qualification | Native desktop behavior is separate |
| Linux arm64 CLI | Actual release binary passes data/plugin/MCP checks under emulation | Native arm64 hardware |

The artifact set passes 29 aggregate SHA-256 entries, the complete macOS/Linux
package matrix, all five updater payload signatures, Linux GPG verification,
and all 31 uploaded asset digests in the verified GitHub attestation. Product
metadata and exact packaging workflow identity are checked independently.
The feed correction below does not alter or replace RC3 assets or tags.

## Native Linux GUI smoke tests

Tests run as an ordinary UID 1000 in isolated Ubuntu/Debian/Fedora containers,
using a private DBus session, Xvfb, Openbox, accessibility inspection, and the
actual signed release binaries. They verify:

- A visible native window and rendered production WebKit content, with screenshots.
- Projects/workspaces/settings navigation and the keyboard command palette.
- Settings view and exact bundled daemon version/commit/API/schema 19.
- Close hides the window while the process remains alive.
- A second application invocation reopens the existing instance.

Ubuntu/Debian deb and extracted AppImage, plus Fedora rpm, pass. Containers
use `WEBKIT_DISABLE_COMPOSITING_MODE=1` for software compositing. The initial
Fedora Xvfb run exposed accessible content but blank screenshots, so accessibility
alone was not accepted as rendering evidence. Software-rendered screenshots
were inspected. No WebKit security-sandbox override is used in the passing runs.
These sessions have no host Docker socket or full desktop notification/login/
secret-service environment; the UI correctly reports unavailable Docker data.

Fedora's extracted AppImage reports
`Could not create default EGL display: EGL_BAD_PARAMETER. Aborting...` and
never reaches usable WebKit content. Explicit CPU/DMABUF controls and a separate
sandbox-disabled diagnostic do not resolve it. The diagnostic is not a pass.
The same Fedora environment runs the rpm. This headless failure does not establish
how the AppImage behaves on a GPU-backed Fedora desktop; that test is required.
AppImage GUI tests elsewhere use extraction and do not qualify FUSE mounting.

## Isolated signed updater tests

The harness uses Tauri `2.12.4` and updater `2.14.0` at the exact production
revision `ba98bfadc4c868c9a7e1a226c83af528e1f11b89`. All 417 harness dependency
versions/sources occur in RC3's desktop lockfile. The app host uses Tauri's mock
runtime, while the real updater client, signature verifier and native installer
run unchanged. The install path is an explicit disposable app/AppImage copy.

The HTTPS fixture binds only to loopback. A temporary private CA is trusted
only by the harness client; TLS and hostname validation stay enabled. Fixture
TLS keys are removed afterward. Foreign signing controls use an ephemeral
private key held only in memory. The production endpoint, trust store, updater
public key, app installation and user data are unchanged.

Each native host, macOS Intel and Linux x86_64, passes a 32-case matrix:

- Accept genuine signed Mac Intel, Mac arm64 and Linux AppImage downloads.
- Reject altered/truncated payloads and altered, missing or foreign signatures.
- Confirm foreign signatures are valid under their own public key, then refuse
  them under the production key.
- Reject altered cached bytes through the install API before changing the
  installed application; rejected downloads/installations preserve its bytes
  and file modes.
- Enforce TLS trust and release-mode HTTPS; reject missing targets; return no
  update for HTTP 204, the same version and an older version.
- Download, verify and install the valid native RC3 payload.

A separate staged macOS RC2→RC3 update and the Linux AppImage update preserve a
populated RC2 database byte-for-byte during installation. After the updated
Mac daemon and actual updated Linux GUI relaunch, exact rows remain: four
projects, five manifest snapshots, 32 audit events, one settings record and
two settings audit records. The Mac app passes strict code-signature validation;
its stapled ticket validates. Updated app/AppImage executable modes remain 0755.

These checks do not invoke the production app's tray update action, notification
permissions, privileged package authorization, or `app.restart()` callback.
The real canonical-feed UI/install/restart path remains a manual gate.

## Linux updater catalog defect and repair

Native-format probes found that the catalog offered only a generic Linux
AppImage. Actual deb/rpm binaries carry their matching bundle markers. The
pinned updater selects the generic AppImage fallback and then rejects it with
`invalid updater binary format`, leaving the installed application unchanged.
This would block deb/rpm users from updating through the proposed stable feed.

[ADR-0031](../adr/0031-linux-updater-installer-formats.md) adds required signed
`linux-x86_64-deb` and `linux-x86_64-rpm` entries. The generic AppImage and both
Mac entries remain. Existing RC3 binaries support the format-specific entries;
no new binary or signing key is needed. Catalog and release-file validation
require all five authenticated payloads. Regression tests cover completeness,
missing signatures/targets and format-matched payload selection.

With a catalog generated by the corrected production mapper, the exact native
updater with each matching bundle marker selects and verifies the proper
AppImage/deb/rpm payload. Each also refuses altered cached bytes before install.
The marker probes modify only the disposable harness executable to match the
markers read from the real packages; signed release binaries are not patched.
The test uses an eligible stable-release metadata fixture to exercise the
catalog builder and actual candidate payloads over loopback; no stable release
or staging catalog is published.

## Manual acceptance still required

1. Clean quarantined macOS installs/native launches on Intel and Apple Silicon,
   including macOS 13. The owner's Apple Silicon result is recorded without
   assuming unspecified integration or updater tests passed.
2. Normal supported Linux desktop/package-menu launches, real GPU-backed X11
   and Wayland sessions, FUSE AppImage launch, and Fedora AppImage rendering
   without headless overrides.
3. Tray menus/quit/reconnect, notification delivery and denial, deep links,
   actual login autostart and cleanup, Keychain/Secret Service, terminal/editor
   handoff and screen-reader behavior in native views.
4. Actual app-driven signed update, notifications, package authorization,
   restart/relaunch and retained data through a reviewed strategy for the
   production-pinned endpoint. Cryptographic tamper rejection is automated;
   this full native integration flow is not certified by the mock host.
5. Representative older installations, restored-backup downgrade with the
   matching older binary, and ordinary OS uninstall/reinstall with populated
   data/preferences and login-item cleanup.

Detailed scripts, JSON receipts, screenshots and logs are retained locally in
`.cache/rc3-qualification-rebuild/extended-2026-10-21/`. Test-owned data and
installations are isolated. No production release was published.
