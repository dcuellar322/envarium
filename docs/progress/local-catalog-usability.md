---
title: Local catalog usability
description: Real-project import, lifecycle, observation, desktop, and deployment verification.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-15
---

# Local catalog usability — October 15, 2026

This follow-up exercises the application against a real development directory,
with both native commands and Docker Compose projects.

## Delivered

- Imported and refreshed eleven runnable projects through the application
  services. All accepted manifests validate and all start plans preview
  successfully. Removed obsolete catalog registrations only after a verified
  SQLite-consistent backup; repository files were retained.
- Added explicit local-development commands, endpoints, readiness checks, and
  non-overlapping ports for three native applications. Repaired relocated Python
  command launchers using the existing lockfile.
- Kept declared Compose services visible and individually controllable before
  their containers exist. Disconnected rows remain unknown and do not claim
  container identity or ownership.
- Pending projects use a review entry point instead of forbidden runtime queries
  and unavailable-data claims. Missing Git/resource data is shown explicitly.
- Initialized routing after the one-time browser credential is removed so the
  router cannot restore it to the address bar.
- Shared the Docker context/TLS resolver across runtime, host, and storage
  observers. A slow disk-usage request no longer marks a responding Engine
  offline; expensive host disk inspection is cached separately from CPU/memory.
- Reused the immutable compiled manifest schema instead of recompiling it on
  every observation. Native process tracking filters group membership before
  requesting process details, reduces steady-state polling, and avoids repeated
  writes of unchanged fingerprints. Stop/kill still require fresh identity
  verification.
- Readiness retries honor their configured interval. Stopped or paused runtimes
  no longer inherit old readiness failures or claim to be healthy; failed or
  degraded runtimes cannot claim healthy status just because checks are absent.
- The public deployment retains a reachable Envarium Pages fallback while
  canonical DNS/TLS is pending, uses a temporary legacy redirect during that
  transition, and switches to a permanent canonical redirect after activation.
  Deployment summaries explicitly report pending DNS/TLS.

## Verification

- Eleven real accepted manifests and eleven start plans covering twenty-five
  declared services validated. The declared port registry reports no overlap.
- Actual npm/Astro, npm/Vite, uv/FastAPI, and Docker Compose starts completed;
  their endpoints and captured logs were checked. Native ownership survived a
  daemon restart. Completed native smoke tests were stopped through Envarium.
- All 54 web unit tests pass with coverage thresholds enforced. One initial
  run under heavy host load timed out; the complete serial rerun passed.
- Targeted race suites pass for manifest validation, Docker connection parsing,
  host observation, health, storage, Compose, and native process ownership.
- Generated contracts, repository checks, and full web type checking pass.
- Full Go tests, vet, lint, and architecture checks pass. Empty stopped-project
  collections retain their array contract instead of creating false partial-data
  errors in the dashboard.
- Seven live Pages routes passed the production smoke check, including their
  canonical metadata for the intended custom hostname.
- The manifest-validation benchmark measured about 33 microseconds per repeated
  validation on this workstation after one schema compilation.
- Five browser E2E scenarios pass, including a real Docker lifecycle and the
  authenticated settings flow. All fourteen visual scenarios pass.

Native packaging and production results are reported with the delivered
revision. Canonical DNS activation still requires zone authorization; a
Pages-only deployment token cannot create that record.
