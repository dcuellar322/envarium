---
title: Add your first project
description: Discover, review, approve, and operate a local repository with Envarium.
category: tutorial
audience: [user]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-10-19
sidebar:
  order: 2
searchTerms: [getting started, add repository, first project, project discovery]
---

[Build from source on macOS or Linux](/download/cli/#build-from-source)
until public installers are available. Windows and WSL2 are coming soon.
Then add one repository:

```bash
envarium doctor
envarium project add /absolute/path/to/project
envarium project list
envarium manifest explain <project>
envarium ui
```

Discovery is deterministic and does not execute repository commands. Review
the proposal and unresolved fields in the browser before accepting it. Add a
portable `.envarium/project.yml` when discovery cannot express intentional
runtime, health, port, or action configuration; start from the examples under
`examples/projects/` and validate it with `envarium manifest validate` after
the project is registered.

After trust and acceptance:

```bash
envarium plan start <project>
envarium start <project>
envarium status <project>
envarium logs <project> --tail 100
envarium ports list
```

Use `--json` or `--jsonl` for automation and `envarium schema <command>` for
the stable envelope. The browser URL printed by `envarium ui` contains a
short-lived bootstrap credential; do not paste it into issues or logs.

If Docker is unavailable, process projects and non-Docker features continue to
work. If any observation is partial or stale, Envarium reports that state
instead of fabricating readiness. See [troubleshooting](troubleshooting.md),
[platform support](platform-support.md), and the [CLI reference](cli.md).
