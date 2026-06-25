# ADR-003 — Embed SPA in the Go binary; path-prefix-aware for subpath deploy

> **Status**: accepted
> **Date**: 2026-06-25

## Context

Solo dev, 5–10 users, free product. The SPA + API need to ship and run cheaply.
The first deploy target is a **subpath** of an existing host: `ct.tuxceria.biz.id/txsurvey`.

## Decision

- **Single artifact**: the built SPA is embedded into the Go binary via `embed.FS`
  behind the `embedspa` build tag (`internal/web/embed_prod.go`). Default builds
  (`go run`, `go test`) do **not** embed, so backend dev needs no frontend build.
  Same-origin means the session cookie is first-party (no CORS).
- **Subpath via nginx prefix-strip**: nginx strips `/txsurvey/` and proxies to the
  Go app at the root (no Go routing changes). The SPA is path-prefix-aware from a
  single source — Vite `BASE_URL` (`frontend/src/lib/paths.ts` derives the router
  basename, API base, and runner share links). `SetTrustedProxies(loopback)` so the
  rate limiter sees the real client IP.

## Alternatives considered

- **Separate SPA host (Vercel) + API host** — two deploys, cross-site cookie
  (`SameSite=None`), CORS credentials. More moving parts. Rejected for this scale.
- **Dedicated subdomain** (`survey.tuxceria.biz.id`) — cleaner (zero code changes,
  shorter share links) and still supported (`VITE_BASE=/`, `COOKIE_PATH=/`). The user
  chose the subpath; this ADR records that the code supports both.
- **Go server made base-path aware** (handle `/txsurvey` in routes) — more invasive
  than nginx strip. Rejected.

## Consequences

- Positive: one binary, one origin, one TLS cert; cheap to run.
- Negative: a frontend change requires rebuilding the Go image; subpath needs the
  nginx config + `VITE_BASE`/`COOKIE_PATH`/`APP_BASE_URL` set consistently.
- Artifacts: `deploy/nginx-txsurvey.conf`, `deploy/production.env.example`.

## Related

- FR: `docs/fr/survey/active/fr-runner.md` (public link construction)
