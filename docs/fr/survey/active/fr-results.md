# FR — Results (responses · analytics · CSV)

> **Module**: survey
> **Status**: Implemented (backfilled)
> **Date**: 2026-06-25
> **ADRs**: `docs/architecture/adr/004-normalized-answers-jsonb-leaf.md`
> **Sisters**: `docs/fr/survey/active/fr-runner.md`

---

## 1. Background

The creator-authenticated, read-only results surface: list/inspect responses,
per-question analytics, and CSV export. Ownership-scoped. No writes.

---

## 2. Functional Requirements

### FR-RES-001 — Responses list + detail
**Code**: `internal/handler/results_handler.go`, `internal/service/results_service.go`, `internal/repository/response_repo.go`

- AC-001-1: `GET /forms/:id/responses` returns a page of responses (newest first) with answers attached.
- AC-001-2: `GET /forms/:id/responses/:rid` returns one response with answers.
- AC-001-3: non-owner / absent form → 404 `FORM_NOT_FOUND`.

### FR-RES-002 — Per-question analytics
- AC-002-1: `GET /forms/:id/analytics` returns `{response_count, completion_rate, questions[]}`; choice/yes-no/rating get option tallies; number/rating get averages.

### FR-RES-003 — CSV export
- AC-003-1: `GET /forms/:id/export.csv` streams `text/csv` (one column per answerable question; option ids resolved to labels); `Content-Disposition: attachment`.

---

## 3. API Surface

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/forms/:id/responses` | session |
| GET | `/api/v1/forms/:id/responses/:rid` | session |
| GET | `/api/v1/forms/:id/analytics` | session |
| GET | `/api/v1/forms/:id/export.csv` | session |

---

## 4. Data Model Touch Points

| Table | Read | Write | Notes |
|---|:---:|:---:|---|
| `responses` | ✓ | — | read-only |
| `answers` | ✓ | — | aggregated in Go |

---

## 5. Cross-links

- ADR: `docs/architecture/adr/004-normalized-answers-jsonb-leaf.md`
- Sister: `docs/fr/survey/active/fr-runner.md`

---

## 6. Contract (machine-readable)

> Drift-detector source. Schema: `docs/fr/_contract-schema.json`.

```yaml
fr_file: docs/fr/survey/active/fr-results.md
covers:
  - FR-RES-001
  - FR-RES-002
  - FR-RES-003

endpoints:
  - id: AC-RES-LIST
    method: GET
    path: /api/v1/forms/{id}/responses
    auth: { mode: session }
  - id: AC-RES-GET
    method: GET
    path: /api/v1/forms/{id}/responses/{rid}
    auth: { mode: session }
  - id: AC-RES-ANALYTICS
    method: GET
    path: /api/v1/forms/{id}/analytics
    auth: { mode: session }
  - id: AC-RES-EXPORT
    method: GET
    path: /api/v1/forms/{id}/export.csv
    auth: { mode: session }

cross_links:
  adr_refs:
    - docs/architecture/adr/004-normalized-answers-jsonb-leaf.md
  sisters:
    - docs/fr/survey/active/fr-runner.md
```
