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
