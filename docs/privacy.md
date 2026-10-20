---
title: Privacy
description: What Envarium stores locally, sends only by choice, and excludes from telemetry.
category: concept
audience: [user, integrator]
since: 1.0.0
lastVerified: 2026-10-19
---

Envarium is local-first and has no required product telemetry or cloud
account. Project metadata, manifests, operations, audit history, logs,
diagnoses, feedback, recipes, plugins, and settings remain on the local machine
unless the user explicitly exports or sends them.

Deterministic discovery reads only bounded allowlisted repository files and
does not execute code. Optional AI onboarding or diagnosis sends the exact
previewed, bounded, redacted evidence bundle to the provider the user selected;
provider use is not required for any deterministic workflow. External coding
clients have their own data-sharing controls; MCP access does not make those
clients subject to Envarium’s provider-preview flow. Provider terms
and retention apply after that explicit request.

Support bundles always exclude source, database contents, resolved secrets,
arbitrary environment values, terminal output, and project application logs.
The exact manifest is previewable before a private archive is written, and the
archive contains only allowlisted system/configuration facts plus recent
redacted internal daemon warnings and errors. Supported log, diagnostic, and
provider-evidence paths apply redaction. Interactive terminal output and
unfamiliar secret formats can still expose private data; review output and
bundles before sharing them.
Plugins are external programs with separately reviewed fingerprints and scopes;
their own network or data behavior remains the plugin publisher's
responsibility. Plugins run under the local OS user; Envarium’s API capability
checks are not an operating-system sandbox.

Optional peer federation sends only the bounded identity, trusted-project, and
registered-environment fields documented in the federation guide. It excludes
repository locations, source, Git changes, logs, terminal output, secrets,
environment values, and runtime-native identifiers. Envarium does not supply
the tunnel or relay remote traffic through a hosted service.

Encrypted team sync contains only explicitly trusted public publisher records
and verified portable configuration bundles. The standard age file excludes
projects, repository paths and contents, machine credentials, fleet records,
operations, logs, terminals, environment values, and runtime state. Decryption
occurs in the CLI process and import is previewed before confirmation.

Optional usage metrics are off by default and have no vendor-controlled
destination. Opt-in requires an explicit HTTPS endpoint and displays the full
payload: a random installation ID, Envarium version, operating
system, architecture, fixed operation-category counters, and generation time.
While enabled, metrics are sent periodically to the configured endpoint. The
installation ID links payloads from the same installation, and the endpoint
can process request metadata such as the source IP address.
It contains no project, machine, action, plugin, provider, path, command, log,
diagnostic, or error detail. Disabling metrics clears the installation ID and
all pending counters. A signed team policy may prohibit opt-in.
