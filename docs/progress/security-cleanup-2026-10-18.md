---
title: Security and technical debt repairs, October 2026
description: Repair scope, regression evidence and remaining release qualification.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-18
---

## Scope and strategy

The review at `450b86e` identified seven reportable security findings and a
technical debt backlog. These repairs enforce the shared application or
transport boundary and preserve legitimate typed operations. ADR-0025 records
browser proof, reviewed-input consent, replay retention and prelaunch cleanup.

| Finding | Repair | Focused evidence |
|---|---|---|
| Cross-port browser cookie replay | Origin-scoped proof on every HTTP and WS request; retire bearer cookies and strip secret subprotocol before upgrade | Session expiry/tab tests; cookie-only GET/HEAD/OPTIONS/plan and all four WS path classes denied; valid proof allowed |
| Unbounded live manifest/evidence reads | Rooted file handles, no symlink sources, regular-file/type/identity checks and hard byte cap; nonblocking Unix opens | Regular/oversized/missing/symlink/FIFO regression tests plus manifest/discovery suites |
| Git remote credentials | Strip URL userinfo, query and fragment; redact malformed or ambiguous credential forms | Username PAT, password, encoded userinfo, SSH/SCP and ordinary remote controls |
| Plugin structured diagnostics | Canonical redactor at application return/persistence boundary; sanitize nested output/logs/errors, reject secret-bearing identifiers | Known-secret canaries in health/inspect/output/logs/errors and alternate fact/action/risk identifier fields |
| Changed AI evidence after consent | Digest binds provider, normalized limits and exact sanitized payload; execution uses the same proposal snapshot | Missing/changed candidate/excerpt/limits consent rejects before provider; exact preview/sent/receipt bytes match; cancellation and succeeded-run recovery retained |
| Fleet target-capability confusion | Resolve registered projects/environments and parent ownership before submission | Forged ordinary-project environment ID, wrong parent, environment-as-project denied; both legitimate capabilities work |
| Fleet read-policy omission | Apply current policy on identity/snapshot/register/probe and target-validation inventory reads | Denied-policy tests cover agent and controller paths, including no peer call through operation validation |

One fresh read-only investigator preceded implementation. One fresh read-only
candidate review identified four remaining routes: plugin identifier fields,
mutable workspace membership, path-filtered site CI versus exact-commit release
gating, and operation target inventory reads. Each was confirmed, repaired and
covered by a focused regression. No second review cycle was used.

## Correctness and cleanup

- Synchronize text redaction against secret registration and deduplicate secrets.
- Preserve unchanged enabled plugins during discovery; missing/changed packages
  lose execution grants and reappearance requires review.
- Carry explicit provider disablement through availability and dispatch. Use
  absolute PATH entries and platform-native Windows EXE/COM validation; never
  infer a batch-shell launch from PATHEXT.
- Capture workspace operation participants durably. MCP get/wait/cancel checks
  those participants even after membership changes; submission/execution and
  opt-in recipes reject changed reviewed workspace revisions.
- Preview exact team publisher keys and signed configuration, and require a
  digest covering the incoming document plus current replacement context.
- Remove the unused HTTP client constructor and prelaunch manifest migrator /
  old-daemon status fan-out. Preserve ordered SQL migrations and active data.
- Bound replay history to 100,000 sequence positions with periodic indexed
  pruning. Old cursors explicitly refresh; audit and metrics history are separate.
- Split manifest validation, operation execution and typed HTTP clients into
  cohesive files. No new dumping-ground abstractions were introduced.
- Align explicit GoReleaser archive names and alpha/beta/nightly grammar with
  the site contract. CI packages real unsigned CLI archives and validates their
  aggregate checksums; stable still requires desktop packages.
- Require successful exact-commit product, full-site and security workflows
  before release packaging. Site CI runs on every master/main push.
- Replace Cloudflare-zone DNS mutation with registrar readiness reporting.
  Namecheap owns the canonical CNAME; Pages fallback stays available.
- Apply the PostCSS selector parser security fix via a narrow dependency override.

## Verification

Focused security, compatibility and independent-review regressions pass. The
original cookie-only, file-type, secret-canary, stale-consent and wrong-target
triggers now reject; legitimate browser/terminal, provider, plugin, fleet and
workspace controls remain covered.

| Verification gate | Commands and result |
|---|---|
| Syntax/import/build boundaries | gofmt, `go vet ./...`, GolangCI-Lint (zero issues), architecture/repository checks: pass |
| Focused exploit alternatives and legitimate controls | New origin-proof/real-WS, bounded-file/FIFO, Git, plugin identifiers, AI consent, fleet and immutable workspace scope tests: pass |
| Owning application checks | `go test ./...` and `go test -race ./...`: pass; migration check: pass |
| Frontend compatibility | Typecheck/lint and 68 unit tests: pass; all 5 browser integration and 16 visual tests: pass |
| Site and packaging | Site quality/validation: pass; actual six-target GoReleaser snapshot, aggregate checksums, alpha normalization and stable desktop requirement: pass |
| Dependencies | govulncheck: no called vulnerabilities; fresh pnpm audit: zero advisories |

The required hosted checks on this commit provide native Linux/Windows/macOS,
full-site browser/visual and desktop build evidence. Release-only signed gates
remain distinct below. No test tolerance, permission check or monitoring
interval was weakened to obtain these results.

The first sandboxed full Go run could not bind test sockets; its unrestricted
rerun passed. Browser integration exposed missing public WS negotiation; the
corrected handshake and complete integration suite passed. A source build of
GoReleaser hit a tool dependency conflict; the checksum-verified official
2.16.0 binary built all six archives successfully.

## Installed-workload resource follow-up

The repaired control plane was installed with 11 registered projects and 16
running user containers. Graceful daemon replacement preserved their container
identities, start times and restart counts. Its first closed-app measurement
averaged 1.27% CPU over three minutes, above the existing under-1% target.

Docker's default two idle connections per host caused observation connection
churn. The observer now retains up to 32 idle connections, and finite metrics
responses consume a bounded trailing chunk before closure. A real HTTP fixture
demonstrates three unchanged metric samples using one connection instead of
four; concurrent eight-request waves also reuse their connections. Verified TLS,
redirect handling and context-specific Unix/named-pipe dialing are preserved.
Full Go tests, race tests, vet and lint pass. Polling intervals, sample contents
and repository-file safety checks are unchanged. A missing-source traversal
experiment did not show a useful improvement and was removed. Final installed
CPU/memory measurements and workflow links accompany the retained fix report.

Normal log appends also avoid a redundant read transaction: the existing unique
digest insert remains the atomic replay guard. A preflight lookup is retained
before segment allocation/rotation, and lookup failures now fail closed.
Regression tests cover replay without extra files or publications and canceled
appends without changes to durable entries, for both full and reusable segments.

## Remaining external acceptance

Signed OS installation/notarization, production updater signing and retained-data
reinstall require protected release credentials and real signed artifacts.
Unsigned snapshots do not satisfy those gates. Canonical DNS/TLS still requires
the Namecheap account change described in the rebrand checklist.

The latest published Tauri-2 updater 2.14.0 still requires vulnerable ZIP 4.
Retain the immutable upstream security fix rather than replacing it with that
published crate or a Tauri-3 alpha. Revisit when a compatible patched Tauri-2
release is published.

The existing GLib advisory GHSA-wrw7-89jp-8q8g remains tracked: the affected
VariantStrIter API is not used by Envarium, and its patched GLib series is
incompatible with Tauri's GTK3 stack. This is not a claim of zero dependency
advisories. The hosted audit and documented exception must remain visible.
