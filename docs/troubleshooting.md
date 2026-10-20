---
title: Troubleshooting by symptom
description: Safe checks and bounded recovery steps for common Envarium runtime and platform failures.
category: troubleshooting
audience: [user, contributor]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-19
searchTerms: [Docker unavailable, port in use, logs unavailable, daemon disconnected, WSL localhost]
---

Start with `envarium doctor`, then `envarium debug logs --level warn` and
`envarium data inspect`. Preview `envarium doctor --bundle --preview`
before writing or sharing an archive. Keep diagnostic output private even
after automated redaction.

Use macOS or Linux for current source builds. **Windows and WSL2 support is
coming soon**; see [platform support](platform-support.md) for the intended
Windows and WSL2 behavior and current availability.

- **Daemon will not start:** another matching daemon may own `daemon.lock` or
  the Unix socket/named pipe. Do not delete a live lock. Stop the process shown
  in the lock or let normal stale-lock recovery run.
- **Docker disconnected:** verify `docker version`, `docker compose version`,
  and the selected Docker context as the same user running Envarium.
- **Terminal unavailable:** Linux needs a PTY-capable supported distribution;
  external launch also needs a supported terminal emulator.
- **Ports unavailable:** host policy may prevent process ownership details.
  Bindings without an accessible PID remain honest partial observations.
- **Manifest rejected:** unknown fields and unsupported versions fail closed.
  Only `envarium.dev/v1` manifests are accepted. Use project onboarding to
  review a replacement for an unsupported manifest version.
- **Database newer than binary:** install the matching/newer Envarium build or
  restore a pre-migration backup to a separate data directory. Never force an
  old binary to open it.
- **Plugin disabled after update:** executable or manifest fingerprints changed.
  Refresh, inspect the new capabilities/scopes, then explicitly trust again.

Issue reports should include platform, Envarium version, expected/actual
behavior, minimal reproduction, and only redacted diagnostic facts. Never
attach credentials, private source, browser bootstrap URLs, raw environment
files, or unreviewed application logs. The
[support-bundle contract](support-bundles.md) documents the exact archive
contents and residual privacy risk.
