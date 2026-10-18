---
title: "ADR-0026: Reviewed checkout transitions"
description: Keep Git branch changes and runtime restarts explicit, serialized and recoverable.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-17
---

- Status: Accepted
- Date: 2026-10-17
- Refines: ADR-0001, ADR-0008, ADR-0013, ADR-0024

## Decision

Expose detailed changed-file and local-branch inventory only on explicit request.
Keep dashboard and scheduled diagnostic Git reads small. Preserve literal NUL-
delimited filenames, independent index/working-tree status and bounded results.

Switch only existing local branches in a trusted, clean, stopped checkout. A
review binds the source branch/HEAD and target branch/HEAD; execution rereads
those facts and rejects changes, conflicts, in-progress operations and branches
used by another worktree. Never force checkout, discard edits, stash, fetch or
recurse into submodules. Disable hooks and fsmonitor; reject configured checkout
filters rather than running repository-selected programs. Preserve ignored files
with Git's no-overwrite-ignore mode.

Use durable, actor-audited operations. Agent-originated branch mutation is denied;
local users review the exact transition. Runtime mutations and checkout changes
share a root-scoped application gate, including a primary registered environment
that shares the checkout. Fresh driver observations must prove that checkout's
runtime is stopped. A runtime plan created before a checkout/configuration change
cannot subsequently execute unchanged under the same gate.

Switching leaves the runtime stopped. Offer a new lifecycle preview using the
new branch's current configuration, requiring a separate restart confirmation.
Do not execute newly checked-out commands based on approval of the old branch.
An interrupted or failed Git operation remains visible and is never silently
rolled back or retried. Existing non-primary worktrees retain isolated runtime
identities and explicit start/restart/stop previews.

Bind browser runtime confirmation and queued execution to the manifest identity
and a daemon-owned checkout revision. A root-scoped revision changes after every
attempted Envarium checkout mutation, even if the Git operation fails partway.
Revisions are opaque random identities: daemon restart or eviction of an idle
entry invalidates old reviews. Retain at most 128 root entries and never evict
an in-flight root. This adds no background Git polling. Normal source edits stay
part of the development workflow; this revision specifically binds Envarium's
managed checkout transition, while current manifest inputs are checked afresh.
