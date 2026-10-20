---
title: Public site Phase 10 evidence
description: Evidence for accessibility, browser, visual, performance, SEO, headers, redirects, smoke automation, and external activation gates.
category: contributor
audience: [maintainer, contributor]
lastVerified: 2026-10-19
---

## Outcome

Repository hardening is complete. Chromium, Firefox, WebKit, and mobile
journeys cover download, onboarding, search, integrations, community,
navigation, and theme behavior. Axe reports no known serious or critical
violations after light-theme contrast and mobile scroll-region fixes. Four
visual baselines cover the primary surfaces.

The build validates unique metadata, canonical identity, required outputs,
image alternatives, a 300 KiB HTML ceiling, 200 KiB JavaScript asset ceiling,
and 200 KiB first-party JavaScript total; current first-party output is about
102 KiB. Production headers and redirects are versioned.

DNS, Pages hostname redirects, repository metadata, live downloads, and the
canonical production smoke remain unchecked until external activation. Follow
`docs/public-site-launch-checklist.md`; do not infer those results from the
local build.

## Public site review — 2026-10-19

This is a focused follow-up to public-site Phase 10 and implementation-plan
Phase 18's documentation/contributor exit criteria. Acceptance requires clear
product and installation copy, explicit platform availability, working
navigation and search, verified generated pages and links, accessible desktop
and mobile layouts, and screenshot evidence. It does not claim completion of
signed-release or clean-machine platform gates.

Changes:

- Replace repeated slogans, invented workflow counts, contributor placeholders,
  and implementation jargon with concrete tasks and next steps. Reduce heading
  sizes and decoration while retaining the approved brand and real product UI.
- Mark Windows and WSL2 as coming soon across home, selection, platform guides,
  requirements, roadmap, and the root installation documentation. Keep their
  download links unavailable even if release metadata includes Windows assets.
- Render download recommendations from prebuilt templates in the browser so
  channel, architecture, deep links, and browser history work on static hosting.
  Retain platform guides without JavaScript. Do not invent release URLs.
- Open documentation search from desktop and mobile marketing navigation,
  focusing its input after the search index UI loads. Cover results, no matches,
  and Escape. Handle clipboard rejection in docs.
- Add license/conduct links and describe website hosting, local search, theme
  storage, and third-party navigation in the privacy page.
- Replace the unavailable GitHub Discussions destination (HTTP 404 observed)
  with a usage-question issue form and align the support policy.
- Rewrite source-relative Markdown links to canonical HTML routes during the
  static build; preserve GitHub source links and fail for missing destinations.
- Correct duplicate top-level headings in two progress documents and improve
  status-color contrast in light mode.

Verification uses the full cross-browser journeys, an exhaustive generated-page
crawl on desktop and mobile, internal route/fragment checks, image checks,
JavaScript-error and overflow checks, axe, and reviewed screenshot baselines.
Optional full-page captures and JSON reports are written to
`output/playwright/site-audit/` by:

```bash
SITE_AUDIT_SCREENSHOTS=1 pnpm --dir site test:e2e
```

The production homepage returned HTTP 200 during review. Local changes still
require the repository's normal deployment. Published binary installation and
upgrades cannot be tested while no stable release is available.


### Verified results

- `make site-quality`: passed, including 127 brand assets, generation, Astro
  checks (no diagnostics), lint, 34 unit tests, static build, content validation,
  and build validation. First-party JavaScript remains about 103 KiB.
- Full Playwright matrix: **86 passed, 18 intentionally skipped**. Chromium,
  Firefox, WebKit, and mobile Chromium cover the applicable journeys. Axe runs
  on desktop and mobile Chromium; full crawls run on those two configurations.
- Exhaustive crawls: **162 HTML pages and 319 internal targets on each viewport**,
  with no missing fragments, broken images, horizontal overflow, or uncaught
  browser errors. **324 full-page screenshots** are saved with the two JSON
  crawl reports under `output/playwright/site-audit/`.
- Visual regression: **6 passed on macOS and 6 passed in the digest-pinned
  Playwright 1.62.1 Noble image** after reviewing and refreshing platform-specific
  references. Existing screenshot thresholds remain unchanged.
- Production smoke: **7 routes and the updater feed passed** on the canonical
  hostname. All **20 outbound marketing-site links returned HTTP 200** after
  replacing the unavailable Discussions route. New issue-form contents will
  become available when these repository changes are published.
- `make repository-check`, `make archcheck`, and `make generate-check`: passed;
  generated OpenAPI clients, SQLite accessors, and manifest JSON Schema stayed
  clean. No migrations or application runtime code changed.

A final manual keyboard check exposed delayed search-input focus on the
marketing-to-docs handoff. A targeted four-browser regression additionally
checks that input focus before typing. Other remaining release, installation,
upgrade, hosting activation, and screen-reader manual gates in the launch
checklist are not implied by these results.


## Deployment follow-up — 2026-10-19

Commit `095aa96` publishes the site review through the existing production
workflow. A pre-deployment check of the canonical site's search exposed a
hosting-only failure: the response CSP blocked Pagefind's WebAssembly module.
The local Astro preview had not served the Cloudflare `_headers` policy.

Add the WebAssembly-specific `script-src 'wasm-unsafe-eval'` permission while
keeping JavaScript string evaluation disallowed. The test preview now loads the
production script and worker CSP from `_headers` for every response. Only the
HTTPS-upgrade transport directive is omitted on the HTTP loopback preview,
because WebKit otherwise requests unavailable loopback TLS resources. Cross-browser search tests assert the header and exercise real result
navigation, preventing preview-only success. This uses the
[WebAssembly-specific CSP permission](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/script-src#unsafe_webassembly_execution).

Deployment and live verification results are reported with the completed
workflow run. Signed binary installation remains a separate release gate.

## Publication boundary — 2026-10-19

This documentation slice maps to Phase 18's public documentation and contributor
onboarding criteria. Keep maintainer evidence in Git without publishing it as
product documentation. Acceptance requires excluded HTML and raw Markdown
routes to return 404, no excluded sitemap or Pagefind entries, no excluded
content in either AI export, retained public guides and contributor references,
and a verified production deployment.

The shared publication policy excludes every current and future `progress/`
report, website launch policy/checklist, rebrand checklist, and internal
capability review. Remove internal-report links from public indexes; remaining
source-relative links to excluded documents resolve to GitHub. Static build
validation and regression tests enforce the boundary. Production verification
is recorded after the deployment completes.

Local verification: `make site-quality` passed with zero Astro diagnostics,
37 unit tests, and 120 validated HTML pages (42 internal documents excluded).
The cross-browser suite passed 91 tests with 21 intentional skips, including
120-page desktop/mobile crawls and publication/search checks. All six macOS
visual baselines passed unchanged. Repository, architecture, and generated
contract checks passed. All 38 progress sources and four excluded maintainer
records remain in Git. Live verification follows the production workflow and
is saved under ignored `output/playwright/publication/live/` artifacts.

### Hosting cache correction

The first production verification caught cached HTTP 200 responses for retired
HTML and Markdown URLs, despite their absence from the new build and search.
The account has no DNS zone for this externally hosted domain, so zone cache
purging is unavailable. ADR-0029 extends the existing small Pages Worker to
return uncached 404 responses before retired assets can be fetched. The same
publication policy generates the guarded paths and narrow invocation routes.
Regression tests simulate a stale asset store and cover aliases, descendants,
HEAD, public-route delegation, and asset failures. Final verification must
repeat the original URLs without cache-busting query strings.

The correction passes `make site-quality` (39 unit tests, 121 HTML pages after
adding ADR-0029), final lint/type/content/build checks, and repository,
architecture, and generated-contract checks. Three browser tests against the
actual local Pages runtime verify all 42 retired HTML/Markdown pairs, both AI
exports, sitemap, public search, and the full 121-page crawl. Additional runtime
checks verify 404/no-store/noindex, bodyless HEAD, public HTTP 200, and the
unchanged empty updater HTTP 204. The pinned local runtime supports compatibility
date 2026-08-27; that date was used for the local edge check. Generated Wrangler
state is excluded from lint and type checks.

## Public documentation editorial review — 2026-10-19

This documentation slice maps to Phase 18's public documentation and contributor
onboarding criteria. Review remaining published sources for internal planning
language, milestone references, account/workstation history, outdated capability
claims, and unsupported release/security assurances. Preserve versioned contracts,
accepted architecture decisions, stable URLs, and the existing 42-document
publication exclusion.

Changes replace phase-based explanations with current architecture and tasks;
correct Windows Job Object/ConPTY/IPC descriptions while retaining coming-soon
availability; replace placeholder-route claims with current routes; add native
worktree isolation and the human-only AI-assisted approval limit; clarify log
buffer ownership; surface superseding ADRs; remove personal credential-setup
history and old CPU measurements; scope compatibility promises to stable releases;
and remove stale hosting launch tasks. Release availability was checked against
GitHub (no releases). Source adapters, router, authorization, release workflow,
and data contracts are the evidence for current behavior. Content validation now
rejects implementation-phase wording in published Markdown only.

Acceptance requires generation and static validation, no phase references in
published HTML/raw Markdown/AI exports, functional search and links, retained
404s for internal reports, appropriate browser/visual checks, and a verified
production deployment. Signed installers and clean-machine platform support are
not inferred from a documentation review.

Verified locally: `make site-quality` passed (39 unit tests, no Astro diagnostics,
121 HTML pages), final type/lint/content/build checks passed, and the full browser
suite passed 91 tests with 21 intentional skips. Six macOS visual references
passed unchanged. Desktop/mobile exhaustive crawls check visible copy and linked
raw Markdown for phase wording; AI export checks cover the same editorial rule.
Search checks inspect destinations and real result navigation: Pagefind may
fuzzily match unrelated tokens, so an empty result set is not used as a content
assertion. Repository, architecture, and generated-contract checks passed.
Production verification follows deployment and is saved under ignored
`output/playwright/docs-editorial/live/` artifacts.

## Marketing and downloads editorial review — 2026-10-19

This slice supports Phase 18's documentation, contributor build, and distribution
criteria. Review the homepage, all feature pages, downloads, and public roadmap
against source behavior and GitHub release availability. Keep internal reports
excluded and Windows/WSL2 explicitly coming soon. No application architecture
decision changes.

Replace internal implementation terminology with project tasks. Correct plugin
trust and sandbox claims, scope rebuild support to Compose, and distinguish local
storage from optional external integrations. Make source builds actionable while
there are no published GitHub releases. Do not display desktop installer steps
for missing packages or CLI archive fallbacks. Document actual SBOM names and
Homebrew availability, with Windows/WinGet still coming soon. Limit the roadmap
to available source workflows and distribution priorities without speculative
hosted services, internal release gates, or dates.

Acceptance: site generation, types, lint, unit and browser checks; reviewed macOS
and Linux visual baselines; working source-build links; desktop/mobile live
crawls, search, sitemap and AI publication checks after the production workflow.
Installer signing and clean-machine support remain separate release criteria.

Local verification passed: `make site-quality` (39 unit tests, 121 built pages),
final Astro/lint checks, repository/architecture/generated-contract checks,
95 browser tests with 21 intentional skips, and six visual checks on each of
macOS and the pinned Linux CI image. The macOS homepage reference was updated
and reviewed. A screenshot review also widened the roadmap action buttons to
keep their labels on one line. Both desktop and mobile full-page/link crawls
passed again after that adjustment. Production verification receipts and screenshots
are saved under ignored `output/playwright/marketing-editorial/live/` artifacts
after deployment.

## Integration and community editorial review — 2026-10-19

This slice maps to Phase 18's integration documentation and contributor onboarding
criteria. Review integration pages, community, contributing, and changelog against
the installers, permission profiles, platform launchers, Makefile, issue forms,
and GitHub Releases. Preserve the existing publication exclusions, stable URLs,
and Windows/WSL2 coming-soon status. No architecture decision changes.

Explain project-local installation versus authorization allowlists, installer
file changes and manual removal, read-only stdio client configuration, custom
data directories, and human approval of AI-assisted revisions. Correct the VS
Code launcher requirements and replace adapter terminology with setup tasks.
Make community reports actionable and contributor checks explicit. Remove
internal release-validation wording and unsupported availability assurances.
Historical release pages must link to their own GitHub release and label the
channel selector as latest downloads, rather than implying it selects that tag.

Acceptance requires site quality, repository/architecture/generated checks,
browser and visual checks, source/configuration links, truthful empty changelog
and RSS, an isolated historical-release build, and production page/search,
publication and screenshot verification after commit, push, and deployment.

Local verification passed: `make site-quality` (39 unit tests, no Astro
diagnostics, 121 HTML pages), repository/architecture/generated-contract checks,
103 browser checks with 21 intentional skips, and six visual checks on each of
macOS and the pinned Linux CI image. Updated community references and integration
and contributor screenshots were reviewed. An isolated build of the actual
changelog pages with two synthetic prereleases verifies exact historical GitHub
links, explicitly latest-channel downloads, missing-note fallback, and two RSS
items. The production release manifest remains empty and unchanged. Fixture
evidence is under ignored `output/playwright/community-editorial/local/`;
production receipts and screenshots follow under its `live/` directory.

## About, security, privacy, and brand editorial review — 2026-10-19

This slice supports Phase 18's public documentation and security-boundary
criteria. Check the four pages and linked policies against current storage,
provider consent, periodic metrics, support-bundle, plugin, release and asset
contracts. Keep Windows/WSL2 coming soon and preserve internal evidence in Git.
No application architecture decision changes.

Replace general slogans with project tasks and useful references. Scope claims
about local storage, redaction, API permissions and provider previews; document
periodic telemetry after opt-in and the installation identifier. Correct the
unpublished stable-release support language. Keep the original brand creation
guide and prompts under source references; publish only usage guidance, artwork,
license and notices. Generate ZIPs deterministically from explicit public asset
directories, rejecting internal milestones in the public guide. Add archive
curation, cross-time-zone reproducibility, download and privacy/security journey
checks. Expand accessibility coverage to all four pages.

Acceptance requires generation and site quality, repository/architecture/contract
checks, browser and visual checks, preservation of source references, and live
page/search/archive/publication verification with screenshots after deployment.

Local evidence: `make site-quality` passes four brand/archive tests, 39 site
unit tests, type/lint checks, generation, build, and validation of 121 HTML pages.
Repository, architecture, and generated-contract checks pass. Browser checks
pass 117 cases with 27 intentional engine/viewport skips. Six visual checks pass
on each of macOS and the pinned Linux CI image; updated light/dark brand
references and manual About, Security, and Privacy screenshots were reviewed.
The original brand README is preserved verbatim in source creation notes.
Archive tests prove internal references are excluded and ZIP bytes match across
time zones. Workflow path filters include the new archive tools. Ignored evidence
is under `output/playwright/trust-editorial/`; deployment and live verification
receipts belong in its `live/` directory.

## Final public-site and CI review — 2026-10-19

This coherent documentation slice supports Phase 18's public documentation and
contributor-build criteria. Review all built public pages, policies, search,
downloads, internal and external links; preserve the unpublished maintainer
records in Git and keep Windows/WSL2 coming soon. Correct current claims rather
than changing application behavior or accepted architecture decisions.

Acceptance: review publication copy and link targets; remove remaining
unsupported-platform and telemetry wording; keep documentation action controls
out of search content and make the error-page search link open search. Run
generation, site quality, browser/visual and repository/contract checks. Commit,
push, and deploy; verify the canonical site and search with desktop/mobile
screenshots. Wait for every workflow triggered by the final commit, including
repository CI and Security, and resolve any failures before reporting success.

Local evidence: site quality passes four brand tests and 39 site unit tests,
type/lint checks, build, and validation of 121 HTML pages. Repository,
architecture, generated-contract, and staged site-generation checks pass.
Browser checks pass 121 cases with 27 intentional engine/viewport skips,
including the new missing-page search recovery and footer-index exclusion.
All six visual checks pass on macOS and on the pinned Linux CI image without
baseline changes. The docs actions and 404 recovery were also reviewed in a
real browser. External-link verification covers 285 published URLs: repository
tree validation for 179 source/edit paths, and 18 distinct HTTP targets serving
106 links with query variants and checked issue-template paths. No broken
targets were found. Ignored local logs, link results, and screenshots are under
`output/playwright/final-site/local/`; final workflow receipts and live
page/search/screenshots belong under its `live/` directory.

The first deployed full-copy comparison matched all 121 pages to the reviewed
build and caught one remaining anonymous-metrics claim in ADR-0016. Correct
that description to match periodic opt-in delivery and its installation ID,
regenerate the AI export, then repeat CI/deployment and live verification for
the final commit. This clarifies existing behavior without changing the ADR's
accepted trust model.

The follow-up `make site-quality` passes, and a visible-copy scan of all 121
rebuilt pages finds no remaining retired wording or anonymous-metrics claims.
