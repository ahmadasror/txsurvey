# ADR-005 — Server-enriched SPA HTML for SEO

> **Status**: accepted
> **Date**: 2026-07-18

## Context

txsurvey ships one React SPA embedded in the Go binary. Client-side metadata is
visible after React runs, but the initial HTML is an empty app shell. Social bots
often do not execute JavaScript, and crawlers should receive meaningful metadata,
content, sitemap responses, and HTTP status codes without a second rendering pass.

Introducing a second SSR application would duplicate routing and add a Node runtime
to a deliberately small single-binary deployment.

## Decision

Keep React as the interactive renderer. When serving the embedded production SPA,
the Go router enriches `index.html` for known application paths with:

- route-specific title, description, robots, canonical, Open Graph, and Twitter tags;
- a small semantic fallback for indexable public pages;
- an explicit status code (`200` for known routes, `404` for unknown routes).

The React app replaces the fallback after loading and republishes the same metadata
during client-side navigation. The Go router also serves the sitemap and app-scoped
robots file from the canonical `APP_BASE_URL`.

## Alternatives considered

- **Full React SSR** — strongest rendering parity, but adds an SSR build/runtime and
  duplicates the current Go single-binary deployment concerns.
- **Static prerender files per route** — simple for fixed routes, but awkward for
  subpath hosting and dynamic SPA fallbacks; it also creates directory redirect
  behavior that can conflict with canonical URLs.
- **Client metadata only** — works for Google after rendering, but leaves social bots
  and first-response audits with an empty shell.

## Consequences

- Positive: crawler/social metadata is present in the first response without a Node
  production runtime; unknown paths receive a real 404.
- Positive: the existing React product remains unchanged after hydration.
- Negative: public route metadata has a small Go-side registry that must stay aligned
  with frontend routes; unit tests and the SEO FR define that contract.
- Negative: dynamic survey pages remain `noindex` by policy, so they do not require
  per-survey server rendering.

## Related

- FR: `docs/fr/survey/active/fr-seo-discovery.md`
- ADR: `docs/architecture/adr/003-spa-embed-subpath-deploy.md`
