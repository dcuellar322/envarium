---
title: Envarium rebrand checklist
description: Repository-complete and external follow-up work for adopting the Envarium name.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

The repository identity and approved visual integration are complete. This
checklist separates implementation from remaining release, account, and
verification work; checked historical runs are not claims that every supported
platform or currently installed bundle has been revalidated.

## Completed in this repository

- [x] Rename the Go module, imports, root command, binary path, CLI examples,
      completion text, generated schema tools, and build scripts.
- [x] Rename npm workspace/package scopes, Rust crate and binary metadata,
      Tauri product name, sidecar, bundle identifier, window title, and deep link.
- [x] Rename application domains, API schema identifiers, generated Go and
      TypeScript contracts, MCP tool prefixes, plugin protocol examples, telemetry
      identifiers, user-agent strings, and remote-agent vocabulary.
- [x] Rename manifest directories, schema IDs, accepted-manifest fixtures,
      examples, test fixtures, local data/database names, IPC names, environment
      variables, cache paths, support artifacts, and log download filenames.
- [x] Rename the implementation plan, design reference, agent guidance,
      documentation, public-site copy, metadata, package drafts, release artifact
      names, issue templates, and CI/release workflows.
- [x] Integrate the approved Envarium E logo, custom wordmark, palette, and
      interface icons across the application, website, docs, startup screen,
      favicons, social card, and native Tauri icon exports.
- [x] Regenerate OpenAPI clients, manifest schema, CLI/MCP documentation, and
      SQL bindings from the renamed sources.
- [x] Align generated Homebrew cleanup paths with the actual desktop bundle ID
      `dev.envarium.desktop`, retaining data on ordinary uninstall.
- [x] Resolve former-name compatibility policy: this is pre-launch software
      with no Switchyard users. No public legacy data importer, schema aliases,
      transition build, or former deep-link compatibility is required.

## Source-control hosting

- [x] Rename the local checkout directory to `/Users/dac/dev/envarium`.
- [x] Rename the hosted repository to `dcuellar322/envarium` and update this
      checkout's `origin` to the renamed SSH URL.
- [x] Update this repository's description, homepage, discovery topics, and
      Cloudflare Pages project variable on GitHub.
- [ ] Review other developers' remotes, forks, submodules, pinned links, branch
      protections, rulesets, deploy keys, webhooks, GitHub App installations,
      and social preview.
- [ ] Update organization/team ownership rules, CODEOWNERS paths if hosted rules
      refer to the old repository, project boards, issue forms that live outside the
      repository, and external status badges.
- [x] Audit this workstation's Codex, VS Code, shell, launch-agent, and common MCP
      client configuration locations; update the active Codex and VS Code project
      paths to `/Users/dac/dev/envarium`.
- [ ] Search other repositories, internal docs, bookmarks, other machines, and
      team bootstrap scripts for the former repository and binary names.

## Domains, website, and discovery

- [ ] Register or select the production Envarium domain and confirm trademark
      and naming clearance before launch.
- [x] Create the `envarium-dcuellar` Cloudflare Pages project beside the existing
      `switchyard-dcuellar` project with `master` as its production branch, deploy
      the generated site, and register `envarium.davidcuellar.tech` as its custom
      domain.
- [ ] Update DNS, registrar settings, Cloudflare Pages project/custom domains,
      redirects, certificates, canonical origin, CSP allowlists, analytics/search
      console properties, sitemap submission, and uptime monitors. The remaining
      immediate blocker is a Namecheap Advanced DNS CNAME with Host `envarium`,
      Value `envarium-dcuellar.pages.dev`, and automatic TTL. An October 17
      authoritative query confirms `dns1.registrar-servers.com` and
      `dns2.registrar-servers.com` serve the zone and the subdomain is absent.
      Keep the existing nameservers; a Cloudflare-proxied record is not required
      for this externally hosted subdomain. Once the record resolves, confirm
      the registered Pages custom domain becomes active and HTTPS succeeds.
- [x] Redirect every former public URL path-for-path where control is retained.
      While canonical DNS/TLS activation is pending, the legacy hostname uses a
      temporary redirect to `envarium-dcuellar.pages.dev/<path>`. Deployment switches
      it to a permanent redirect to `envarium.davidcuellar.tech/<path>` after that
      hostname responds successfully over HTTPS.
- [ ] Update social accounts, community links, package-directory listings,
      screenshots, videos, launch posts, press assets, and SEO title/description data.

## Distribution and signing

- [ ] Rename GitHub Releases and any container registry, Homebrew tap/formula,
      Scoop bucket, WinGet package, npm organization, or package mirror used for
      distribution; reserve the new coordinates before publishing.
- [ ] Update code-signing/notarization profiles, macOS bundle registration,
      Windows publisher metadata, updater endpoints/public keys, release channels,
      SBOM/provenance subjects, and checksum consumers.
- [x] Treat the new desktop bundle identifier and deep-link scheme as a distinct
      installed application. Use only Envarium identity and fresh-install
      behavior; retire private former-name development state after verification.
      No automatic legacy discovery or link takeover is implemented.
- [ ] Configure the `desktop-release` environment and credentials in the
      [release setup checklist](release.md#release-account-setup). October 17
      inspection found only `production` and `site-preview` environments and
      only the two Cloudflare repository secrets; signing/updater setup is absent.
- [ ] Regenerate all release artifacts and verify their filenames, embedded
      metadata, signatures, quarantine/notarization state, and upgrade behavior on
      every supported platform.

## CI, infrastructure, and secrets

- [x] Retain the existing scoped Cloudflare account/token secrets and update the
      repository's Pages project variable from `switchyard-dcuellar` to
      `envarium-dcuellar`.
- [ ] Review any remaining repository/environment secret names, OIDC subject
      conditions, artifact buckets, cache keys, service accounts, alert routes,
      and dashboards configured outside GitHub.
- [ ] Re-authorize integrations whose permissions are tied to the former
      repository or bundle identifier.
- [ ] Update allowlists for the new CLI executable, local IPC endpoint, deep link,
      browser origin, remote-agent user agent, and release download hosts.

## Existing users and local state

- [x] Transfer this workstation's development state. With the former daemon stopped, create a
      SQLite-consistent copy at `~/Library/Application Support/Envarium/envarium.db`,
      initially preserve the former directory as rollback data, apply schema 18, migrate active
      project/manifest/runtime identifiers, copy durable logs/plugin/runtime state,
      and verify foreign keys and migration status.
      This was a private development transfer. The user subsequently approved
      retiring original Switchyard data and legacy artifacts because neither
      product has existing users; no public legacy upgrade path is required.
- [x] Update active Codex and VS Code project registrations to the renamed checkout;
      confirm no matching macOS launch agent is installed on this workstation.
- [ ] Update MCP registrations and any remaining editor tasks, shell aliases,
      systemd units, Windows services, desktop shortcuts, firewall prompts, saved
      terminal profiles, and automation scripts on other clients or machines.
- [x] Update the local Codex trusted-project path to `/Users/dac/dev/envarium`.
- [x] Remove retired former-name development data, CLI, sockets, and rename
      backups after confirming they are not used by the active Envarium daemon.
      October 17 cleanup removed the Switchyard data directory, broken CLI
      symlink, and three superseded development backups (about 26 GiB); the
      active Envarium database and running project workload were retained.

## Brand and legal review

- [x] Approve and implement the Envarium logo/icon system; regenerate Tauri, favicon, social-card,
      documentation, and installer sizes from one source asset.
- [ ] Review notices, copyright ownership, privacy/security contact addresses,
      support channels, license headers, contributor agreements, and trademark usage.
- [ ] Run a final case-insensitive search of source, generated output, release
      archives, screenshots, and external systems for the former name.

## Final verification

- [x] Run `make quality` from the final rebrand worktree with Docker available;
      generation checks, repository/architecture checks, lint, type checking,
      unit and race tests, real Compose browser flows, visual tests, vulnerability
      scanning, MCP inspection, cross-platform Go compilation, and the macOS
      desktop build all pass.
- [ ] Build signed CLI and desktop artifacts on macOS, Linux, and Windows.
- [ ] Install fresh, verify future Envarium upgrade/data-retention behavior,
      exercise Envarium deep links, connect MCP clients, and uninstall cleanly.
      A public upgrade from Switchyard is intentionally outside release scope.
- [x] Import this workstation's eleven runnable development projects, validate
      their manifests/start plans, smoke-test native and Compose lifecycle/log
      controls. The October 15 native cold-start failure recorded in
      [local catalog usability](progress/local-catalog-usability.md) was resolved;
      later [native verification](progress/phase-12.md) includes visible, hidden,
      and closed-app measurements with the running workload. The subsequently
      branded development build still needs installed/release appearance checks.
- [x] Deploy the renamed public site, pass its seven-route production smoke
      check, and verify a temporary legacy redirect to the reachable Pages origin.
- [ ] Crawl the production site and documentation for broken links, mixed branding,
      incorrect canonical URLs, and old download artifacts.
