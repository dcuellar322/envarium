---
title: Public site launch contract
description: Canonical identity, domain, audience, privacy, and launch boundaries for the Envarium website.
category: contributor
audience: [maintainer, contributor]
since: 1.0.0
lastVerified: 2026-10-17
---

Envarium's public identity is **Envarium — Local Development Command
Center**. The primary promise is: **Your local development command center.**

The product is an open-source, local-first command center for managing local
repositories and development environments as projects rather than as
disconnected containers, scripts, and terminal tabs. Public copy pairs the
name with “Local Development Command Center” to distinguish it from unrelated
products named Envarium.

## Audience and outcomes

The primary audience is developers who maintain several local projects and use
a mix of Docker Compose, native processes, Git, terminals, and coding agents.
The public journeys are:

1. Understand the product and its trust model.
2. Choose a supported installation and complete a first project.
3. Learn, troubleshoot, integrate an agent, or contribute.

Primary conversions are a verified download, the five-minute quickstart, the
GitHub repository, a correctly routed Discussion or issue, and a focused pull
request.

## Canonical and hosting policy

- Canonical origin: `https://envarium.davidcuellar.tech`
- Documentation: `/docs/`
- Downloads: `/download/`, linked directly to GitHub Release assets
- Hosting: Cloudflare Pages static output
- Preview deployments: non-indexable
- Generated Pages production hostname: permanent path-preserving redirect to
  the canonical origin
- Optional `davidcuellar.tech/envarium` route: redirect only, never duplicate
  content

The repository controls build output, redirect requirements, headers, and
smoke checks. Cloudflare project creation, DNS ownership, the custom hostname,
and GitHub repository metadata are external launch checklist items.

## Naming and collision review

Search results contain unrelated historic and current software named
Envarium. Metadata, social cards, release stories, and major documentation
entry pages therefore use the full descriptive identity. The repository and
binary names remain `envarium`; a later legal review may supersede this
decision without changing public URL paths.

## Privacy

The site ships without behavioral analytics, advertising identifiers,
fingerprinting, or cookies. Ordinary Cloudflare delivery and GitHub navigation
are described on the privacy page. A future analytics provider requires an
explicit purpose, bounded events, retention period, and opt-out policy before
code is added.

## Brand and screenshot inventory

The approved identity is the rounded E and path-based Envarium wordmark in
`design/brand/envarium-v1/`. Its `palette.json` owns the shared canvas, panel,
border, text, blue-violet accent, and status colors. `pnpm brand:sync` generates
consumer palette CSS, copies artwork, creates the interface icon sprite,
regenerates native icons, and renders the social PNG. `pnpm brand:check` rejects
drift. The `/brand/` page publishes the complete kit and individual downloads.
Light-mode links and high-contrast UI use accessible functional shades while
the brand gradient remains blue to violet.

Published product media under `site/src/assets/product/` is copied from reviewed
deterministic Playwright baselines with `pnpm --dir site screenshots`. The
publication names are independent of the capture platform:

- dashboard
- project detail and live logs
- port registry
- workspace progress
- agent/terminal session
- manifest review
- diagnostic automation review

The fixtures use generic project names and `/Users/dev/projects/...` paths.
Before every public capture, reviewers verify that no personal path, private
project name, secret, token, or live diagnostic data appears.

## Browser verification environment

Full-site CI runs in the digest-pinned Playwright 1.62.1 Noble image declared in
`.github/workflows/site-ci.yml`. Its browsers and fonts match the reviewed
Linux visual references. Use that environment for Linux baseline review;
macOS has separate references. A host runner's system font can change text
wrapping and page height even when the browser version matches.

When updating `@playwright/test`, update the workflow image together with it,
run the full browser and visual suites, and review any changed references.
Keep overflow assertions and visual thresholds intact. The branding suite
checks home, brand, and docs routes at both 320 px and 375 px; verify wider
system fonts as well when changing the header or headline layout.
