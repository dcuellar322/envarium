# Codex Start Here — Envarium

Use this file after creating the Envarium repository and copying in the implementation package.

## Initial bootstrap prompt

```text
You are implementing Envarium, a local project-oriented development command
center. Read AGENTS.md and ENVARIUM_IMPLEMENTATION_PLAN.md completely enough
to understand the architecture, then implement only Phase 0 and Phase 1.

Preserve the modular-monolith domain boundaries. Do not create god files,
generic helper packages, hidden service locators, or speculative future
abstractions. Establish the full quality pipeline and architecture checks
before adding feature depth.

For each phase:
1. map work to the documented exit criteria;
2. implement complete vertical slices;
3. add tests as code is added;
4. run the quality gates;
5. write docs/progress/phase-XX.md with evidence;
6. do not declare completion for unverified criteria.

The interactive design reference is
`design/envarium-interactive-mockup.html`. Do not implement the complete UI
in Phase 1; only build the walking-skeleton screen required by the plan.
```

## Subsequent phase prompt

Use the standard phase prompt in Section 24 of `ENVARIUM_IMPLEMENTATION_PLAN.md`.

## Recommended Codex workflow

- Start in plan mode.
- Keep one phase per branch or draft PR.
- Use subagents for independent review or repository exploration, not to create competing architecture.
- Run a review pass using the plan's review prompt before merging.
- Commit generated schemas/clients only after confirming they are deterministic.
