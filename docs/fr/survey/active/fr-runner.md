# FR — Public Runner (fetch + logic-aware submit)

> **Module**: survey
> **Status**: Implemented (backfilled)
> **Date**: 2026-06-25
> **ADRs**: `docs/architecture/adr/001-dual-logic-engine.md`, `docs/architecture/adr/003-spa-embed-subpath-deploy.md`, `docs/architecture/adr/004-normalized-answers-jsonb-leaf.md`
> **Sisters**: `docs/fr/survey/active/fr-forms.md`, `docs/fr/survey/active/fr-results.md`

---

## 1. Background

The anonymous respondent surface. Fetch a published form by slug, then submit a
completed response. Submit validation is **logic-aware**: the server replays the
reachable path (`logic_engine.go`) and enforces required-ness only on reachable
questions, rejecting answers to skipped ones. Rate-limited per IP.

---

## 2. Functional Requirements

### FR-RUN-001 — Fetch published form by slug
**Code**: `internal/handler/public_handler.go` (`GetForm`), `internal/service/response_service.go` (`GetPublicForm`)

- AC-001-1: `GET /public/forms/:slug` returns `{questions, logic_rules, settings}` for a published, non-deleted form; **omits** `owner_id`.
- AC-001-2: 404 `FORM_NOT_FOUND` when absent / not published.
- AC-001-3: rate-limited (120/min/IP).

### FR-RUN-002 — Submit a response
**Code**: `public_handler.go` (`Submit`), `response_service.go` (`Submit`, `validateSubmission`), `response_repo.go` (`Insert`)

- AC-002-1: `POST /public/forms/:slug/responses` with `{answers:[{question_id, value}]}` persists the response + answers in one transaction.
- AC-002-2: per-type answer validation (`validateAnswer`); required-on-reachable-path enforced (422 `REQUIRED`); answers to unreachable questions rejected (422 `INVALID_ANSWER`).
- AC-002-3: answer to an unknown question / a statement → 422.
- AC-002-4: rate-limited (20/min/IP → 429 `RATE_LIMITED`).

### FR-RUN-003 — Progress capture (paradata, progress-only)
**Code**: `public_handler.go` (`Start`, `Progress`), `response_service.go` (`StartSession`, `UpdateProgress`), `response_repo.go` (`StartSession`, `AdvanceProgress`)

- AC-003-1: `POST /public/forms/:slug/start` opens an **in-progress** response (`completed=false`, `started_at`/`last_seen_at` stamped) and returns `{response_id}`; 404 `FORM_NOT_FOUND` if the slug isn't a published form. Rate-limited 30/min/IP.
- AC-003-2: `POST /public/forms/:slug/progress` with `{response_id, position}` advances `furthest_position` **monotonically** (`GREATEST`, never regresses) and bumps `last_seen_at`. A ping to an already-completed response is a silent no-op; an unknown/malformed id → 404 `RESPONSE_NOT_FOUND`; negative `position` clamps to 0.
- AC-003-3: **in-progress rows are inert paradata.** Every owner-facing surface (response count, results list, completion-rate denominator, CSV, analytics) is scoped to `completed`, so partial rows change none of those numbers — they are captured for a future funnel/drop-off view only. Submit still writes `completed=true` and now stamps `completed_at`.

---

## 3. API Surface

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/public/forms/:slug` | public |
| POST | `/api/v1/public/forms/:slug/responses` | public |
| POST | `/api/v1/public/forms/:slug/start` | public |
| POST | `/api/v1/public/forms/:slug/progress` | public |

### Error Codes

| Code | HTTP | Trigger |
|---|---|---|
| `FORM_NOT_FOUND` | 404 | Not published |
| `REQUIRED` | 422 | Missing required-on-path answer |
| `INVALID_ANSWER` | 422 | Bad value / unreachable / unknown question |
| `RATE_LIMITED` | 429 | Over per-IP cap |

---

## 4. Data Model Touch Points

| Table | Read | Write | Notes |
|---|:---:|:---:|---|
| `responses` | ✓ | ✓ | one per submission |
| `answers` | — | ✓ | one row per answered question, JSONB `value` |

---

## 5. Cross-links

- ADR: `docs/architecture/adr/001-dual-logic-engine.md`, `docs/architecture/adr/004-normalized-answers-jsonb-leaf.md`
- Sister: `docs/fr/survey/active/fr-forms.md`

---

## 6. Contract (machine-readable)

> Drift-detector source. Schema: `docs/fr/_contract-schema.json`.

```yaml
fr_file: docs/fr/survey/active/fr-runner.md
covers:
  - FR-RUN-001
  - FR-RUN-002
  - FR-RUN-003

endpoints:
  - id: AC-RUN-FETCH
    method: GET
    path: /api/v1/public/forms/{slug}
    auth: { mode: public }
  - id: AC-RUN-SUBMIT
    method: POST
    path: /api/v1/public/forms/{slug}/responses
    auth: { mode: public }
  - id: AC-RUN-START
    method: POST
    path: /api/v1/public/forms/{slug}/start
    auth: { mode: public }
  - id: AC-RUN-PROGRESS
    method: POST
    path: /api/v1/public/forms/{slug}/progress
    auth: { mode: public }

db:
  writes:
    - table: responses
      columns_declared: [id, form_id, completed, meta, submitted_at, started_at, last_seen_at, completed_at, furthest_position]
    - table: answers
      columns_declared: [id, response_id, question_id, value, created_at]

cross_links:
  adr_refs:
    - docs/architecture/adr/001-dual-logic-engine.md
    - docs/architecture/adr/003-spa-embed-subpath-deploy.md
    - docs/architecture/adr/004-normalized-answers-jsonb-leaf.md
  sisters:
    - docs/fr/survey/active/fr-forms.md
    - docs/fr/survey/active/fr-results.md
```
