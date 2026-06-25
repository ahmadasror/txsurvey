# FR — Public Runner (fetch + logic-aware submit)

> **Module**: survey
> **Status**: Implemented (backfilled)
> **Date**: 2026-06-25
> **ADRs**: `docs/architecture/adr/001-dual-logic-engine.md`, `docs/architecture/adr/004-normalized-answers-jsonb-leaf.md`
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

---

## 3. API Surface

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/public/forms/:slug` | public |
| POST | `/api/v1/public/forms/:slug/responses` | public |

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

endpoints:
  - id: AC-RUN-FETCH
    method: GET
    path: /api/v1/public/forms/{slug}
    auth: { mode: public }
  - id: AC-RUN-SUBMIT
    method: POST
    path: /api/v1/public/forms/{slug}/responses
    auth: { mode: public }

db:
  writes:
    - table: responses
      columns_declared: [id, form_id, completed, meta, submitted_at]
    - table: answers
      columns_declared: [id, response_id, question_id, value, created_at]

cross_links:
  adr_refs:
    - docs/architecture/adr/001-dual-logic-engine.md
    - docs/architecture/adr/004-normalized-answers-jsonb-leaf.md
  sisters:
    - docs/fr/survey/active/fr-forms.md
    - docs/fr/survey/active/fr-results.md
```
