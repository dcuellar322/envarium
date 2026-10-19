---
title: Desktop and headless installation
description: Install, upgrade, uninstall, or build Envarium for desktop and CLI-only use.
category: how-to
audience: [user]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-19
---

No public stable installers are published yet. Build from source on macOS or
Linux using the instructions below, or check [downloads](/download/) for package
availability. **Windows and WSL2 support is coming soon.**

## macOS desktop

Published desktop releases are expected to provide a signed, notarized DMG.
Open the DMG, drag Envarium to Applications, and launch it normally. The
application starts or attaches to the same per-user daemon used by the CLI.
The menu-bar tray remains available when the main window is hidden.

Use **Launch at login** in the tray only if Envarium should start for that
macOS account. Use **Keep running when window closes** to choose between
hide-to-tray and exiting the desktop adapter. **Quit Envarium** exits the
adapter but intentionally leaves an already running daemon available to CLI,
browser, MCP, and terminal sessions.

Release builds check the configured HTTPS release feed only when **Check for
Updates…** is selected. Every downloaded bundle must pass the embedded public
key signature check before installation. Debug and locally unsigned builds do
not enable updates.

Deep links use these bounded forms:

```text
envarium://project/<project-id>
envarium://workspace/<workspace-id>
```

An already running instance receives the link, verifies daemon compatibility,
opens the corresponding local route, and focuses the existing window.

## Linux desktop

The release workflow builds AppImage, deb, and rpm bundles; public stable
packages are not available yet. Use the native package
for integration with the desktop application menu and updater; make an
AppImage executable before launching it. WebKitGTK 4.1 and a supported Secret
Service implementation must be available. Native notifications and autostart
use the desktop environment's standard APIs through Tauri; a headless session
may report those capabilities unavailable while the Go control plane remains
usable.

External terminal handoff checks `TERMINAL`, then
`x-terminal-emulator`, GNOME Terminal, Konsole, and kitty. Embedded terminals
use a real PTY and do not depend on an external emulator.

## Windows desktop

**Coming soon.** Planned releases will provide signed MSI and NSIS installers. Verify the publisher
with `Get-AuthenticodeSignature` before installation. The daemon uses an
owner-only named pipe for privileged local clients, Job Objects for managed
process trees, and ConPTY for embedded sessions. Windows Terminal (`wt.exe`)
is required only for external terminal handoff; embedded ConPTY sessions do not
require it. VS Code handoff uses `code.cmd` when the VS Code shell command is
installed.

The updater verifies its independent Tauri signature in addition to Windows
package signing. Launch at login uses the per-user startup integration and does
not require administrator privileges.

## WSL2

**Coming soon.** The planned setup runs the Linux binary inside WSL and uses the browser UI it prints. Windows and
WSL local IPC endpoints are intentionally separate. See
[platform support](platform-support.md#wsl2-behavior) for Docker, filesystem,
notification, and autostart behavior.

## Build from source

On macOS or Linux, install the Go version from `.go-version`, Node from
`.node-version`, pnpm from `package.json`, and Rust from `rust-toolchain.toml`.
Clone the repository, then run:

```bash
pnpm install --frozen-lockfile
make desktop-quality
make desktop-build
```

The native bundles are written under
`desktop/src-tauri/target/debug/bundle/`. A source build is intentionally
unsigned and has no update authority. Maintainers use the protected
`desktop-release` GitHub environment to supply Apple signing/notarization,
Windows signing, and Tauri updater credentials; secrets are never committed or
placed in an application config file.

## Headless or CLI-only use

The Go binary is the supported installation for SSH hosts, CI workers, and
developers who do not want a native shell:

```bash
make build
mkdir -p "$HOME/.local/bin"
install -m 0755 bin/envarium "$HOME/.local/bin/envarium"
envarium doctor
envarium ui
```

`envarium doctor`, `envarium ui`, and ordinary project commands start the
per-user daemon on demand. Offline `data` commands and commands such as
`version`, `schema`, and shell completion do not start it. No Tauri or Rust runtime is required by the built Go binary.

## Uninstall and data choices

Removing `Envarium.app` removes only the desktop adapter. It does not delete
repositories, containers, volumes, project state, logs, or the daemon database.
This is the safe default and allows a later reinstall or continued CLI use.

Default durable control-plane data locations are:

```text
macOS:   ~/Library/Application Support/Envarium/
Linux:  ~/.config/Envarium/
Windows: %AppData%\Envarium\
```

Desktop-only preferences are under the per-application Tauri config directory
for `dev.envarium.desktop`. Launch-at-login can be disabled from the tray
before uninstalling; macOS Login Items can also remove it afterward.

Homebrew cask drafts are generated for maintainer review; no published cask is
currently offered. If installed from a reviewed cask, its uninstall definition uses that same bundle identifier for preferences,
WebKit data, caches, and saved window state. Ordinary uninstall retains data;
`brew uninstall --cask --zap envarium` additionally removes those files and the
`Envarium` control-plane data directory. Review backups and stop the daemon
before explicitly requesting zap.

For a complete removal, first stop active project/workspace operations and
quit the desktop adapter. Stop the daemon gracefully (for example, terminate
the `envarium daemon` process from Activity Monitor), then choose one of:

1. Keep the `Envarium` data directory for a future reinstall.
2. Move it to a private backup and verify the backup before deletion.
3. Permanently delete it only after reviewing `envarium.db`, `logs/`, and
   runtime metadata. This does not remove Docker volumes or modify repositories;
   those remain explicit project-runtime decisions.

Never remove the data directory while the daemon is running. A newer database
is rejected by an older binary rather than being silently downgraded or
mutated.

## Upgrade, downgrade, and reinstall

An upgrade verifies the existing schema, makes a non-overwriting consistent
backup before the first migration, and preserves projects, manifest snapshots,
audit history, operations, workspaces, plugins, diagnostics, and settings.
Use `envarium data inspect` before and after the upgrade.

Downgrade is restore-based: keep the current database, copy a pre-migration
backup to a separate data directory, and run the matching old binary against
that copy. Reinstalling the same or newer compatible v1 bundle attaches to the
retained daemon data. See [v1 migration](migration-v1.md) for the exact safe
procedure and [release engineering](release.md) for the required release checks.

Envarium uses its own application identity, data directory, and `envarium://`
links. Importing data or deep links from another application is not supported.
