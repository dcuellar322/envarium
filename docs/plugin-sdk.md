---
title: Plugin SDK and compatibility policy
description: Build and test capability-scoped out-of-process Envarium plugins.
category: reference
audience: [integrator, contributor]
since: 1.0.0
lastVerified: 2026-10-19
---

The public Go package `envarium.dev/envarium/sdk/plugin` contains the wire
types, manifest validation, bounded JSON-RPC server, client, and reusable
`sdk/plugin/plugintest` conformance harness. Plugin code must not import an
`internal/` package.

## Build the sample adapter

```bash
go build -trimpath -o envarium-fixture-plugin ./examples/plugins/fixture
mkdir -p "<data-dir>/plugins/fixture-inspector"
cp envarium-fixture-plugin examples/plugins/fixture/plugin.json \
  "<data-dir>/plugins/fixture-inspector/"
envarium plugin refresh
envarium plugin list
```

Review the full fingerprint shown by `plugin list`, then make separate trust
and least-privilege decisions:

```bash
envarium plugin trust fixture-inspector --fingerprint <sha256> --yes
envarium plugin enable fixture-inspector \
  --scope project.metadata.read --scope project.files.read \
  --scope project.operate --yes
envarium plugin inspect fixture-inspector <project>
envarium plugin run fixture-inspector <project> fixture.echo \
  --input '{"reviewed":true}' --yes
```

The sample inspects three known filenames and implements a bounded echo action;
it never runs a command. It is proof of the protocol, not a privileged core
runtime adapter.

## Authoring contract

Create a `plugin.Manifest`, implement `plugin.Handler`, and call
`plugin.Serve(context.Background(), os.Stdin, os.Stdout, manifest, handler)`.
Write diagnostics only to stderr. Run the conformance kit from the plugin's own
test suite before publishing.

Manifests declare only capabilities the executable implements and all scopes it
might request. The host may grant a strict subset. A handler must continue to
enforce the grants supplied during initialize and must return structured facts,
actions, health, or operation receipts rather than terminal text.

## Versions and deprecation

`envarium.plugin/v1` requires an exact match. There is no compatibility
guessing or fallback execution. A mismatch remains discoverable as an
actionable error and is never enabled.

These compatibility promises apply to published stable
releases; development builds may change before the first stable release:

- fields may be added only when older peers can safely ignore or default them;
- methods, required fields, capabilities, and scopes are not repurposed;
- a superseded stable protocol remains supported for at least two Envarium
  minor releases and 90 days, whichever is longer;
- deprecation appears in release notes, SDK documentation, discovery health,
  and the permission-review UI; and
- an urgent security removal may shorten the window, with the reason and
  migration path published alongside the release.

Plugin semantic versions identify executable releases but do not override the
protocol decision. Changing either the manifest or executable always requires
the user to review a new fingerprint.
