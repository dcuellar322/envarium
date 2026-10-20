---
title: "Phase 18: Cross-platform hardening and open-source v1.0"
description: Implementation evidence for Envarium product phase 18.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-20
---

## Implemented

- Replaced platform assumptions with native Linux and Windows adapters: portable
  listener inspection, Linux terminal/editor launch, owner-only Windows named
  pipes, Job Object process-tree ownership, ConPTY terminals, Windows Credential
  Manager reads, and reviewed URL launch validation.
- Promoted the product, project manifest, plugin protocol, plugin manifest,
  generated schema, desktop shell, and examples to stable `1.0.0`/v1 contracts.
  Legacy alpha manifests normalize in memory and have an explicit preview-first,
  backup-preserving file migration command.
- Added offline database inspection, explicit backup, and preview/write migration
  commands. Automatic upgrades now verify the source and a private consistent
  backup before applying migrations, refuse newer schemas, and preserve projects,
  accepted manifests, audit history, settings, and all other durable records.
- Added native adapter and sidecar CI matrices, a cross-platform sidecar builder,
  GoReleaser archives/checksums/CycloneDX SBOMs, platform desktop bundles,
  Apple/Windows/Linux signing hooks, Tauri updater signatures, Sigstore bundles,
  GitHub artifact attestations, CodeQL, dependency review, and secret scanning.
- Added fail-closed nightly/alpha/beta/stable tag classification, exact-tag
  checkout for manual releases, prerelease marking, and a scheduled unsigned
  cross-platform nightly CLI snapshot with short artifact retention.
- Published the v1 compatibility/deprecation policy, platform and WSL behavior,
  data migration and restore-based downgrade procedures, release matrix,
  contributor setup, adapter guide, manifest reference, privacy policy, support
  policy, troubleshooting guide, threat model, and runnable Compose/process
  examples.

## Architecture decisions

No accepted dependency direction changed. This phase completes the native
adapter and platform-order boundaries in ADR-0002 and ADR-0015, durable
migration rules in ADR-0005, runtime and terminal process ownership in ADR-0007,
the thin signed desktop shell in ADR-0009, local IPC and browser security in
ADR-0013, and optional authenticated federation in ADR-0016. The stable
identifiers are compatibility promises over those accepted boundaries rather
than a new architectural dependency.

## Exit criteria

- [x] Primary workflows have native macOS coverage and required Linux/Windows
  host matrices; the release checklist records the manual installer matrix.
- [x] Upgrade fixtures prove projects, accepted manifests, audit history, and
  preferences survive; every write migration creates and verifies a backup.
- [x] The v1 security review has no unresolved critical or high findings.
- [x] CLI, manifest, HTTP, MCP, plugin, and product compatibility and deprecation
  policies are public in the repository.
- [x] A new contributor can install, generate, test, build, and run the project
  from documented commands, working example manifests, and the enforced
  §19.2 fixture inventory.

## Verification evidence

The complete local quality gate passed on macOS: generated contracts, Go and
Rust formatting, vet, GolangCI, architecture checks, all Go and race tests,
schema migration tests, `govulncheck`, Vue lint/typecheck/unit coverage, browser
workflow and visual suites, production web/Go builds, Clippy, Rust tests, and
debug desktop packaging. Linux and Windows adapter suites compile from the
local checkout and are required as native hosted jobs before merge. The
GoReleaser v2 configuration and every GitHub workflow pass their configuration
linters. A local v1 daemon smoke test now reports API v1/schema 17; database
upgrade coverage preserves the revisioned settings singleton and value-free
audit history alongside the earlier project, manifest, and operation records.
The smoke discovered and approved the repository deterministically, listed the
project, inspected the database, and shut down cleanly.

Release-channel unit tests reject unknown or injection-shaped tags and classify
nightly, alpha, beta, and stable forms deterministically. The manual release
path resolves an existing tag before any build job begins; scheduled nightly
artifacts never receive stable signing or updater authority.

Signed/notarized release artifacts are deliberately created only by a reviewed
`v*` tag in the protected release environment. The workflow refuses partial
publication and leaves a draft release for human inspection; no signing key or
production release was fabricated during local verification.

## 2026-07-17 cross-platform regression evidence

- The forced-stop integration helper now ignores both POSIX `SIGTERM` and
  Windows `os.Interrupt`, matching the CTRL_BREAK path used by the Windows
  supervisor and preserving the expected `stopped_forced` result.
- Native macOS adapter tests pass, and the Linux and Windows adapter packages
  cross-compile from this checkout. The corrected Windows behavior remains
  required on the hosted Windows runner.

## 2026-07-20 V1 hardening evidence

The V1 security pass closed the findings identified in local browser, process,
filesystem, transport, certificate, persistence, and static-analysis review:

- Browser sessions now have idle and absolute expiry, browser mutations require
  same-origin requests and constant-time CSRF validation, dynamic API responses
  are non-cacheable, correlation identifiers are bounded and validated, and
  non-loopback browser listeners fail closed.
- Plugin and agent executables are resolved through symlinks, confined to their
  approved roots, checked for regular executable files, fingerprinted again
  immediately before launch, and supervised as whole process trees on Unix and
  Windows. Terminal Compose file paths use the same confinement rules.
- Private keys are accepted only from regular owner-private files. Telemetry
  does not follow redirects, fleet TLS validation handles the complete peer
  chain, support bundles publish without overwriting existing files, and agent
  installer outputs cannot escape through symlinks.
- Unsigned-to-signed conversions and persisted counters are range checked or
  saturated. Log deletion uses a rooted filesystem handle, workspace SQL uses
  explicit statements, and GolangCI/gosec reports zero findings under the
  repository's documented boundary-specific exclusions.

The production-readiness and maintainability pass also addressed behavior found
under real UI load:

- Runtime observations use a one-second, per-project singleflight snapshot with
  clone isolation and lifecycle invalidation. Port registry observations use a
  ten-second coalesced snapshot while safety-sensitive port suggestions always
  refresh. The global sidebar no longer triggers expensive host-wide listener
  scans, and project port data loads only on the Ports tab.
- Workspace project states are translated exhaustively into the operation
  kernel's durable step states. This fixes a constraint failure discovered by
  the browser start/stop flow and is covered by regression tests.
- Nine oversized Vue views and panels were split into domain-specific
  composables/components, dead UI code and duplicate mockups were removed, and
  architecture checks now enforce the 250-line non-style Vue target. Prettier is
  pinned and enforced alongside ESLint and the generated API type boundary.
- Playwright uses one worker, strict ports, and isolated non-reused servers, so
  a developer's unrelated local application cannot silently satisfy a test
  server probe.

Fresh verification from this checkout passed:

- `make lint`, `make test`, and `make test-race`; the frontend suite contains 49
  tests across 23 files and meets all configured coverage thresholds.
- `make test-e2e` with five browser workflows, including terminal persistence,
  a native process lifecycle, workspace lifecycle, and a real Docker Compose
  build/start/observe/stop/teardown flow.
- `make test-visual` with 14 Chromium visual and responsive-state checks.
- `make generate-check`, `make migrate-check`, `make platform-check`,
  `make test-plugin-sdk`, production web and Go builds, and the MCP Inspector
  tools/resources/prompts smoke test.
- `govulncheck` found no reachable or imported Go vulnerabilities; the one
  advisory in a required module is not called. `pnpm audit --audit-level high`
  found no known JavaScript vulnerabilities. A local RustSec scan of all 533
  locked desktop crates found no vulnerability advisories. It reported 17
  informational maintenance/unsoundness warnings in Tauri transitive crates;
  the affected `glib::VariantStrIter` API is not used by Envarium and the
  patched `glib` series is not compatible with the latest Tauri Linux GTK3
  graph. The security workflow scans the nested desktop lockfile so upstream
  fixes cannot pass unnoticed.
- Rust formatting, strict Clippy, all 11 Rust tests, and native macOS debug app
  and DMG packaging passed. The documentation site passed generation, type and
  content checks, lint, 19 tests, production build, and validation of 141 pages.

This local evidence does not substitute for the required hosted native Linux
and Windows jobs or release-environment signing, notarization, updater signing,
attestation, and artifact publication. The local desktop bundles are unsigned
debug artifacts; reviewed release credentials remain confined to the protected
release environment.

## 2026-07-20 pull-request CI follow-up

The first hosted V1 pull-request run exposed Windows-only assumptions that
cross-compilation cannot execute: plugin protection checked Unix mode bits,
support fixtures inferred ACLs from portable file modes, golden files inherited
CRLF checkout conversion, JSON fixtures embedded unescaped Windows paths, and
SQLite file URLs did not normalize drive-letter paths. The plugin process test
also stripped the `.exe` suffix from its copied helper. These boundaries now
use platform-specific executable policy, platform-neutral fixtures, preserved
executable suffixes, forced-LF goldens, and one Windows-safe SQLite DSN
constructor. The Go and frontend formatting checks are also split between jobs
with their required toolchains.

Focused package tests, both formatting gates, `make platform-check`, and the
full `make lint` gate pass locally. Hosted Windows execution and the remaining
pull-request workflows remain the merge gate.

## 2026-10-17 generated-site CI follow-up

The previous public-site CI failure was caused by a documentation change whose
generated `llms-full.txt` update was missing. Add `make site-generate-check` to
the unified local quality gate and use the same target in public-site CI. The
target regenerates CLI/API/schema documentation and both LLM indexes, then
requires a clean diff; it preserves the existing hosted consistency check.
Contributor guidance now requires these outputs in the same commit as their
source changes. The native adapter CI matrix also executes the new socket
package directly on macOS, Linux and Windows.

The local full-site browser run also exposed Astro's agent-environment behavior:
the preview CLI detaches, so Playwright sees its managed process exit while the
server remains orphaned. The test launcher now uses the pinned Astro preview
API in the foreground with explicit graceful shutdown; it serves the real
production build and preserves every existing browser, accessibility and
visual assertion. Required verification includes the unified local quality
gate and the full hosted matrix for the final master commit.

The first master commit, `d2f469a`, passes the full local `make quality`, all
hosted core CI jobs, public-site CI and public-site deployment. The full local
site suite passes 38 browser/accessibility cases with 14 intentional skips and
four visual cases. The new foreground preview server terminates with its test
run; no listener remains on its test port.

The hosted Rust audit then exposed RUSTSEC-2026-0367, RUSTSEC-2026-0368 and
RUSTSEC-2026-0369 in the Windows updater's ZIP 4 dependency. The published
updater still requires ZIP 4. Pin the
[upstream ZIP 9 fix](https://github.com/tauri-apps/plugins-workspace/commit/ba98bfadc4c868c9a7e1a226c83af528e1f11b89)
at its exact revision, with its required stable Tauri 2.12.4 runtime and 2.7.3
build tooling. The lockfile resolves ZIP 9.0.5. This is a temporary immutable
upstream pin until a patched updater release is published. Update verification
and archive support remain enabled; no advisory ignore is added.

The resulting local Rust audit reports zero vulnerabilities, with four upstream
maintenance warnings and one existing GLib unsoundness warning. Strict Clippy
and all seventeen native tests pass. The Windows sidecar job now also runs
`cargo check --all-targets --locked` against the native manifest so the
Windows-only archive dependency is compiled in hosted CI. Native app packaging
and the release-only updater path are additionally checked locally; compile-only
updater fixture values are not signed or installed release configuration.

The first hosted Windows-native check catches a previously uncompiled startup
path: an unconditional call to the macOS-only autostart builder method. Use the
plugin's cross-platform initializer instead, retaining explicit LaunchAgent
selection on macOS and the platform implementation elsewhere. The Windows
native compile gate remains enabled as regression coverage.

## October 2026 repair follow-up

The historical evidence above is retained. Current repairs and external
acceptance gates are recorded in the [repair evidence](security-cleanup-2026-10-18.md)
and [acceptance ledger](acceptance-status.md). Prelaunch manifest migration was
retired under ADR-0025; ordered database migrations remain supported.

## 2026-10-20 first-release platform scope

This slice implements [ADR-0030](../adr/0030-first-release-platform-scope.md).
Release qualification now covers macOS and Linux; Windows/WSL2 and WinGet
remain coming soon. It addresses the platform and distribution portion of
Phase 18's exit criteria without claiming signed installer acceptance.

- The protected release workflow builds macOS arm64/Intel and Linux x86_64
  desktop bundles only. GoReleaser and nightly snapshots emit four macOS/Linux
  CLI archives. Windows adapter/sidecar CI remains visible but nonblocking.
  Local quality gates require Linux cross-compilation; Windows cross-compilation
  is available separately through `make platform-check-windows`.
- Stable artifact validation requires both Mac DMGs, Linux AppImage/deb/rpm,
  and all four CLI archives with checksums. The authenticated updater requires
  the two Mac architectures and Linux x86_64, ignoring Windows assets.
- Package draft generation produces a reviewed Homebrew cask without invoking
  WinGet generation. Distribution monitoring leaves WinGet unavailable.
- Updated the roadmap, release checklist, platform/installation docs, ADR index,
  public-site/rebrand checklists, README, and acceptance ledger to match.
- Regression coverage uses a complete macOS/Linux-only release fixture, checks
  every missing supported artifact/updater target/signature, and verifies
  Windows assets do not enter the feed or package drafts. The real CLI archive
  smoke check requires exactly four platform/architecture archives.

Verification:

- Contract generation is clean (`make generate-check`); repository and
  architecture checks, Go formatting/vet/GolangCI, Vue lint/typecheck, all Go
  tests, plugin SDK checks, empty-database migrations, and the revised Linux
  cross-compilation gate pass. `govulncheck` reports no reachable/imported
  vulnerabilities (three advisories remain in unused required modules).
- `go test -race -p 2 ./...` passes. The initial unrestricted race run hit a
  plugin helper's 15-second startup deadline; the reduced-parallelism full run
  passed without implementation or test changes.
- All 93 frontend tests across 37 files pass with coverage on the pinned Node
  26.7.0 runtime and two workers. An initial run on the system's older Node
  timed out in the unchanged Docker-unavailable UI test under concurrent load;
  that test also passes in isolation.
- All six application browser workflows and 17 visual checks pass, including
  real Docker Compose and native worktree lifecycle flows.
- Site checks, lint, all 40 unit tests with coverage, production build, and
  content validation pass (122 HTML pages). Workflow actionlint, GoReleaser
  configuration validation, and all five release/signing-script tests pass.
- An unpublished GoReleaser snapshot builds exactly four macOS/Linux CLI
  archives; release staging, aggregate checksums, normalization, and the stable
  desktop completeness smoke check pass against those actual archives.
  Distribution monitoring leaves WinGet unavailable.
- Documentation generation is identical on repeat across all eight generated
  references/indexes. The standard `site-generate-check` compares to HEAD and
  therefore reports the intentional uncommitted documentation changes; the
  remaining quality checks run with that target omitted. A clean-checkout
  `make quality` remains required after committing the generated updates.
- Desktop formatting, strict Clippy, all 21 Rust tests, the native macOS sidecar,
  and macOS debug `.app`/DMG packaging pass. Production web/Go builds and the
  MCP Inspector tools/resources/prompts smoke test pass.
- The full site browser run records 219 passes, 84 intentional skips, and one
  WebKit keyboard-scroll assertion failure on the unchanged download
  verification page. That exact case passes in isolation (2.5 seconds) without
  source/test changes. All seven site visual comparisons pass without baseline
  updates. The full mobile accessibility crawl and public-page link/layout
  crawl pass.

Signed builds, hosted CI for the changed commit, and macOS/Linux clean-machine
install/update/migration/reinstall/uninstall remain release acceptance gates.
Windows/WSL2 launch and signing are deferred; no release is published by this
repository change.

## 2026-10-20 signed release candidate preparation

The first candidate uses the existing beta channel as `v1.0.0-beta.rc.1`.
The release workflow now propagates its validated tag version to the native
app configuration, bundled sidecar, Rust compatibility check, and updater
archive names. This prevents a candidate from identifying itself as stable
`1.0.0` and preserves the subsequent upgrade to the stable release. Regression
coverage checks the prerelease archive name and rejects a different bundled
prerelease version. Signing/notarization evidence will come from the protected
release workflow on the exact CI-qualified tag.

Focused preparation checks pass: release-channel Go tests verify the exact
`GITHUB_OUTPUT` version, workflow actionlint passes, all 22 Rust tests pass
with `ENVARIUM_VERSION=1.0.0-beta.rc.1`, and site tests (40), type/content checks,
lint, and documentation regeneration pass. The first candidate will remain a
draft prerelease for installer qualification; it is not a stable publication.

Linux desktop release packaging is pinned to Ubuntu 22.04, matching the
documented baseline and avoiding an accidental glibc requirement increase from
GitHub's `ubuntu-latest` migration to Ubuntu 26. Source CI remains current.

The first tagged build (`v1.0.0-beta.rc.1`, workflow 39439880867) exposed a
release-only configuration path bug: `pnpm -C ../..` runs the CLI from the
repository root, so `tauri.release.conf.json` was not found relative to the
desktop working directory. Apple key staging and Linux GPG key import passed;
the native build stopped before signing. The protected workflow now uses the
absolute workspace configuration path and the local release script uses its
repository-relative path. The failed tag remains immutable; the repaired
commit will qualify again in CI before `v1.0.0-beta.rc.2` is built.

An unsigned local smoke build through the exact `pnpm -C ../.. exec tauri`
directory switch passes with the repaired path, and both the generated app
bundle and bundled Go binary report `1.0.0-beta.rc.2`. Workflow actionlint and
documentation regeneration pass. The RC1 CLI archives/checksums/SBOM job passed;
the workflow correctly withheld a draft because native bundles failed.

## 2026-10-20 RC2 installer qualification

The immutable `v1.0.0-beta.rc.2` tag at `a8a87a6` passed exact-commit hosted CI
and the protected signing/notarization workflow. Artifact-level installer,
browser workflow, Linux package, migration, and retained-data checks are
recorded in the [RC2 qualification report](rc2-installer-qualification-2026-10-20.md).

Publication remains on hold: `data migrate --write` incorrectly reports an
unsupported newer schema as current even though daemon startup safely refuses
it, and documented plugin refresh/trust commands fail idempotency validation.
Real project-scoped MCP start/stop, observe-profile denial, and persisted agent
audit events pass against the packaged Intel binary. Clean-machine native
desktop integration and signed updater acceptance
also remain open. Source-build and container evidence must not be read as
completion of those human release gates; Windows/WSL2 remains deferred.

## 2026-10-20 CLI repairs for RC3

This slice addresses the two RC2 qualification failures under Phase 18's
primary-workflow and preserved-data exit criteria. No accepted architecture
decision changes; the plugin OpenAPI headers now describe the existing server
requirement rather than adding a new authorization rule.

- Migration preview and write commands reject a stored schema newer than the
  binary supports with `DATABASE_SCHEMA_NEWER` and conflict exit status 4.
  Read-only `data inspect` continues to report both versions for diagnosis.
- Plugin refresh, trust, enable, disable, health, and inspection now carry an
  explicit caller-generated idempotency key through the typed HTTP client.
  The OpenAPI source and regenerated Go/TypeScript bindings require that header.
  Browser mutation-header helpers retain their inferred required-key type.
- Regression coverage exercises all human/JSON and preview/write combinations
  against newer data, checks the stable failure and unchanged database bytes,
  and forbids backup/daemon side effects. A real CLI/IPC/daemon test runs the
  sample plugin's refresh/trust/grants/health/inspection/operation/revocation
  flow and checks trust confirmation and changed-fingerprint refusal.

The RC2 report remains historical evidence. RC3 must pass exact-commit CI and
the protected signed build before artifact-level qualification. Clean-machine
native desktop and signed updater gates remain explicit and are not closed
by these code repairs.

Local preparation verification passes the full Go suite, focused CLI/API
regressions, Go vet/GolangCI (zero issues), architecture/repository checks,
frontend formatting/lint/typechecking, and all 93 frontend tests. Go and
TypeScript clients and documentation were regenerated; content validation
passes. `govulncheck` reports no reachable/imported vulnerabilities. An initial
race run under concurrent local checks hit the existing cancellation test's
five-second database-setup deadline; its retry and the hosted race gate remain
part of qualification rather than being represented as a clean initial run.
