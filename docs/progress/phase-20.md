---
title: 'Phase 20: Envarium identity and operational experience'
description: Implementation evidence for the Envarium rebrand and project-first runtime operations slice.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-15
---

## Implemented

- Renamed repository-owned product identifiers, paths, packages, generated
  contracts, manifests, UI copy, public-site metadata, fixtures, commands, and
  release configuration to Envarium.
- Added driver-neutral `deploy` and `kill` lifecycle actions to the domain, CLI,
  HTTP/OpenAPI, UI, MCP, federation, durable-operation, and audit paths.
- Added exact lifecycle previews and confirmation in the project UI, with a
  server-side confirmation requirement for destructive plans.
- Added whole-project and per-service start, stop, restart, deploy, rebuild,
  pause, resume, force-stop, and teardown controls where each runtime supports
  them.
- Added service inspection for runtime identity, container/process metadata,
  resource samples, health, ports, and observation time.
- Added structured, searchable, filterable, pausable, resource-aware project
  logs with plain-text and NDJSON downloads over existing redacted exports.
- Reviewed every documented Arcane feature and recorded delivered, equivalent,
  existing, and intentionally out-of-bound capabilities.

## Files and modules added

- `internal/runtime/process/kill.go` owns verified native process force-stop.
- `web/src/domains/projects/components/ProjectServicesPanel.vue` and
  `RuntimePlanDialog.vue` own service operations and plan review.
- `docs/adr/0018-project-runtime-operations.md` records the accepted boundary.
- `docs/reference/arcane-capability-review.md` records the complete product map.
- `docs/envarium-rebrand-checklist.md` separates completed repository work from
  external rename work.

## Architecture decisions

ADR-0018 keeps Envarium project-first and runtime-neutral. Docker Compose and
native-process adapters implement the same typed lifecycle actions. UI, CLI,
MCP, and federation remain thin adapters over the application boundary. Raw
Docker administration is not exposed, and native force-stop verifies durable
process identity before signalling an owned group.

## Tests added

- Compose plan tests cover deploy arguments/effects and selected-service kill.
- Native-process plan tests cover deploy and kill selection/risk.
- HTTP tests cover rejected unconfirmed and accepted confirmed destructive plans.
- Vue tests cover preview-before-execute, profile selection, service-scoped
  deploy/resume/kill, structured fields, and log filtering.

## Commands run and results

- `make generate-go` passed and regenerated schema, OpenAPI, and SQL contracts.
- `openapi-ts` passed and regenerated the TypeScript API client.
- `go test ./...` passed across every Go package.
- `go test -race -p 4 ./...` passed across every Go package. The bounded
  parallelism prevents an unrelated plugin-process fixture from exhausting the
  host while retaining race instrumentation.
- `go vet ./...`, pinned `golangci-lint run`, architecture checks, repository
  checks, and `git diff --check` passed.
- `govulncheck` reported no called vulnerabilities. It found three advisories in
  required modules, but confirmed that Envarium does not call the vulnerable
  symbols.
- Linux/amd64 and Windows/amd64 compilation checks passed for the platform,
  process, runtime, plugin, support, terminal, CLI, and port adapter packages.
- The SQLite empty-database migration check and plugin SDK/containment suites
  passed.
- The pinned MCP Inspector smoke passed for the observe profile's tools,
  resources, prompts, and destructive-tool exclusion.
- Vue application, Node, and test type checking, ESLint, Prettier, and the full
  Vitest suite passed: 24 files and 53 tests, with the configured coverage gate.
- Playwright passed the real Compose lifecycle flow, including previewed and
  confirmed deploy and service force-stop, plus native-process lifecycle,
  accessibility, and visual coverage.
- The production Vue bundle and renamed Go binary/desktop sidecar built
  successfully.
- Rust formatting, Clippy with warnings denied, and all 11 Tauri unit tests
  passed. An unsigned debug `Envarium.app` bundle built successfully.
- The documentation site generated, type-checked, linted, tested, built, and
  validated successfully: 146 pages, 110 documentation pages, and 56 source
  files. Chromium end-to-end, accessibility, and visual checks passed; one
  intentionally skipped browser case remained skipped.

## Acceptance criteria status

- [x] Repository-owned identity is Envarium across runtime and generated code.
- [x] Lifecycle actions are previewable, typed, service-selectable, permissioned,
      durable, auditable, and supported by both runtime drivers where applicable.
- [x] Project logs and service details provide the operational workflows selected
      from the Arcane review.
- [x] Every reviewed Arcane capability has an explicit Envarium disposition.
- [x] External rename tasks are separated into an actionable checklist.

## Known limitations

- Hosting, DNS, distribution coordinates, signing accounts, and users' local
  machine configuration cannot be changed from this checkout.
- Envarium intentionally does not become a raw Docker image, volume, network,
  registry, vulnerability-scanning, or Swarm administration console.
- Existing former-name user data requires an explicit operator migration decision;
  the new application uses Envarium paths and identifiers.

## Deferred work

Only the unchecked external actions in the rebrand checklist and capabilities
marked as architectural boundaries are deferred. A boundary item requires a new
product decision rather than silent scope expansion.

## Manual verification steps

1. Start the daemon and web UI against a trusted Compose project and a trusted
   native-process project.
2. Preview and confirm each supported whole-project and single-service action;
   verify exact commands, progress, activity, final state, and health.
3. Stream multiline and JSON logs, exercise every filter, pause/resume rendering,
   inspect fields, and download both formats.
4. Complete the external rebrand checklist in each owned hosting, distribution,
   signing, and developer environment.

## Brand asset slice — October 15, 2026

- Scope: the Phase 20 repository identity deliverable; create an additive logo
  and icon package for review, without replacing existing application assets.
- Exit checks: consistent mockup palette, reusable logo and standalone symbol,
  light/dark/monochrome variants, app/favicon/tray exports, and a coordinated
  interface icon set. Package: `design/brand/envarium-v1/`.
- The built-in image generation tool supplied the initial concept. Editable SVG
  artwork, a custom path-based wordmark, and 12 interface icons complete the set.
  The generation prompt and original reference are retained with the package.
- Validation: rendered contact sheet visually inspected; SVG XML, absence of
  external dependencies, PNG dimensions/alpha, and ICO/ICNS container structure
  checked. Source SVGs and a ZIP package are available for human review.
- No application code, database, API/schema, or architecture boundary changed.
  Runtime and application test suites are not relevant to this asset-only slice
  and were not rerun. Native installed-icon appearance and brand integration
  remain unverified; existing phase completion evidence is unchanged.

## External migration follow-up — October 15, 2026

- Renamed the local checkout to `/Users/dac/dev/envarium`, changed `origin` to
  `git@github.com:dcuellar322/envarium.git`, and verified the renamed remote at the
  existing commit.
- Added migration 18 instead of modifying the already-applied native-runs migration.
  It converts the runtime origin constraint, active project paths and identity,
  accepted/proposed manifest identifiers, application-support paths, and other
  durable runtime references to Envarium. Focused migration and CLI tests pass.
- Created a SQLite-consistent copy of the stopped former application state under
  `~/Library/Application Support/Envarium`, retained the original directory and an
  automatic pre-v18 backup, applied schema 18, and verified the project path,
  runtime constraint, manifest identity, and foreign keys.
- Updated the active Codex and VS Code project paths. No matching macOS launch
  agent or installed Envarium/Switchyard application bundle was present.
- Updated the GitHub description, homepage, topics, remote URL, and Pages project
  variable. The production workflow now creates the new Pages project in the
  original Cloudflare account, attaches the canonical custom domain, and deploys a
  path-preserving permanent redirect to the old Pages project.
- The temporary Pages project created in the wrong Cloudflare account during
  account discovery was deleted after explicit authorization.
- Upgraded the pinned Go patch release from 1.27.0 to 1.27.2 after the final
  `govulncheck` gate identified newly published, reachable standard-library
  advisories fixed by 1.27.2. Updated GolangCI-Lint from 2.13.0 to 2.14.0 so
  lint analysis supports the patched Go export-data format.
- Restored the local Rustup launcher and used the repository-pinned Rust 1.98.0
  toolchain. Removing the generated Cargo target cache cleared absolute paths to
  the former checkout. Rust formatting, Clippy with warnings denied, all 11
  desktop tests, `Envarium.app`, and `Envarium_1.0.0_x64.dmg` now pass/build.
- The complete `make quality` target passed in one final invocation with Docker
  available, including generated-file consistency, repository and architecture
  checks, Go/Vue/Rust lint and type checks, unit and race suites, plugin and
  migration checks, Linux/Windows compilation, `govulncheck`, five browser E2E
  scenarios, 14 visual scenarios, the MCP Inspector smoke, and native desktop
  packaging.
