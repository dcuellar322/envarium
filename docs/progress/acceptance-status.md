---
title: Current acceptance status
description: Separate implemented roadmap evidence from external release acceptance.
category: contributor
audience: [contributor, maintainer]
lastVerified: 2026-10-20
---

This ledger is the current acceptance view. The main plan preserves its original
phase 0–19 checklists. Each phase progress file retains historical evidence;
checked implementation criteria do not certify a signed public release.

| Area | Current status | Evidence |
|---|---|---|
| Phases 0–17: local control plane, adapters and UI | Implemented, automated coverage recorded per phase | [Phase 0](phase-00.md) and phase 1–17 progress files |
| Phase 18: native adapters, stable contracts and migrations | Implemented; macOS/Linux CI required, Windows source CI nonblocking | [Phase 18](phase-18.md) |
| First release platform scope | macOS/Linux only; Windows/WSL2 and WinGet remain coming soon, excluded from release and updater gates | [ADR-0030](../adr/0030-first-release-platform-scope.md) |
| Phase 18: signed install, notarization, updater lifecycle, retained-data reinstall | Pending protected release qualification on actual signed artifacts | [Release checklist](../release.md) |
| Phase 19: optional fleet, team config, sync and telemetry | Implemented; local-only defaults and permissions remain enforced | [Phase 19](phase-19.md) |
| Phase 20: Envarium identity and operational controls | Repository implementation complete | [Phase 20](phase-20.md) |
| Browser/data/fleet security and technical debt review | Repairs implemented; verification results recorded in the repair evidence | [Repair evidence](security-cleanup-2026-10-18.md) |
| Public canonical DNS/TLS | Namecheap CNAME and TLS activation still external prerequisites; Pages fallback remains available | [Rebrand rollout](../envarium-rebrand-checklist.md) |
| Distribution accounts, OS signing and updater production keys | Account-owned prerequisites remain pending until configured and exercised | [Release checklist](../release.md) |

No hosted run or unsigned build is evidence that signed installer/update
acceptance has passed. Do not publish a stable release until that checklist is
completed for its exact commit and artifacts.
