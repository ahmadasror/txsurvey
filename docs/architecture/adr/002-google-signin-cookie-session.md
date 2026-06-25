# ADR-002 — Google sign-in only + app-minted cookie session

> **Status**: accepted
> **Date**: 2026-06-25

## Context

Creators need to authenticate; respondents do not (public link). The sibling Go
services consume a backend-issued Bearer JWT, but txsurvey is its own identity
provider. "Connect to Google" was scoped to **sign-in only** (no Sheets) — data
lives in Postgres, exported as plain CSV.

## Decision

- **Google OAuth2 sign-in only** (scopes `openid`, `email`, `profile`) with
  state + PKCE. After callback, upsert the user and mint the app's **own** HS256
  JWT.
- Store the session JWT in an **httpOnly, SameSite=Lax cookie** (`session`), not an
  `Authorization` header — the SPA is same-origin (embedded), so a first-party
  cookie avoids token-in-JS exposure and CORS-credential friction.
- `COOKIE_PATH` scopes the cookie (e.g. `/txsurvey`) so it isn't shared with other
  apps on a host.

## Alternatives considered

- **Email/password** — adds password reset, hashing, more attack surface; the brief
  said "connect to Google". Rejected for v1.
- **Bearer token in localStorage** — XSS-exfiltratable; needs manual attach on every
  call. Rejected.
- **Google Sheets sync** — needs per-user refresh-token storage + app verification
  for sensitive scopes. Out of scope; CSV export covers the need.

## Consequences

- Positive: minimal auth surface; non-sensitive scopes need no Google verification.
- Negative: deviates from the sibling "consume Bearer" convention (one cookie read in
  `middleware.SessionAuth` instead). Documented in CLAUDE.md.
- Operational: while the OAuth app is in Testing, creators must be added as test users.

## Related

- FR: `docs/fr/survey/active/fr-auth.md`
