---
title: Security model and trust boundaries
description: Local assets, authorization boundaries, mitigations, residual risks, and release verification.
category: concept
audience: [user, integrator, contributor, maintainer]
since: 1.0.0
lastVerified: 2026-10-19
---

This document describes Envarium's security design and its limits. It is not
a security certification or a statement that a release has no vulnerabilities.
The secure default is a single user operating approved local repositories;
repository content, browser pages, processes, containers, plugins, agents, and
AI output are not trusted merely because they are local.

## Assets and boundaries

- Durable project state, manifests, operation/audit history, log segments, and
  desktop preferences.
- Credentials referenced through Keychain, Secret Service, or Credential
  Manager; secret values must remain memory-only and redaction-aware.
- Authority to start/stop process trees and Compose projects, launch reviewed
  actions, allocate ports, and mutate workspaces.
- Unix sockets or Windows named pipes for privileged clients; loopback HTTP for
  an authenticated same-origin browser session.
- External executables: Docker/Compose, Git, shells, editors, coding agents,
  AI-provider CLIs, and explicitly trusted out-of-process plugins.
- Release artifacts, updater metadata, signing keys, and generated contracts.

## Threats and controls

| Threat | Primary controls | Residual risk |
|---|---|---|
| Malicious repository executes during discovery | bounded allowlisted reads; no discovery commands; explicit trust before lifecycle/action/agent operations | an approved repository may intentionally define powerful reviewed actions |
| Cross-origin browser mutation | loopback binding, bootstrap exchange, origin-scoped session proof on every HTTP/WS request, Origin checks for HTTP/WebSocket, no permissive CORS | malware running as the same OS user remains outside the browser threat boundary |
| Another local user invokes privileged API | mode `0600` Unix socket; Windows named-pipe DACL grants only LocalSystem and current-user SID | administrators can inspect or control a user's processes by OS design |
| PID reuse or orphaned child escapes lifecycle | executable/start/cwd fingerprint, process groups on Unix, Job Objects with kill-on-close on Windows, bounded graceful-to-force stop | externally started processes remain observed rather than silently adopted |
| Terminal becomes generic browser shell | server-resolved trusted launch plans, owner-scoped session attachment, bounded input/output, memory-only scrollback, audit metadata | terminal output can contain application data and is visible to its authorized owner |
| Secret reaches logs/provider/support bundle | credential references, supported redaction paths, bounded provider preview, repository source excluded from support bundles | terminal output and unfamiliar secret formats may contain private data; review before sharing |
| Plugin or agent exceeds permission | exact protocol, executable fingerprint trust, capability/scope intersection, per-project profiles, typed tools, audit, no generic MCP shell | trusted executables retain OS authority; API permissions do not sandbox them |
| Prompt/log injection changes authority | evidence treated as inert untrusted text, schema validation, evidence citations, deterministic-first results, human review | a provider can return poor advice but cannot authorize an action |
| Update or artifact tampering | protected release environment, platform signing, updater signatures, SHA-256 checksums, keyless Sigstore bundle, SBOM, GitHub artifact attestation | signing-account compromise requires key rotation and release revocation |
| Migration corrupts or strands data | single-daemon lock, compatibility preflight, quick-check, non-overwriting consistent backup, post-backup verification, newer-schema refusal | disk loss affecting both source and same-disk backup requires external backups |

## Limits and release verification

Trusted plugins and actions execute with the local user's authority. Envarium
cannot protect against a compromised account with equivalent OS access, and
redaction cannot recognize every novel secret format. Optional providers receive
only reviewed evidence, but their own retention terms apply after dispatch.

The source includes Windows named-pipe permissions, Job Objects, and ConPTY;
**Windows and WSL2 support is coming soon**. Source implementation does not
establish installer support or clean-machine validation. See
[platform support](../platform-support.md).

CI includes CodeQL, dependency review, secret scanning, vulnerability checks,
lint, architecture checks, race tests, contract generation, and adapter tests.
Release review must check current results, dependency advisories, signatures,
attestations, and platform installation behavior for the exact release commit.
A passing check is evidence for its tested scope, not a blanket assurance about
all dependencies or deployments. See [release engineering](../release.md).

Report suspected vulnerabilities through the repository's
[private security reporting route](https://github.com/dcuellar322/envarium/security/advisories/new).
Do not put secrets or exploit details in a public issue.
