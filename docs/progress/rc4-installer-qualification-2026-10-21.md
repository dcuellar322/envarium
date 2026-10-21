---
title: RC4 installer qualification — October 21, 2026
description: Security dependency rebuild, native app updater evidence, and signed-version replay blocker.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-21
---

## Publication hold

**Do not publish RC4.** The signed-version replay test accepts a genuine RC3
payload advertised as RC4 on both native macOS Intel and Linux. Its payload
signature is valid, but the version in the feed is unauthenticated. RC4 omits
`requireSignedVersion`, and its signer emits no version in the trusted comment.
[ADR-0032](../adr/0032-version-bound-updater-signatures.md) records the repair.
A new immutable signed candidate must contain the native policy and version-aware
signer; RC4's 31 assets and tag remain unchanged.

## Identity and completed checks

Candidate `v1.0.0-beta.rc.4` is built from product and workflow commit
`3fba51bb7440e56dd3645cba841ab1f8ab61ead3` by
[the protected signing workflow](https://github.com/dcuellar322/envarium/actions/runs/39554945509).
Required [core CI](https://github.com/dcuellar322/envarium/actions/runs/39553685524),
[site CI](https://github.com/dcuellar322/envarium/actions/runs/39553685539) and
[Security](https://github.com/dcuellar322/envarium/actions/runs/39553685635)
pass on that exact revision. Windows remains coming soon.

RC4 contains the Go 1.27.3, rustls-webpki 0.103.16 and webpki-root-certs 1.0.10
certificate-trust patches, plus Vite 8.2.4. All four actual CLI archives report
Go 1.27.3. Independent verification passes 29 aggregate SHA-256 entries, all
five updater payload signatures, the complete Mac/Linux package matrix, and
all 31 uploaded asset digests in the exact-source GitHub attestation.

Both Mac disk images and apps pass signature and stapled-ticket validation;
Intel Gatekeeper, version, bundle identifier, team and architecture checks pass.
Apple Silicon is inspected statically on this Intel host. All actual Intel Mac,
Linux x86_64 and emulated Linux arm64 CLIs pass fresh/schema-3/schema-17 migration,
newer-schema refusal, backup retention, plugin trust/revocation/fingerprint,
MCP permission/mutation and persistent agent-audit checks. Signed deb packages
pass Ubuntu 22.04 and Debian 12 installation/removal/reinstall; Fedora 44 rpm
passes the same checks, with populated records retained.

Actual native GUI smoke checks pass for Ubuntu and Debian deb, Ubuntu extracted
AppImage, Debian AppImage through the actual extraction-and-run runtime, and
Fedora rpm under Xvfb software compositing. They cover rendered
WebKit UI, palette/settings, hide-on-close, single-instance reopening and bundled
daemon identity/schema. Fedora extracted AppImage still fails its headless
rendering check. Debian's direct extracted-AppRun attempt in the extended notification/portal
environment fails to reach usable UI; a subsequent run of the actual AppImage
runtime with a fresh XDG runtime directory passes all GUI smoke checks. AppImage desktop/FUSE/GPU behavior remains a manual acceptance gate.

A staged actual RC2-to-RC4 Mac native updater installation passes signature,
strict codesign, bundled-daemon relaunch and exact populated database rows.
The original 32-case matrices pass on both native macOS Intel and Linux, including
byte/signature/key tampering and valid installation. The extended 35-case matrices
fail at the added real RC3 replay test. They are not passing 35-case qualification.

## Actual Linux app-driven updater

The unmodified, signed RC3 AppImage runs as an ordinary user in disposable
Debian 12, using its unchanged production hostname and public updater key.
A local HTTPS fixture on container loopback serves the real RC4 signature and
payload; only that container's hostname resolution and CA trust are changed.
They are restored afterward and temporary TLS private keys are removed. No
host trust, production DNS/feed, app binary or signing key is modified.

The test dispatches the production tray menu's D-Bus event through a simulated
StatusNotifier host. The real notification-daemon receives update notices.
The application rejects altered payload bytes and retains its installed bytes
and database rows, then accepts the genuine payload, writes the exact official
RC4 AppImage and invokes its own automatic GUI restart. The restarted native
app reports RC4 and renders the populated UI. All exact rows remain: four
projects, five manifest snapshots, 32 audit events, one settings record and
two settings audit records. AppImage extraction-and-run avoids FUSE, so this
is not FUSE acceptance or a visible desktop tray-host certification.

Functional assertions pass repeatedly. The final complete run passes with clean
teardown, restored container DNS/trust and removed TLS private keys. Earlier
harness teardown failures while inspecting hardened-process environments or
closing an X display are retained as fixture diagnostics.

The test confirms the compatible RC3 daemon remains running after the RC4 GUI
restart. Updating the GUI does not refresh that already running process. Existing
candidate users must review active projects, quit the GUI, stop their owned old
daemon safely, and relaunch to load the patched Go sidecar. Confirm the running
version with `envarium --json desktop snapshot` (`data.system.version`). The
test stops only its isolated daemon; no owner project or production process is
stopped. A fresh RC4 installation starts the patched daemon directly.

## Remaining acceptance

The replay repair and actual-byte qualification of the next candidate are
required before considering publication. Clean/quarantined Mac installs and
native UI on Intel/Apple Silicon (including macOS 13), normal GPU-backed Linux
X11/Wayland sessions, AppImage/FUSE/Fedora rendering, native tray/notification/
login/secret-store/handoff/accessibility behavior, privileged deb/rpm updater
authorization, and retained-data uninstall/reinstall remain manual gates.
The owner's RC3 Apple Silicon success does not qualify new candidate bytes.

Detailed receipts and scripts are retained in `.cache/rc4-qualification/`.
The production stable endpoint still returns HTTP 204; no release is published.
