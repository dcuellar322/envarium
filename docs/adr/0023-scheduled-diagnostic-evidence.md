---
title: "ADR-0023: Collect scheduled diagnostic evidence for notification rules"
description: Preserve automatic alerts without repeatedly gathering unrelated Git and cleanup recommendations.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

- Status: Accepted
- Date: 2026-10-17
- Refines: Phase 17 diagnostic collection

## Context

A longer resource measurement exposed a five-minute CPU spike. The automation
scheduler builds full diagnoses for every trusted project, including Git status
and cleanup recommendations, even when zero automation recipes exist. Git child
commands consumed two CPU seconds in one pass. Simply skipping projects without
recipes would also remove automatic diagnostic notifications, which is not an
acceptable monitoring tradeoff.

## Decision

Keep the five-minute pass for every trusted project, with current runtime,
health, logs, ports, resource warnings, operations, configuration and approved
actions. These inputs cover every notification-producing rule and supported
recipe trigger. Omit Git status and storage cleanup enrichment on scheduled
passes, which feed only non-notifying recommendations. Filter scheduled diagnoses
to notification-producing deterministic hypotheses, so omitted enrichment cannot
create false stale-project or incomplete Git recommendations. Label the omitted
scope explicitly in the saved bundle warnings.

Manual diagnoses, optional AI evidence, and explicit recipe evaluation retain
full collection. Due enabled recipes still use current action permission checks,
cooldowns and daily limits. No recipe or alert is disabled, and no polling,
health, resource or diagnostic cadence changes. Collectors without the optional
scheduled-evidence boundary retain the existing full collection fallback.
Verify parity of all notification codes, summaries and suggested actions between
full and scheduled evidence, plus recipe dispatch and cancellation/error paths.

## Consequences

Routine automatic monitoring no longer runs Git commands or computes unrelated
cleanup recommendations. Full detail remains available on demand. New notifying
rules or recipe triggers must include their dependencies in scheduled evidence
and extend the parity test before acceptance.
