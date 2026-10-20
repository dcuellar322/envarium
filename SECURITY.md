# Security policy

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Until a private
security contact is published, use GitHub's private vulnerability reporting for
this repository. Include affected versions, impact, reproduction steps, and any
suggested mitigation. Maintainers will acknowledge a complete report as soon as
practical and coordinate disclosure after a fix is available.

## Supported versions

Before the first stable release, report source-build issues with the exact
commit and platform. Windows and WSL2 support is coming
soon. Once stable v1 releases are available, security fixes target the latest
stable v1 minor release. The main branch receives fixes first; prerelease builds
do not carry stable support or compatibility promises.

## Security baseline

- Repositories are untrusted until explicitly approved.
- Deterministic discovery reads bounded, allowlisted files and never executes
  repository commands.
- Process execution uses argument arrays, explicit working directories,
  cancellation, and risk classification; shell interpretation is opt-in.
- Browser access is loopback-only, authenticated, CSRF-protected, and subject
  to WebSocket origin checks.
- Credentials use the operating-system secret store. Supported logs, exports,
  diagnostics, and provider evidence apply redaction, which cannot recognize
  every secret format.
- Agent and plugin capabilities are least-privilege, permission-checked, and
  audited. Envarium does not expose a generic shell MCP tool.
  Trusted plugins and commands retain the local OS user's authority; API
  capabilities do not provide an operating-system sandbox.
- Destructive operations require preview and explicit authorization.

The reviewed v1 boundaries, threats, controls, residual risks, and findings are
in [docs/security/threat-model.md](docs/security/threat-model.md). The complete
design baseline remains normative in Section 18 of
`ENVARIUM_IMPLEMENTATION_PLAN.md`.
