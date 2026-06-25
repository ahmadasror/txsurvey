# txsurvey

A Typeform-like survey/form builder — a small, free SaaS for 5–10 creators.
Creators sign in with Google, build forms with **conditional logic**, share a
public link, and read responses / analytics / CSV. Respondents answer
**one question per screen** on any device.

- **Backend:** Go (Gin + pgx/v5 + PostgreSQL + golang-migrate + slog), API-first.
- **Frontend:** React SPA (Vite + TypeScript + Tailwind + shadcn/ui), TanStack Query.
- **Auth:** Sign in with Google (OAuth2, sign-in only); app-minted JWT session in an httpOnly cookie.
- **Deploy:** single Go binary with the SPA embedded (one artifact, first-party cookie).

## Quick start (Docker)

```bash
cp .env.example .env          # fill in GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET
docker compose up -d --build  # postgres + app on :8080
# open http://localhost:8080
```

The server runs embedded migrations on boot (look for `migrations applied successfully`).

## Local development

Two processes — Go API on :8080, Vite dev server on :5173 (proxies `/api` to the API):

```bash
# terminal 1 — API + Postgres
docker compose up -d postgres
cp .env.example .env
make run                      # go run ./cmd/server  (auto-migrates)

# terminal 2 — SPA
cd frontend && npm install && npm run dev   # http://localhost:5173
```

The session cookie is `SameSite=Lax`; in split-host dev the SPA calls the API with
`credentials: include` and the API allows `CORS_ALLOWED_ORIGINS` (default `http://localhost:5173`).

## Google OAuth setup

1. Google Cloud Console → **APIs & Services → Credentials → Create OAuth client ID** (type *Web application*).
2. **Authorized redirect URIs:** `http://localhost:8080/api/v1/auth/google/callback` (dev) and your production URL.
3. **OAuth consent screen:** External; scopes `openid`, `email`, `profile` (non-sensitive — no verification needed).
4. Put the client id/secret in `.env`.

**Unverified-app caveat:** while the consent screen is in *Testing*, only **test users**
added in the console can sign in (Google shows an "unverified app" notice). For 5–10
known creators, just add each creator's email as a test user — no verification submission required.

## Build & deploy (single artifact)

```bash
make build        # builds the SPA, stages it at internal/web/dist, then
                  # `go build -tags embedspa` -> bin/server (SPA embedded)
./bin/server      # serves API + SPA on the same origin
```

The `embedspa` build tag controls embedding: default builds (`make run`, `go test`) do **not**
embed, so backend dev never needs a frontend build; the `Dockerfile` and `make build` use the tag.

## Environment

| Var | Purpose |
|---|---|
| `DATABASE_URL` | Postgres DSN (required) |
| `JWT_SECRET` | session signing secret, ≥ 32 chars (required) |
| `SESSION_TTL` | session lifetime (default `24h`) |
| `COOKIE_SECURE` | `true` in production (HTTPS) |
| `APP_BASE_URL` | this app's public origin (OAuth return target) |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` / `GOOGLE_REDIRECT_URL` | Google OAuth |
| `CORS_ALLOWED_ORIGINS` | split-host dev only |

## Make targets

`make run` · `make dev` · `make fe-dev` · `make build` · `make test` · `make lint` ·
`make migrate-new name=...` · `make up` / `make down`

## Testing

```bash
go test ./...                 # unit + integration (integration needs DATABASE_URL)
go test ./... -short          # unit only (skips DB-backed tests)
cd frontend && npm run build  # typecheck + SPA build
```

Integration/E2E tests (`tests/`) run the API in-process against Postgres and cover the
happy path (create → publish → submit → analytics → CSV) and branching logic.

## Architecture

```
cmd/server            entrypoint (config, migrate, pool, router, graceful shutdown)
internal/
  config              env-struct config (validated at startup)
  database            pgxpool + embedded golang-migrate migrations
  middleware          request-id, slog logger, CORS, SessionAuth (cookie), rate limit
  router              Gin setup + route groups (public vs creator-authenticated)
  handler             thin HTTP handlers (JSON envelope)
  service             business logic (forms, questions, logic engine, responses, results)
  repository          pgx data access (raw SQL, owner-scoped)
  model / dto         domain types / request-response shapes
  web                 optional embedded SPA (build-tagged)
pkg/{auth,apperror,response}   session JWT, client errors, JSON envelope
frontend/             Vite + React SPA (builder, runner, dashboard/results)
```

The **conditional-logic engine** lives in two mirrored places that must stay in lockstep:
`internal/service/logic_engine.go` (authoritative — re-validates the reachable path on submit)
and `frontend/src/lib/logicEngine.ts` (runner navigation; never trusted).
