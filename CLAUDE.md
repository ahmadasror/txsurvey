# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

txsurvey is a Typeform-like survey SaaS: a Go API-first backend + a React SPA, for 5–10 creators.
See `README.md` for setup and the OAuth/test-user caveat.

## Commands

```bash
make run                     # go run ./cmd/server (auto-runs embedded migrations)
make test                    # go test ./...  (integration tests need DATABASE_URL)
go test ./... -short         # unit only (skips DB-backed tests in tests/)
go test ./tests/ -run TestE2EHappyPath -v    # one integration test
make build                   # SPA build -> stage at internal/web/dist -> go build -tags embedspa
make lint                    # go vet + gofmt check
cd frontend && npm run dev   # Vite dev server :5173 (proxies /api to :8080)
cd frontend && npm run build # tsc --noEmit + vite build (typecheck + bundle)
docker compose up -d postgres   # just the DB for local `make run`
```

Run a single Go test: `go test ./internal/service/ -run TestReachablePath -v`.

## Lean SDD (spec-driven workflow)

Full process: `docs/sdd-workflow.md`. The short version — for a **non-trivial**
feature, land 3 artifacts:

1. **FR + contract block** — `docs/fr/survey/active/fr-<feature>.md` ending in a
   `## Contract (machine-readable)` YAML block (schema `docs/fr/_contract-schema.json`).
   Code-first is fine; pair the FR in the same commit (or immediate follow-up).
2. **ADR** for a non-obvious decision — `docs/architecture/adr/NNN-*.md`.
3. **Tests green** — `go test ./...` + `cd frontend && npm run build`.

Commands: `/spec <feature>` (fr-writer authors the FR) · `make spec-validate`
(schema gate — keep clean) · `make spec-drift` (advisory: FR endpoints/tables exist
in `routes.go`/migrations). Trivial changes (copy, refactor, dep bump) skip the FR.

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

- **Migrations are forward-only; numeric order == build order.** golang-migrate only applies versions greater
  than the current one, so a new migration must take the next number (don't back-fill a lower number after a
  higher one has run). They're embedded via `//go:embed migrations/*.sql` and run at boot.

- **SPA embedding is build-tagged.** Default builds (`make run`, `go test`, `go build ./...`) do NOT embed the
  SPA (`internal/web/embed_dev.go`), so backend dev never needs a frontend build. Production uses
  `-tags embedspa` (`embed_prod.go` + `internal/web/dist/`, staged by `make build`/Dockerfile).

- **Session auth is a cookie, not a Bearer header** — the one deliberate deviation from the sibling services.
  txsurvey mints its OWN JWT after Google sign-in (`pkg/auth`) and stores it in the httpOnly `session` cookie;
  `middleware.SessionAuth` reads the cookie. `userID(c)` (key `"user_id"`) is the single read point.

- **Answer storage:** normalized `answers` rows (one per answered question) with a JSONB `value` leaf for the
  type-varying payload. Analytics are aggregated in Go (`ResultsService.computeSummary`), not via JSONB SQL.

- **gofmt is enforced** (`make lint`). The module is `go 1.25` (the toolchain auto-upgraded on first tidy);
  the Dockerfile uses `golang:1.25-alpine`.

- **Port 8080 may be occupied** by an environment proxy; tests use `httptest` (no port). When running the
  server locally to poke it, set `SERVER_PORT` to a free port.

## Adding a question type or operator

Question types: extend the `question_type` enum (new migration), `model.QuestionType` consts +
`validateQuestion` (metadata rules) + `validateAnswer` (per-type answer check) + the runner's `QuestionScreen`
+ `lib/questionTypes.ts`. Logic operators: extend the enum, `model.LogicOperator`, `conditionMatches` in
BOTH engines, and the builder's `LogicEditor` operator list.
