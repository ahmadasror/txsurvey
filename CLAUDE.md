# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

txsurvey is a Typeform-like survey SaaS: a Go API-first backend + a React SPA, for 5–10 creators.
See `README.md` for setup and the OAuth/test-user caveat.

## Commands

```bash
make run                     # go run ./cmd/server (auto-runs embedded migrations)
make check                   # umbrella gate: lint+unit tests+coverage+routes+docs+FE tests/build (DB-free)
make test                    # go test ./...  (integration tests need DATABASE_URL)
go test ./... -short         # unit only (skips DB-backed tests in tests/)
cd frontend && npm run test  # frontend unit tests (vitest) — incl. dual-engine parity
go test ./tests/ -run TestE2EHappyPath -v    # one integration test
make build                   # SPA build -> stage at internal/web/dist -> go build -tags embedspa
make lint                    # go vet + gofmt check
cd frontend && npm run dev   # Vite dev server :5173 (proxies /api to :8080)
cd frontend && npm run build # tsc --noEmit + vite build (typecheck + bundle)
docker compose up -d postgres   # just the DB for local `make run`
```

Run a single Go test: `go test ./internal/service/ -run TestReachablePath -v`.
Integration tests (`tests/`) TRUNCATE, so the harness **refuses any `DATABASE_URL` whose db
name lacks `test`** — point it at a `*_test` DB (e.g. `…/txsurvey_test`); `make run`'s dev DB
is `txsurvey`. The harness runs migrations itself before truncating.

## Lean SDD (spec-driven workflow)

Full process: `docs/sdd-workflow.md`. The short version — for a **non-trivial**
feature, land 3 artifacts:

1. **FR + contract block** — `docs/fr/survey/active/fr-<feature>.md` ending in a
   `## Contract (machine-readable)` YAML block (schema `docs/fr/_contract-schema.json`).
   Code-first is fine; pair the FR in the same commit (or immediate follow-up).
2. **ADR** for a non-obvious decision — `docs/architecture/adr/NNN-*.md`.
3. **Tests green** — `go test ./...` + `cd frontend && npm run build`.

Commands: `/spec <feature>` (fr-writer authors the FR) · `make docs-check`
(one-shot alignment gate: schema + cross-link/ADR-index coherence + advisory drift —
run before committing docs) · `make spec-validate` / `make spec-drift` / `make docs-status`
are the individual pieces. Trivial changes (copy, refactor, dep bump) skip the FR.

When you add/change an endpoint or a table, update the owning FR's contract block in
the same change — `make spec-drift` will otherwise flag the new route as unspecced.

## Architecture (big picture)

Request flow: `Gin router → middleware → handler (thin) → service (logic) → repository (pgx, raw SQL) → Postgres`.
Layering matches the sibling Go projects (txhcs, lightworkflow): `cmd/ internal/{config,database,middleware,router,handler,service,repository,model,dto,web} pkg/{auth,apperror,response}`.

Two trust zones in `internal/router/routes.go`: **creator-authenticated** (session cookie, owner-scoped)
and **public** (anonymous, rate-limited). Every form-scoped repo query carries
`owner_id = $ AND deleted_at IS NULL` — that is the entire authorization model; there is no separate ACL layer.

### Things that will bite you if you don't know them

- **The conditional-logic engine exists TWICE and must stay in lockstep:**
  `internal/service/logic_engine.go` (authoritative — re-validates the reachable path on every submit)
  and `frontend/src/lib/logicEngine.ts` (runner navigation only, never trusted). Change one → change both,
  and update both test suites (`logic_engine_test.go` is the spec).

- **Submit validation is logic-aware, not linear.** `ResponseService.Submit` replays `reachablePath` from the
  submitted answers: `required` is enforced only on *reachable* questions, and an answer to a *skipped*
  question is rejected (anti-tamper). A "linear" form is just the no-rules case of the same code.

- **Jumps are forward-only; `always` = unconditional jump.** `jump_to` may only target a *later* question,
  enforced in BOTH `logic_service.go validateRule` (`target.Position > source.Position`, else 422
  `INVALID_LOGIC_RULE`) and `LogicEditor.tsx` (target dropdown filtered to `position >
  source.position`). The `always` operator (migration 006) matches even when the source is unanswered, so
  `always` + `jump_to` routes straight to a chosen later question with no condition — the builder's **"Lompat
  langsung"** button. Reordering questions can leave a rule pointing backward; the engine's visited-set + step
  cap keep the runner safe, but such a rule can't be re-created/edited in the UI.

- **Migrations are forward-only; numeric order == build order.** golang-migrate only applies versions greater
  than the current one, so a new migration must take the next number (don't back-fill a lower number after a
  higher one has run). They're embedded via `//go:embed migrations/*.sql` and run at boot.

- **SPA embedding is build-tagged.** Default builds (`make run`, `go test`, `go build ./...`) do NOT embed the
  SPA (`internal/web/embed_dev.go`), so backend dev never needs a frontend build. Production uses
  `-tags embedspa` (`embed_prod.go` + `internal/web/dist/`, staged by `make build`/Dockerfile).

- **Theming is a hex→HSL pipeline ("Soft Studio" design system), not hardcoded CSS.**
  `frontend/src/lib/themes.ts` defines the 5 presets (Pine/Sand/Grape/Coral/Ink, **default Pine**) as hex
  token sets, converts them to the `H S% L%` triples shadcn expects via `lib/theme.ts hexToHslTriple`, and
  applies them as CSS variables on a themed container through `themeStyle(theme, font)`. Per-form display
  font (`editorial|modern|soft|serif`) is a `settings.font` field that drives `--font-display`; the UI/body
  font (Hanken Grotesk) is fixed. Unknown/legacy preset ids fall back to Pine. To change a palette, edit only
  the hex in `themes.ts` — but the `:root` defaults in `index.css` must keep mirroring the Pine triples, and
  fonts load via Google Fonts in `index.html`. Runner/Builder/Results wrap their root in `themeStyle(...)`;
  Dashboard/Login/Legal use the `:root` default. `settings.font` is a plain JSONB field — no migration.

- **SPA routes (`router.tsx`):** public `/login`, `/legal`, `/r/:slug` (runner) sit outside the auth guard;
  `/`, `/templates` are under `DashboardLayout`; `/forms/:id` and `/forms/:id/results` are full-bleed.
  `TemplatesPage` seeds a draft form (create → PATCH settings → POST each question) then opens the Builder —
  it relies on the backend auto-filling empty option ids (`question_service.go`).

- **Session auth is a cookie, not a Bearer header** — the one deliberate deviation from the sibling services.
  txsurvey mints its OWN JWT after Google sign-in (`pkg/auth`) and stores it in the httpOnly `session` cookie;
  `middleware.SessionAuth` reads the cookie. `userID(c)` (key `"user_id"`) is the single read point.

- **Answer storage:** normalized `answers` rows (one per answered question) with a JSONB `value` leaf for the
  type-varying payload. Analytics are aggregated in Go (`ResultsService.computeSummary`), not via JSONB SQL.

- **Slug lifecycle.** A form's slug is minted once at create from its (often placeholder — the UI defaults an
  empty title to "Survei tanpa judul") title, so renaming later does NOT change the URL. While a form is a
  **draft** the slug can be re-pointed via `PATCH /forms/:id {slug}` → `FormService.resolveSlug` (slugified,
  must be unique among live forms), surfaced in the Design dialog's **"Tautan publik"** field. Once
  **published the slug is frozen** (422 `SLUG_LOCKED`) so already-shared `/r/:slug` links keep working; a
  taken slug is 422 `SLUG_TAKEN`. Slug uniqueness is a partial unique index (live rows only).

- **Browser-tab titles** come from `useDocumentTitle(...parts)` (`lib/useDocumentTitle.ts`), called once per
  page; it renders a `part · … · txsurvey` breadcrumb and drops falsy parts (so a still-loading form title
  collapses to just `txsurvey`). Add a call when you add a page — `index.html`'s `<title>txsurvey</title>` is
  only the pre-JS fallback.

- **gofmt is enforced** (`make lint`). The module is `go 1.25` (the toolchain auto-upgraded on first tidy);
  the Dockerfile uses `golang:1.25-alpine`.

- **Port 8080 may be occupied** by an environment proxy; tests use `httptest` (no port). When running the
  server locally to poke it, set `SERVER_PORT` to a free port.

## Deployment (this checkout doubles as the production host)

There is **no CD** — pushing to `main` and a green CI run do **not** deploy. The live service runs on this
same VPS as the `txsurvey-app-1` Docker container (image `txsurvey-app:prod`, built from this repo with the
SPA embedded), published on the docker bridge at `172.17.0.1:8092` behind the external `static-web` nginx +
Cloudflare tunnel. It serves **https://brainzap.net/txsurvey** and **https://ct.tuxceria.biz.id/txsurvey**
(path-prefix `/txsurvey/`, baked in via `frontend/.env.production` `VITE_BASE`).

Redeploy after changes land on `main`:

```bash
docker compose -f docker-compose.prod.yml -p txsurvey up -d --build
```

This rebuilds the embedded-SPA image and recreates the app container (brief downtime). Postgres
(`txsurvey-postgres-1`, not redefined in the prod compose) and uploaded assets (`./data/uploads`, a host
volume) persist. Migrations run idempotently at boot. Secrets come from `deploy/production.env` (gitignored —
never commit it). Google OAuth redirect URIs / test users are Console-only and cannot be changed from code.

## Adding a question type or operator

Question types: extend the `question_type` enum (new migration), `model.QuestionType` consts +
`validateQuestion` (metadata rules) + `validateAnswer` (per-type answer check) + the runner's `QuestionScreen`
+ `lib/questionTypes.ts`. Logic operators: extend the enum (new migration), `model.LogicOperator` + `ValidLogicOperator`,
`conditionMatches` in BOTH engines, and the builder's `LogicEditor` operator list. An operator that takes no
comparison value (like `is_empty` / `always`) must ALSO be excluded from `needsValue` in BOTH
`logic_service.go validateRule` and `LogicEditor.tsx`, or it will 422 on "requires a comparison value".
