---
title: "ADR-0025: Origin proof and immutable reviewed inputs"
description: Close cross-port browser replay and bind external disclosure to reviewed bytes.
category: concept
audience: [contributor, integrator]
lastVerified: 2026-10-18
---

- Status: Accepted
- Date: 2026-10-18
- Refines: ADR-0013, ADR-0014 and ADR-0016

## Context

Cookies ignore TCP ports. A development server sharing the browser API's
loopback hostname can receive its cookie. Cookie possession must therefore
never independently authorize API reads or interactive WebSocket attachment.
Live repository files and imported configuration can also change after review.

## Decision

Require the session's origin-scoped proof on every browser API request,
including reads and side-effect-free plans. Do not issue bearer session cookies.
Keep the bootstrap-issued proof in origin-scoped session
storage, send it as X-CSRF-Token for HTTP, and offer it in an authentication
WebSocket subprotocol before attachment. Negotiate only the public
`envarium.v1` protocol so browser handshakes complete. Never negotiate/echo the secret or
place it in URLs. Preserve same-origin checks, idle/absolute expiry, IPC bootstrap and
terminal ownership. Invalid proof must not extend session lifetime.

External evidence dispatch and trust-changing imports require the digest of
the reviewed canonical input. Revalidate before dispatch/import and reject
changed input rather than silently using new bytes. New consent is required
after provider, limits or payload changes.

Repository readers use rooted filesystem operations, regular-file validation,
nonblocking opens on Unix and hard byte caps. Evidence reads reject symlinked
sources rather than allowing aliases to bypass sensitive-path policy. Current
valid manifest edits remain observable; immutable caches do not hide changes.

Remote capabilities derive from a resolved target, with an environment bound
to its actual parent project. Current signed policy restricts reads as well as
mutations, including inventory fetched to validate an operation target.
Workspace operation authorization uses immutable submitted members, and
execution requires the reviewed workspace revision. Unsupported policy promises must not be presented as enforced.

The replay journal retains the latest 100,000 sequence positions, pruning on
startup's first publish and every 1,024 events. Clients older than the retained
window must refresh authoritative state. Audit records and resource history
have separate ownership and are not pruned by this policy. Freed SQLite pages
are reused; no blocking full-database VACUUM runs in monitoring paths.

Accept only the current stable manifest contract, without an obsolete-format
normalizer or an old-daemon status fallback. Ordered database migrations remain
supported; accepting an old manifest format is a separate compatibility decision.

Signed configuration is mandatory for all team bundles, so there is no optional
`requireSignedConfiguration` switch. The registry publisher allowlist is named
`allowedRegistryPublishers` and restricts curated registry metadata. Explicit
local trust of a native plugin remains a separate owner-level decision; this
policy does not claim to sandbox native code.

## Consequences

Cookies carry no browser authority. Browser
clients must send proof on reads and WebSockets; privileged IPC is unchanged.
Reviewed-input digests are additive contract changes enforced fail-closed.
Adversarial special files, changed input, expired sessions and rejected scopes
receive regression coverage alongside ordinary workflows.
