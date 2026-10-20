---
title: Platform support
description: Supported operating systems, packages, native capabilities, and WSL2 boundaries.
category: reference
audience: [user, contributor]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-20
---

The first release targets **macOS and Linux only**. macOS desktop bundles cover
Apple Silicon and Intel; Linux desktop bundles cover x86_64. CLI archives cover
arm64 and x86_64 on both systems. **Windows and WSL2 support is coming soon.**
Adapters exist in the source, but public support requires installer and
clean-machine validation. The table
below distinguishes current source-build targets from planned platform baselines.
Check [downloads](/download/) for published packages; no public stable binaries
are available yet.

| Platform | Source-build baseline or planned target | Release bundle targets | Source adapter capabilities |
|---|---|---|---|
| macOS | macOS 13 or newer | signed/notarized `.app` and DMG | Unix socket, process groups, PTY, port APIs, Keychain, notifications, launch agent |
| Linux | Ubuntu 22.04+/Debian 12+ and current Fedora-family desktops | AppImage, deb, rpm | Unix socket, process groups, PTY, portable port APIs, Secret Service, notifications, XDG autostart |
| Windows (coming soon) | Planned: Windows 11 and Windows Server 2022 with Desktop Experience | signed MSI and NSIS | owner-only named pipe, Job Objects, ConPTY, portable port APIs, Credential Manager, notifications, startup task |
| WSL2 (coming soon) | Planned: supported distributions on Windows 11 | use the Linux binary inside WSL | Linux socket/process/PTY behavior inside the distribution; explicit interop limits below |

FreeBSD and other Unix-like systems may compile through conservative adapters
but are not v1 release targets.

## Primary workflow matrix

Every release candidate must pass project discovery and approval, Compose and
native-process lifecycle, status/log/metric inspection, ports, actions,
terminal resize/reconnect/termination, MCP observe and authorized mutation,
plugin conformance, data backup/migration, desktop attach, autostart, native
notification, upgrade, uninstall-with-data-retained, and clean reinstall.

The GitHub `CI` workflow executes native adapter tests on macOS, Linux, and
Windows and builds the sidecar on all three. Windows jobs are nonblocking
source coverage; they do not qualify a Windows release or prevent macOS/Linux
release qualification. The release workflow builds only macOS/Linux bundles,
produces SBOMs, signs platform and updater artifacts, and creates attestations.
The primary workflow matrix above is required on macOS/Linux for the first
release; Windows/WSL2 qualification is deferred under
[ADR-0030](adr/0030-first-release-platform-scope.md). The operator checklist is in
[release engineering](release.md).

`make quality` includes Linux cross-compilation through `make platform-check`.
Windows cross-compilation remains available
separately with `make platform-check-windows` for contributors working on the
upcoming platform.

## WSL2 behavior

The following describes the intended behavior for upcoming WSL2 support.

Install and run the Linux `envarium` binary inside each WSL distribution
whose repositories it manages. Its data directory, Unix socket, child
processes, terminals, Docker context, and repository paths stay inside that
distribution's security boundary.

The Windows desktop does not attach directly to a WSL Unix socket, and a WSL
CLI does not attach to the Windows named pipe. Use `envarium ui` inside WSL
and open its loopback URL through Windows' localhost forwarding. Docker Desktop
integration works when `docker` and `docker compose` work inside that same WSL
distribution. Windows notifications and startup settings are not projected
into WSL; use the Linux desktop integration only when the distribution has a
supported graphical session.

Repositories under the WSL filesystem are recommended. `/mnt/c` repositories
work but inherit Windows filesystem performance, case-sensitivity, permissions,
and file-watching limitations. Envarium reports adapter failures rather than
claiming parity when the host disables localhost forwarding, systemd user
services, Secret Service, or Docker integration.
