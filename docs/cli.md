---
title: Envarium CLI contract
description: Commands, automation modes, JSON envelopes, and exit codes for the Envarium CLI.
category: reference
audience: [user, integrator]
platforms: [macos, linux, windows, wsl]
since: 1.0.0
lastVerified: 2026-07-17
searchTerms: [envarium command, jsonl, exit code, shell completion]
---

The `envarium` binary is both the daemon and its terminal client. Client
commands use private local IPC and start the same binary as a detached daemon
when no healthy daemon is already listening. The command result does not change
based on which process started the daemon.

## Project and operation commands

```text
envarium list
envarium add <repository>
envarium project list
envarium project get <id|unique-slug|path>
envarium project add <repository> [--allow-outside-roots]
envarium project trust <project> --yes
envarium project remove <project> --yes

envarium operation list [--project <project>] [--limit 100]
envarium operation get <operation-id>
envarium operation cancel <operation-id>

envarium manifest explain <project>
envarium manifest diff <project>
envarium manifest validate <project>
envarium open <project> [--print]
envarium ui [--path /projects/<id>]
envarium plugin list
envarium plugin refresh
envarium machine list
envarium machine show <machine>
envarium machine probe <machine>
envarium machine snapshot <machine>
envarium doctor
envarium doctor --bundle --preview
envarium doctor --bundle [--output <new-file.zip>]
envarium debug logs [--level debug|info|warn|error] [--limit 200]
envarium settings show
envarium settings export <new-json-file>
envarium settings apply <json-file> --yes
```

Selection checks opaque ID first, then an exact unique slug, then a canonical
repository path. Missing and ambiguous selections fail instead of guessing.
Catalog removal never changes repository files. Trust and removal require
`--yes`; Envarium does not hide an interactive confirmation inside automation
mode.

Project discovery is confined to the canonical roots shown in Settings. The
`--allow-outside-roots` flag is an explicit one-request exception: it permits
deterministic reads but does not trust the repository or execute its code. MCP
proposal creation has no outside-root override. Settings export writes a new
private full document; apply is revision-checked, strict, bounded, requires
`--yes`, and reports any fields waiting for daemon restart. See the
[durable settings guide](settings.md).

`doctor --bundle --preview` prints the exact support evidence allowlist and
denylist without writing a file. Removing `--preview` writes a new mode-`0600`
ZIP atomically and refuses to overwrite an existing path. The archive contains
only `manifest.json` and `internal-errors.ndjson`: build/API/schema identity,
adapter availability, sanitized daemon configuration, and at most 100 recent
redacted internal warnings or errors. It excludes repository source, database
contents, resolved credentials, environment values, terminal output, and
project application logs. `debug logs` reads the same dedicated two-segment
redacted control-plane log and never falls back to application output. See the
[support-bundle contract](support-bundles.md).

`ui` starts or attaches to the local daemon and prints a short-lived,
authenticated loopback URL. `--path` accepts only a local application route;
remote origins, fragments, dot segments, backslashes, and caller-supplied
bootstrap credentials are rejected. The native desktop adapter uses the same
command after its compatibility preflight.

`envarium desktop snapshot --json` is a versioned, bounded native-adapter
contract containing daemon identity, project runtime/health summaries,
workspaces, recent operations, diagnostic alerts, host pressure, and
port-conflict count. It is
read-only except that normal CLI attachment may start the bundled daemon when
none is running. It is not intended as a replacement for the richer public
query commands.

`plugin list` and `plugin refresh` perform read-only package discovery. Trust
requires the exact displayed fingerprint and `--yes`; enable is a separate
decision requiring reviewed `--scope` values and `--yes`. `plugin disable`
revokes every grant. `plugin health`, `plugin logs`, and `plugin inspect` expose
bounded supervision and structured observations. `plugin run` accepts one
advertised typed action, a JSON object, and `--yes`, then queues a durable
audited operation. No plugin command exposes an arbitrary shell.

Remote-machine commands are optional and preserve local-only operation:

```text
envarium machine add <name> <https-endpoint> --server-fingerprint <sha256>
  --ca <absolute-path> --client-certificate <absolute-path>
  --client-key <absolute-path> [--grant inventory.read] --yes
envarium machine access <machine> --grant inventory.read
  [--grant project.operate] [--grant environment.manage] --yes
envarium machine disable <machine> --yes
envarium machine remove <machine> --yes
envarium machine run <machine> <remote-project> <start|stop|restart|deploy|kill|rebuild>
  [--environment <remote-environment>] [--request-id <stable-id>] --yes
```

The CLI never accepts tunnel credentials or certificate values, only absolute
local certificate file references. Registration pins the reviewed server
certificate and starts with the specified complete grants. Every later grant
replacement and lifecycle operation requires `--yes`. See the
[federation guide](federation.md) for the separate agent-listener flags and
trust model.

Signed shared configuration is optional and remains independent of federation:

```text
envarium team publisher list
envarium team publisher trust --name <name> --public-key <base64> --yes
envarium team key generate --output <new-private-key-file>
envarium team bundle sign <kind> <bundle-id> --name <name> --version <version>
  --payload <json> --key <private-key-file> --output <new-bundle-file>
envarium team bundle install <bundle-file> --yes
envarium team template render <bundle> --set name=value
envarium team policy
envarium team registry
envarium team sync key-generate --output <new-age-identity>
envarium team sync export --recipient <age-recipient> --output <new-age-file>
envarium team sync preview <age-file> --identity <age-identity>
envarium team sync import <age-file> --identity <age-identity> --preview-sha256 <reviewed-digest> --yes

envarium telemetry status
envarium telemetry enable --endpoint <https-url> --yes
envarium telemetry send
envarium telemetry disable
```

Private signing keys and age identities are written only as new owner-only
files and never enter the daemon. Sync decrypts in memory and the preview lists
signature-verified bundle IDs and replacements before import. `telemetry
status` shows the complete current payload; telemetry is disabled and records
nothing until explicit opt-in. See the
[team configuration guide](team-configuration.md) and [privacy statement](privacy.md).

## Diagnosis and safe automation

```text
envarium diagnose <project> [--provider <configured-provider>]
envarium diagnose latest <project>
envarium diagnose feedback <diagnosis> <hypothesis>
  --verdict accurate|false_positive [--note <local-note>]
envarium diagnose run <diagnosis> <action> --yes
envarium diagnose notifications [project] [--all]
envarium diagnose notifications acknowledge <notification>

envarium automation list [project]
envarium automation create <project> <action>
  --name <name> --trigger <deterministic-code>
  [--cooldown 3600] [--max-per-day 3]
envarium automation enable <recipe> --yes
envarium automation disable <recipe>
envarium automation evaluate <project>
```

Diagnosis always collects a bounded redacted bundle and runs deterministic
rules before an optional provider. Provider output can cite only evidence and
accepted non-destructive actions present in that bundle. `diagnose run` reloads
the durable result and queues only a cited accepted action through the normal
operation kernel; it cannot be used as an arbitrary action or shell launcher.

Feedback is stored locally and is never sent as telemetry. Notifications are
deduplicated local records for crashes, ports, resource pressure, and unhealthy
dependencies. Cleanup findings are previews only.

Automation recipes are created disabled. Enabling requires a separate `--yes`
review, and evaluation reacts only to deterministic trigger codes. Every recipe
has a cooldown and UTC daily limit, can be listed or disabled at any time, and
may run only read-only or declared test/check/inspect actions. Operation IDs are
returned for every dispatched action so outcomes remain inspectable and
audited.

## Runtime commands

Trusted Compose and native-process projects expose live runtime queries and
durable lifecycle operations:

```text
envarium status <project>
envarium plan <start|stop|restart|deploy|kill|pause|unpause|rebuild|teardown> <project> [--profile name]
envarium start|deploy|rebuild <project> [--profile name]
envarium stop|restart|kill|pause|unpause <project>
envarium teardown <project> [--volumes] --yes
envarium logs <project> [--service name] [--since 10m] [--run id] [--operation id] [--tail 200]
envarium logs <project> [--service name] [--run id] [--operation id] --export plain|ndjson
envarium metrics <project> [--service name]
```

## Ports, Git, and trusted actions

```text
envarium ports list
envarium ports next [--range 15000-19999] [--protocol tcp|udp]
  [--project <project>] [--exclude 15001,15002]
envarium git <project>
envarium action list <project>
envarium action run <project> <action> [--yes] [--allow-outside-root]
```

`ports list` combines accepted declarations, persistent stopped-project
reservations, attributed runtime bindings, and unmatched OS listeners. Every
row reports provenance, and conflict state is derived before startup. `ports
next` considers all of those facts plus explicit exclusions; `--project`
ignores that project's declarations and reservations, but never ignores a live
binding.

`git` is a fresh, read-only porcelain-v2 snapshot with branch or detached
state, change categories, ahead/behind, stashes, last commit, remotes,
merge/rebase state, and worktrees.

Actions are available only for trusted projects. Built-ins include terminal,
VS Code, Codex, Claude Code, and Git pull; accepted endpoint and manifest
actions add browser, test, migration, and project-specific commands. `action
run` returns a durable `action.run` operation immediately. Use `operation get`
to read its outcome. Destructive actions require `--yes`. Working directories
are confined to the trusted root after symlink resolution unless the caller
also supplies the explicit `--allow-outside-root` permission.

Lifecycle commands return a durable operation immediately. Read it with
`envarium operation get <id>` or list recent project operations. `stop`
preserves containers and volumes. `teardown` is destructive, requires `--yes`,
and removes volumes only when the reviewed plan and operation include
`--volumes`.

`deploy` reconciles the trusted current definition: Compose pulls newer images,
force-recreates declared services, removes orphans, and preserves volumes;
native-process services restart from their current manifest definition. `kill`
force-stops selected services without a graceful timeout, while retaining the
same ownership checks and operation audit trail.

Compose discovery records profile names without adding their services to the
default topology. `start` and `rebuild` may opt into one or more names declared
by the trusted manifest with repeatable `--profile` flags. The project-detail
Start control presents the same allowlisted choices. A project-wide `stop` and
`teardown` select all Compose profiles, so optional services started during the
run are not left behind.

Native process runtimes support `start`, `stop`, `restart`, `deploy`, and `kill`. They reject
container-only pause, unpause, rebuild, teardown, and volume semantics. Start
plans show executable, argument array, working directory, environment-key
count, and keychain-reference count without revealing environment values.

`status` includes project ownership, service state, ports, and container or
process identity. Compose observations also include Docker context and Engine
versions. Process observations include verified PID/run fingerprints, restart
count, exit code, and last-known process identity. Logs support JSON Lines and
retain stdout/stderr plus project, service, source, and run identity. Metrics
return current CPU and memory samples; Docker also reports network counters.
Log reads come from the redacted persistent archive, not a daemon-lifetime
buffer. `--since` accepts a positive Go duration or RFC 3339 timestamp. Run and
operation filters retain lifecycle correlation in both queries and exports.

Daemon log bounds and extra redaction rules are configurable without changing
project manifests:

```text
envarium daemon \
  --log-ring-entries 2000 \
  --log-segment-bytes 1048576 \
  --log-retention-age 168h \
  --log-retention-bytes 268435456 \
  --redact-pattern 'company-secret-[A-Za-z0-9]+'
```

## Workspaces and worktree environments

Worktree discovery is an explicit post-trust operation. Registration reads Git
administrative metadata, projects monorepo subdirectories into every checkout,
and persists stable environment identities and isolation allocations:

```text
envarium environment list <project>
envarium environment register <project>
envarium environment get <environment-id>
envarium environment hostname <environment-id> <name.localhost>
envarium environment routes
```

Workspaces are reviewed JSON graph documents. Members may be trusted project
IDs or registered environment IDs. Dependencies must be acyclic; profiles can
select a dependency-closed subset and cap parallelism.

```text
envarium workspace list
envarium workspace get <workspace-id>
envarium workspace create <definition.json>
envarium workspace update <workspace-id> <update.json>
envarium workspace delete <workspace-id> --yes
envarium workspace start <workspace-id> [--policy rollback|continue]
  [--profile <profile-id>] [--run-recipes]
envarium workspace stop <workspace-id> [--profile <profile-id>]
  [--remove-data --yes]
```

Start and stop return durable operations immediately. Start follows dependency
order while running independent branches in parallel; stop reverses that
order. Rollback and continue policies report per-member outcomes. Stop
preserves data by default. `--remove-data` is destructive and is rejected
without `--yes`.

Friendly routing is disabled unless the daemon is started with a loopback-only
`--routing-address`. It accepts HTTP targets only, reports unavailable and
conflicting routes explicitly, and returns `503` for unknown hostnames. Local
TLS is not implied by this option.

## Automation modes

All query and mutation commands accept global flags after or before the command:

```text
--json             one indented versioned envelope
--jsonl            one compact envelope per list/stream item
--non-interactive  guarantee that no prompt is attempted
--no-color         suppress ANSI output
```

Envarium currently emits no ANSI sequences when output is redirected, and
all Phase 4 commands are prompt-free. JSON has this top-level shape:

```json
{
  "schemaVersion": "envarium.cli/v1",
  "command": "project.get",
  "data": {}
}
```

Errors are written to stderr as one object when `--json` or `--jsonl` is set:

```json
{
  "schemaVersion": "envarium.cli/v1",
  "error": {
    "code": "PROJECT_NOT_FOUND",
    "message": "project ... was not found by ID, slug, or path",
    "exitCode": 3
  }
}
```

`envarium schema cli <command>` prints the current draft 2020-12 envelope
schema and its generated OpenAPI data-model reference. Use `envarium
completion bash|zsh|fish|powershell` to generate shell completion.

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Success |
| 2 | CLI usage, flags, confirmation, or output-mode error |
| 3 | Requested project or operation was not found |
| 4 | Ambiguous selection, state conflict, or invalid proposal |
| 5 | Local daemon unavailable or failed to start |
| 6 | Internal or unclassified API failure |

Command packages only resolve input and render output. Catalog, manifest, and
operation behavior remains in application services behind the generated local
API.
## Offline data and manifest migration

`envarium data inspect`, `data backup`, and `data migrate` operate directly
on the configured data directory without starting the daemon. Inspection and
migration preview are read-only. `data migrate --write` requires the daemon to
be stopped and creates a verified non-overwriting pre-migration backup before
changing schema. `data backup --output <path>` refuses to overwrite its output.

Prelaunch manifest migration was retired before launch. Use deterministic
onboarding to produce a reviewed stable v1 manifest. Database backup and
migration commands retain the normal JSON envelope and semantic error behavior.
