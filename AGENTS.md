# AGENTS.md — txsurvey

> ⚠️ **FIRST RUN (Codex migration pilot):** before doing any work, read and execute
> [`CODEX-HANDOFF.md`](./CODEX-HANDOFF.md). This repo is Ahmad's pilot for moving off
> Claude Code onto Codex; the handoff is the one-time setup + exit-test mission brief.
> Once the pilot passes, this banner can go and `AGENTS.md` becomes the plain
> canonical project doc.

## Authoritative conventions

The full, tool-agnostic engineering guide for this repo lives in **[`CLAUDE.md`](./CLAUDE.md)**
(~10 KB). **Read it fully before editing.** It is the source of truth until §2.1 of the
handoff consolidates it into this file. Do not duplicate it here (avoid drift).

## Fast orientation

- **Stack:** Go (Gin) API-first backend + React SPA. Survey SaaS (Typeform-like).
- **Layering:** `router → middleware → handler (thin) → service (logic) → repository (pgx, raw SQL) → Postgres`.
- **Dev loop:** `make run` · `make check` (umbrella gate) · `make test` (integration; DB name
  must contain `test` — harness TRUNCATEs) · `go test ./... -short` (unit) · `make build` · `make lint`.
- **Spec workflow (leanSDD):** FR → `docs/fr/survey/active/`, ADR → `docs/architecture/adr/`,
  gates `make docs-check` / `make spec-drift`. See `docs/sdd-workflow.md`.

## Invariants that will bite you (details in `CLAUDE.md`)

- **Dual logic engine** — `internal/service/logic_engine.go` (authoritative) and
  `frontend/src/lib/logicEngine.ts` (untrusted runner) must change **in lockstep**;
  `logic_engine_test.go` is the spec.
- **Forward-only migrations**, next number only — embedded, run at boot.
- **Cookie-session auth** (not Bearer) — `middleware.SessionAuth`, `userID(c)` single read point.
- **Owner-scoped authorization is the whole model** — every form query carries
  `owner_id = $ AND deleted_at IS NULL`; no separate ACL layer.
- **Build-tagged SPA embedding** — default builds don't embed the SPA; prod uses `-tags embedspa`.
- **Slug frozen once published** (422 `SLUG_LOCKED`); hex→HSL theming via `themeStyle(...)`.

## Guardrails

- No CD — deployment is manual (`CLAUDE.md §Deployment`). Never touch prod or commit `deploy/production.env`.
- Conventional commits, one concern each. Don't force-push `main`.
