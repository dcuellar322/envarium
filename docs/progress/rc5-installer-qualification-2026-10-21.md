---
title: RC5 installer qualification — October 21, 2026
description: Version-bound signed updater, Linux GUI and installer evidence, and remaining human acceptance gates.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-21
---

## Publication status

**Keep RC5 private pending the remaining acceptance checks.** RC3 is held for
certificate-trust advisories; RC4 is held for acceptance of an older signed
payload under a newer feed version. Their tags and assets remain immutable.
Windows and WSL2 remain coming soon and do not block this macOS/Linux release.

RC5 product tag `v1.0.0-beta.rc.5` identifies
`ec898755eae91aa9571d1ad322956f5a626fd488`. The reviewed packaging adapter is
`eb49bc21c3f8c343958f969d06fa27147421c89a`, pinned by
`v1.0.0-beta.rc.5-packaging.1`. The protected
[signed rebuild](https://github.com/dcuellar322/envarium/actions/runs/39570180023)
uses the original product tag. Product version, embedded commit and updater
signature version stay `1.0.0-beta.rc.5`; RPM package metadata is
`1.0.0-0.beta.rc.5`. See [ADR-0032](../adr/0032-version-bound-updater-signatures.md)
and [ADR-0033](../adr/0033-rpm-prerelease-metadata.md).

Required exact-product [core CI](https://github.com/dcuellar322/envarium/actions/runs/39561316744),
[site CI](https://github.com/dcuellar322/envarium/actions/runs/39561316640) and
[Security](https://github.com/dcuellar322/envarium/actions/runs/39561316669) pass.
The separate packaging revision also passes
[core CI](https://github.com/dcuellar322/envarium/actions/runs/39567364322),
[site CI](https://github.com/dcuellar322/envarium/actions/runs/39567364287) and
[Security](https://github.com/dcuellar322/envarium/actions/runs/39567364243).

The signed rebuild passes all jobs and creates a draft prerelease with 31 assets.
Independent checks verify all 29 aggregate SHA-256 entries, all five updater
payload signatures against the committed public key and signed RC5 version,
and every uploaded asset digest. GitHub's verified attestation identifies the
exact packaging commit/ref above; `metadata.json` identifies the original product
commit/tag. The qualified CLI workflow bytes match the final uploaded archives.
The existing unused GLib VariantStrIter advisory remains tracked in the
[security report](security-cleanup-2026-10-18.md#remaining-external-acceptance);
passing policy checks do not mean zero transitive advisories.

The first RC5 build failed before creating a release: the updated RPM library
rejected a hyphen in RPM Version, and Intel's independent app ticket validation
encountered Apple CloudKit network loss. The packaging adapter normalizes only
RPM metadata and permits bounded retries only for stapler's network exit 68.
Missing tickets, signature failures and cancellation still fail closed.

## CLI and installed Linux packages

The native qualification host is Intel macOS 15.8.1. Linux GUI checks run in
disposable containers without host data or Docker-socket access.

All four actual CLI archives pass workflow checksum and identity verification
and contain Go 1.27.3. Executed Intel macOS, Linux x86_64 and emulated Linux
arm64 CLIs pass fresh schema-19 startup, schema-3/schema-17 migration and backup,
offline inspection, private backup/non-overwrite, newer-schema refusal in
human/JSON preview/write modes, plugin trust/grants/revocation/fingerprint,
MCP observe mutation denial, scoped develop start/stop and persistent agent audit.
Apple Silicon CLI inspection is static on this Intel host.
Both Mac DMGs/apps pass signature, stapled-ticket, bundle/team/version and
architecture checks. Intel Gatekeeper passes; both updater archives contain the
correct architecture and Go 1.27.3 sidecar. All actual Linux desktop sidecars
also contain Go 1.27.3: deb/RPM binaries are identical, and AppImage's ELF rewrite
preserves the identical `.go.buildinfo` section. The ordinary `go version -m`
reader cannot map that rewritten AppImage layout; direct section inspection
and execution of the actual sidecar supply the evidence instead.

Ubuntu 22.04 and Debian 12 deb installations and Fedora 44 RPM installation pass
the same data/plugin/MCP checks, removal/reinstall and reattachment to populated
data. Removing the previous private candidate and installing RC5 also retains
its populated project/plugin/audit/settings records on all three distributions.
The installed RPM reports Version `1.0.0`, Release `0.beta.rc.5`, and its actual
native executable contains the RPM bundle marker. RPM internal digests and the
detached release-key signature verify. Ubuntu and Debian verify all three Linux
detached GPG signatures. Detached signatures do not create a trusted APT/RPM
repository or an embedded RPM header signature.

The first deb test script reused an old AppImage extraction directory and
stopped after its package/data checks. A fresh candidate directory passes
ordinary-user extraction and actual bundled RC5 sidecar checks on both systems;
the initial fixture failure remains in the receipts.

## Actual Linux GUI and app-driven update

All six native GUI combinations pass under Xvfb/Openbox as UID 1000:

| Distribution | Installed package | Actual AppImage runtime |
| --- | --- | --- |
| Ubuntu 22.04 | deb: pass | pass |
| Debian 12 | deb: pass | pass |
| Fedora 44 | RPM: pass | pass |

Each check exercises rendered production WebKit UI and accessibility tree,
keyboard palette, Settings navigation, hide-on-close, single-instance reopening
and the bundled RC5 daemon/schema identity. Settings and post-update screenshots
are visually inspected. Software compositing is enabled; no explicit WebKit
security-sandbox override is set. AppImage uses `APPIMAGE_EXTRACT_AND_RUN=1`.
Fedora AppImage now passes; the older candidate's headless rendering failure
does not describe this candidate. Physical GPU, Wayland and FUSE remain untested.

An unchanged, signed RC3 AppImage uses its production updater hostname and key
inside disposable Debian. The actual tray D-Bus event invokes the production
updater, with a real notification-daemon and simulated StatusNotifier watcher.
A local HTTPS fixture changes trust and hostname resolution only inside that
container. It rejects tampered RC5 bytes without changing the installed app or
data, then installs the exact official RC5 AppImage and automatically restarts
the actual GUI. The restarted native app and bundled CLI identify RC5. It then
rejects both an old/unversioned RC3 payload advertised as RC6 and a genuine
version-bound RC5 payload relabeled as RC6, leaving installed bytes and rows
unchanged. All exact rows remain: four projects, five manifest snapshots,
32 audit events, one settings record and two settings audit records.
The full test and teardown pass; container trust/DNS are restored and temporary
private TLS keys are removed. Host trust, production DNS/feed and signing keys
are unchanged.

The compatible RC3 daemon remains running after GUI restart. Existing preview
users must review active projects, quit the GUI, safely stop their owned old
daemon, and relaunch before the patched Go sidecar is active. Check the running
version with `envarium --json desktop snapshot` (`data.system.version`). Fresh
RC5 installations start the patched daemon directly. Qualification stops only
its isolated daemons and does not interrupt owner projects.

## Native updater matrices and Mac upgrade

Both 35-case matrices pass on native macOS Intel and Linux: 70 cases total.
They use the exact pinned native updater, release configuration/public key and
real candidate payloads in a mock application host. All 417 helper dependency
versions/sources match the product lock. Each host verifies both Mac archives
and Linux AppImage, including valid downloads, genuine old RC3 replay rejection,
altered/truncated bytes, altered/missing/foreign signatures, and positive
foreign-key controls. Other cases cover tampered cached installation,
rejection before changing installed bytes/modes, HTTP and untrusted TLS refusal,
204/same/older/no-target behavior, and a final successful native installation.
The helper's injected private fixture CA is client-scoped and removed afterward.

The production target mapper and byte/version-authenticated candidate signatures
also pass native default selection for AppImage, deb and RPM, signed downloads
and cached-tamper rejection before installer invocation. Only the test-helper
bundle marker is varied; signed product binaries are unchanged. This fixture
uses candidate RC5 metadata; stable-publication eligibility is covered separately
by the 45 passing site tests and does not make this private candidate eligible
for the stable feed.

A staged actual RC2-to-RC5 Intel Mac native update passes. The database is
byte-identical during app replacement; strict codesign, executable mode,
new bundled-daemon relaunch and all exact populated rows pass afterward.
This exercises native installation and daemon relaunch; a clean native Mac GUI
relaunch remains part of human acceptance.

## Remaining human acceptance

1. Install the final RC5 DMGs from a browser/quarantined download on clean Intel
   and Apple Silicon Macs, including macOS 13. Exercise the native GUI and
   upgrade/relaunch. The owner's RC3 Apple Silicon success does not qualify RC5
   bytes; signatures and static inspection do not replace this check.
2. Run RC5 in normal GPU-backed Linux X11 and Wayland desktop sessions. Verify
   AppImage/FUSE mounting, desktop launcher integration and rendering, including
   Fedora. The passing extraction-and-run tests bypass FUSE.
3. Verify visible native tray behavior, notification permission/delivery,
   autostart/login items, secret-store prompts, editor/terminal/browser handoff,
   keyboard navigation and screen-reader behavior on the supported desktops.
   The simulated tray host and inspected accessibility tree cover only a subset.
4. Exercise deb/RPM updater installation with a real desktop authorization
   prompt, including denial/cancellation and retry. Automated format selection
   and signed download do not qualify privileged interactive installation.
5. Accept the real desktop uninstall/reinstall experience: retained data,
   preferences/login-item cleanup and a separately restored older-version data
   copy. Existing private RPM previews with malformed historical headers may
   require explicit removal/reinstall while retaining data; do not force a
   downgrade of newer data.

Detailed scripts, logs, JSON receipts and screenshots remain in
`.cache/rc5-qualification/`. Final artifacts are staged separately under
`.cache/release-candidates/v1.0.0-beta.rc.5/rebuild-39570180023/` so no bytes from
the failed first build enter final qualification. Both native clients confirm
the live canonical stable feed returns no update under ordinary public TLS
trust. GitHub lists no published releases. All three qualification containers
and their temporary dependency snapshots are removed after evidence collection.
