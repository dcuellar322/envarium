# Contributing to Envarium

Thank you for helping build Envarium. The project values explicit,
maintainable implementations and reviewable changes over broad speculative
abstractions.

## Before making changes

1. Read `AGENTS.md` and the relevant part of
   `ENVARIUM_IMPLEMENTATION_PLAN.md`.
2. Check `docs/adr/` for accepted constraints.
3. Read the current and preceding reports under `docs/progress/`.
4. Keep the change to one roadmap phase or one coherent vertical slice.
5. Create an ADR before changing an accepted architecture decision.

## Development expectations

- Package Go code by domain and preserve adapter → application → domain
  dependency direction.
- Keep transport handlers, CLI commands, MCP tools, and Tauri code thin.
- Add tests with behavior, including failure, cancellation, permission, and
  reconciliation paths when relevant.
- Do not add global mutable state, service locators, generic helper packages,
  unrestricted shell entry points, or speculative future-phase features.
- Use conventional commit messages such as `feat(catalog): ...`,
  `fix(runtime): ...`, `test(operations): ...`, and `docs(adr): ...`.

## Verification

Run the relevant superset of the quality commands documented in `AGENTS.md`.

After changing documentation or API contracts, run `pnpm site:generate` and
include `docs/generated/`, `site/public/llms.txt`, and `site/public/llms-full.txt`
in the same commit. `make quality` and public-site CI both run
`make site-generate-check` to detect stale generated documentation.
The repository `Makefile` and CI workflows are the canonical check entry points.
Generated contracts and schemas must be reproducible and
clean after generation.

Every completed phase updates `docs/progress/phase-XX.md` with commands,
results, acceptance-criteria status, limitations, and manual verification.

New contributors can follow [getting started](docs/getting-started.md). Adapter
changes must also follow [adapter development](docs/adapter-development.md) and
run `make platform-check`. Public contract changes update
[compatibility](docs/compatibility.md), generated clients/schemas, and migration
guidance in the same pull request.

## Pull requests

Explain the user-visible and architectural outcome, link the phase or issue,
and complete the repository pull-request checklist. Prefer small commits with a
single intent. Do not mix unrelated formatting or refactoring into feature
work.

By contributing, you agree that your contributions are licensed under the
Apache License 2.0.
