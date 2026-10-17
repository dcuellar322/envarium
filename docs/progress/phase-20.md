---
title: 'Phase 20: Envarium identity and operational experience'
description: Implementation evidence for the Envarium rebrand and project-first runtime operations slice.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-17
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

- Remaining hosting, DNS, distribution, signing, and other-machine changes
  require their respective account access and operational verification. Local
  follow-ups and CI fixes are recorded below rather than classified as external.
- Envarium intentionally does not become a raw Docker image, volume, network,
  registry, vulnerability-scanning, or Swarm administration console.
- The user confirmed neither product has existing users. No public former-name
  data migration or compatibility layer is required; retired private development
  state can be cleaned up. Envarium fresh-install and future upgrade/uninstall
  verification remains part of release acceptance.

## Deferred work

The rebrand checklist tracks remaining release, external-account, and
verification work. Capabilities marked as architectural boundaries require a
new product decision rather than silent scope expansion. Phase 20's completed
repository identity criteria do not imply signed-release or all-platform
migration acceptance.

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
- Production CI created and deployed the `envarium-dcuellar` Pages project,
  registered the new custom domain, and converted the legacy Pages project into a
  verified path-preserving 301 redirect. The custom hostname still needs its DNS
  CNAME because the retained Pages deployment token has no zone read/edit scope.
  The workflow now provisions that record when authorized and skips the API call
  once the hostname resolves.
- The security workflow surfaced advisories published on October 14 after the
  original quality run. Updated `sharp` from 0.35.4 to 0.35.5 and `rustls` from
  0.23.45 to 0.23.46. The JavaScript audit has no high-severity finding, RustSec
  reports no vulnerabilities, the site rebuild/validation passes, and Rust format,
  Clippy, and all 11 desktop tests pass. One moderate JavaScript advisory and seven
  allowed upstream Rust maintenance/unsoundness warnings remain non-blocking.

## Integrated branding slice — October 17, 2026

### Scope and acceptance

The user approved the October 15 brand kit for full integration. This slice
completes the Phase 20 identity deliverable across the app, desktop shell, and
public website, retaining the hierarchy and states in the canonical mockup.

- [x] Application/sidebar, narrow icon rail, marketing header/footer, docs header,
      startup screen, and social card use the approved E and custom wordmark.
- [x] App, website, and startup CSS derive their colors from the kit's
      `palette.json`. Blue/violet brand gradients remain fixed; functional light
      and high-contrast colors retain readable contrast. Terminal colors follow
      the current app theme, including changes while attached.
- [x] The shared interface sprite drives domain/navigation/action icons in the
      application and feature/brand icons on the website.
- [x] Favicons, touch/install icons, Tauri platform exports, and a macOS template
      tray icon are generated or copied from approved kit sources.
- [x] The `/brand/` page offers light/dark/monochrome logos, icons, exact palette
      values, individual downloads, and the complete updated ZIP.
- [x] Published product screenshots show the integrated app using sanitized real
      UI fixtures. macOS and Linux visual baselines were captured and reviewed.
- [x] Asset/palette drift is rejected by `brand:check`, the frontend/site CI jobs,
      and the relevant Make quality targets.

### Implementation and architecture

- `scripts/sync-brand.mjs` owns consumer CSS, sprite, copied assets, native Tauri
  icon exports, the tray RGBA template, and social PNG rendering. ICNS entries
  are sorted because the Tauri encoder otherwise varies their ordering.
- `BrandIcon.vue`, `BrandIcon.astro`, `BrandLogo.astro`, and `DocsSiteTitle.astro`
  keep presentation-specific adapters small. Core/domain services are unchanged.
- `site/scripts/sync-product-screenshots.ts` publishes reviewed fixture captures
  to stable, platform-neutral paths under `site/src/assets/product/`.
- The interactive reference, root README, brand guide, public-site inventory,
  rebrand checklist, and machine-readable site documentation reflect the new
  identity. No accepted architecture decision changed; no new ADR was needed.

### Verification evidence

Installed workspace tool entry points were used directly when the local pnpm
version manager could not fetch/verify its pinned release from the registry;
package-manager versions, dependency versions, and lockfiles were not changed.

- Vue/Node/test TypeScript checking, application ESLint, and Prettier passed.
- Astro checking and site ESLint passed with zero diagnostics.
- App Vitest: 29 files / 68 tests passed with the configured coverage gate.
- Site Vitest: 3 files / 19 tests passed with the configured coverage gate.
- App browser journeys: all 5 passed, including real native-process and Docker
  Compose lifecycle workflows and keyboard/settings coverage.
- App visual suite: all 16 passed on macOS; all 16 captured successfully with
  actual Linux Chromium from the cached official Playwright 1.62.1 container.
  Focused tests assert exact palette values, brand assets, narrow-rail behavior,
  and high-contrast white-logo visibility.
- Site browser matrix: 56 passed / 16 intentional skips across Chromium,
  Firefox, WebKit, and mobile Chromium. Branding tests cover both themes, docs
  and marketing consistency, downloads, install icons, and narrow overflow.
  Axe checks include the new brand page with no serious/critical violations.
- Site visual suite: all 6 passed on macOS and all 6 captured successfully with
  real Linux Chromium, including dark and light brand-page captures. Remote
  full-page capture used a longer timeout to accommodate transport overhead.
- Production Vue and Astro builds passed. Site content/build validation passed:
  117 documentation sources, 153 HTML pages, bounded JavaScript, valid metadata.
- `brand:check` passed for 127 consumer assets. Published SVG XML, native PNG,
  ICO/ICNS decoding, 32 px RGBA tray alpha/size, and 1200×630 social dimensions
  passed structural verification. The ZIP's entries and integrity were checked.
- Rust formatting, Clippy with warnings denied, and all 17 desktop tests passed.
  The updated Go sidecar and unsigned debug `Envarium.app` bundle built locally.
- Architecture and repository checks passed; the empty-database migration test
  passed. API clients, JSON Schema, SQL bindings, and generated API contracts
  have no diff. `git diff --check` passed.

### Limits and manual checks

This is a presentation/asset integration. The full Go/race/vulnerability suites
were not rerun because no Go code or dependencies changed; this entry does not
claim a new complete `make quality` run. Native Windows/Linux installed icon
appearance, macOS Dock/tray cache behavior, production signing/notarization, and
live website deployment were not exercised. The debug build is not a signed
release and has not replaced the user's installed app.

To smoke-check a release: install the packaged build, verify Dock/taskbar/tray
appearance in the supported OS themes, open the startup and disconnected states,
switch the app to high contrast, and check the `/brand/` downloads and social
preview after deploying the website.

## Rebrand follow-up and pre-launch cleanup — October 17, 2026

### Scope and acceptance

This follow-up maps to Phase 20's repository identity and external checklist
criteria. It verifies the integrated kit, corrects package cleanup identity,
records DNS/release prerequisites, resolves former-name compatibility policy,
and reconciles outdated completion claims. No accepted architecture boundary
changes. Signed release acceptance remains separate from repository completion.

- The approved kit is integrated; the full 127-asset check passes on macOS and
  Linux. A hosted failure after the branding commit exposed a system-font
  dependency in the social image and a native icon's one-channel rounding
  difference. Supporting social text is now outlined, with identical rendered
  pixels on both platforms. Generated native PNG checks require identical
  dimensions and pixels except for at most one channel differing by one
  intensity level. Source/copy/social checks remain exact. A regression test
  rejects larger or multiple channel differences; the brand check is retained
  in both frontend and full-site CI.
- Homebrew zap uses the actual `dev.envarium.desktop` identity for native config,
  preferences, caches, WebKit storage, and saved application state. A test reads
  the canonical Tauri config to catch future identity drift. Ordinary uninstall
  retains durable state; explicit zap is documented as data removal.
- Authoritative DNS queries identify Namecheap's `dns1.registrar-servers.com`
  and `dns2.registrar-servers.com`; the canonical subdomain returns NXDOMAIN.
  The checklist now specifies the Namecheap CNAME rather than requiring a
  Cloudflare-proxied record. DNS was inspected, not changed.
- GitHub configuration inspection found no published releases or
  `desktop-release` environment, and only Cloudflare repository secret names.
  The release guide lists the required signing credentials, updater feed/key
  setup, reviewed version/tag, package ownership, protected environments, and
  signed installation matrix. Secret values were not accessed.
- The user confirmed neither Envarium nor Switchyard has existing users and
  explicitly authorized cleanup. No new former-name importer, manifest alias,
  transition build, or deep-link compatibility is included. The active schema-18
  development database remains supported by its existing ordered migrations;
  applied migration history was not rewritten.
- After checking that no process used the retired paths, removed the Switchyard
  data directory, broken `~/.local/bin/switchyard` symlink, and three superseded
  development backups (about 26 GiB). Earlier entries describing retained
  originals/backups record their historical state; those copies are now retired.
  The live Envarium database remains schema 18, both Envarium processes remain
  running, and all 16 original project containers retain the same IDs, start
  times, restart counts, and running state.
- Updated the historical native-startup note to point to the subsequent working
  native dashboard and resource measurements. Repository completion no longer
  implies that every remaining task is external or that signed release
  installation has passed.

### Verification

The site quality target passes: Astro reports zero errors/warnings/hints,
ESLint passes, 20 unit tests pass, the production build emits 153 pages, and
content/build validation checks 117 documentation sources with bounded assets
and JavaScript. Brand drift/rounding checks pass, the social card was visually
reviewed, ZIP integrity passes, and architecture/repository checks pass.
Generated site references and LLM indexes were regenerated from source.
The four-project browser suite passes 56 scenarios with 16 intentional skips;
all six site visual tests pass without changing baselines.
Hosted CI verification is reported against the exact pushed commit separately;
it is not inferred from these local results.
No Go, Rust, database schema, or application runtime behavior changed in this
follow-up; this entry does not claim another full local `make quality` run.

### Hosted narrow-screen follow-up

The first pushed follow-up passed main CI, security, and production deployment,
but full-site CI reached a previously masked responsive failure. At a 375 px
viewport, the Linux desktop font made the hero's intrinsic grid track 391 px
wide. The same failure was reproduced with DejaVu Sans in an isolated Linux
browser; the original cached Playwright image selected a narrower fallback
font, so that environment alone had not reproduced it.

The hero now uses a zero-minimum responsive grid track and a fluid headline
that fits without splitting its long word. Mobile header spacing is reduced;
below 360 px the approved standalone mark replaces the full wordmark, keeping
Download and Menu accessible. The brand page's interface downloads use one
column at that size. The strict overflow assertion remains and now checks both
320 px and 375 px on the home, brand, and docs routes across all four browser
projects. Linux screenshots and exact page/viewport widths were reviewed.

The complete site quality target passes again. On macOS, the expanded browser
suite passes 60 scenarios with 16 intentional skips, and all six visual tests
pass without baseline changes. The reproduced Linux narrow-screen cases pass
at both widths. Hosted CI remains the final check on the pushed revision;
test assertions and committed visual tolerances were not relaxed.

The Linux visual references were captured in the official Playwright Noble
image, whose default font differs from the hosted Ubuntu runner's DejaVu Sans.
Full-site CI now uses the digest-pinned Playwright 1.62.1 image matching those
reviewed references, including its preinstalled browser binaries and fonts.
It no longer installs a second browser/dependency set into that environment.
Keep this image aligned with `site/package.json` when updating Playwright.
No visual baseline or comparison tolerance was changed to accommodate drift.

The cleanup also removed the unused `dev.switchyard.desktop` cache, WebKit
directory, and binary-cookie file after confirming no process had them open.
Current Envarium identity and browser data were retained.
