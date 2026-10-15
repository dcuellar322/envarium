---
title: Public site launch contract
description: Canonical identity, domain, audience, privacy, and launch boundaries for the Envarium website.
category: contributor
audience: [maintainer, contributor]
since: 1.0.0
lastVerified: 2026-07-17
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

The current product mark is the `S` rail icon under `web/public/`. Site colors
are generated from the product's canvas, panel, border, text, blue-violet
accent, and status tokens. Initial product media is sourced from deterministic
Playwright baselines:

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
