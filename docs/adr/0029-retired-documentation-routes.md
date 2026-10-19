---
title: "ADR-0029: Retired documentation routes"
description: Return explicit 404 responses for unpublished documentation before stale assets can be served.
category: concept
audience: [contributor, maintainer]
lastVerified: 2026-10-19
---

- Status: Accepted
- Date: 2026-10-19
- Refines: ADR-0017 and ADR-0028

## Context

Removing a document from a static deployment removes its generated page and
indexes, but a previously cached Pages asset can remain accessible. Website
publication policy must also hold when a retired URL is requested directly.
Cloudflare documents [asset retention after deployment](https://developers.cloudflare.com/pages/configuration/serving-pages/#asset-retention).

## Decision

Extend the existing bounded Pages Worker with explicit 404 handling for paths
excluded by the shared documentation publication policy. Generate its invocation
routes from that same policy. Cover directory descendants, extensionless paths,
HTML aliases, and raw Markdown. These requests never fetch the retired asset;
they return the site's static error page with HTTP 404, `Cache-Control: no-store`,
and `X-Robots-Tag: noindex`. HEAD responses have no body.

Public pages continue to use ordinary static delivery. The updater feed keeps
its existing behavior. No secrets, storage, runtime content discovery, user
mutations, or external requests are added. This extends ADR-0028's narrow
hosting response exception, while retaining static content generation.

## Consequences

Unpublishing a document requires a successful deployment of both the content
filter and generated route guard. Local tests simulate a stale asset store;
production checks must request retired URLs without cache-busting parameters.
This controls site publication, not visibility of files retained in a public
source repository.
