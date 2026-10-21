---
title: "ADR-0032: Version-bound updater signatures"
description: Authenticate the release version as part of every updater signature.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-21
---

- Status: Accepted
- Date: 2026-10-21
- Extends: ADR-0028 and ADR-0031.

## Context

Native qualification accepts a genuine signed RC3 payload when a fixture feed
advertises it as RC4. TLS protects feed transport, but the payload signature
alone does not bind the unsigned catalog's version. An attacker controlling the
feed could replay an older signed release under an inflated version. The pinned
updater already supports signed versions, but its compatibility default allows
historical signatures without them. The old CLI emits no signed version.

## Decision

Enable `requireSignedVersion` in production. Pin official Tauri CLI 2.12.3,
whose build signer writes the app version into Minisign's authenticated trusted
comment. Use the existing key and canonical HTTPS endpoint. Require one matching
signed version during artifact collection, final release validation, and stable
catalog generation, before any artifact or catalog is published. Verify complete
payload bytes at packaging and native installation as before.

All five Mac/Linux updater payloads must bind the exact candidate or stable
version. Unsigned version fields, unversioned historical signatures, duplicate
version fields, and mismatches fail closed. Existing immutable RC tags/assets
are retained as evidence and remain unpublished; a new candidate contains the
native policy and new signer. No installed public release requires legacy
signature compatibility because the first release has not been published.

## Consequences

Tamper qualification includes replaying actual older signed payloads under a
newer advertised version, in addition to byte/signature/key tampering. No key
rotation, legacy bypass, or weaker transport is introduced. Updating from an
older candidate to a newly signed release remains possible; enforcing the new
policy requires installing the new native app.

The exact CLI security upgrade and platform binaries receive a narrowly scoped
package-age exception. The general dependency age policy stays enabled.
