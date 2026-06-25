# FR — Form Builder (forms · questions · logic rules · publish)

> **Module**: survey
> **Status**: Implemented (backfilled)
> **Date**: 2026-06-25
> **ADRs**: `docs/architecture/adr/001-dual-logic-engine.md`
> **Sisters**: `docs/fr/survey/active/fr-runner.md`, `docs/fr/survey/active/fr-results.md`

---

## 1. Background

The creator-authenticated builder API: own-scoped CRUD over forms, their questions
(11 types), and conditional-logic rules, plus publish/unpublish. Every query is
scoped `owner_id = $ AND deleted_at IS NULL`.

---

## 2. Functional Requirements

### FR-FORM-001 — Forms CRUD + publish
**Code**: `internal/handler/form_handler.go`, `internal/service/form_service.go`, `internal/repository/form_repo.go`

- AC-001-1: `POST /forms` creates a draft form with a globally-unique slug.
- AC-001-2: `GET /forms` lists the caller's forms (with question + response counts); `GET /forms/:id` returns nested questions + logic rules.
- AC-001-3: `PATCH /forms/:id` edits title/description/settings; `DELETE /forms/:id` soft-deletes.
- AC-001-4: `POST /forms/:id/publish` requires ≥1 answerable question (else 422 `PUBLISH_EMPTY`); `POST /forms/:id/unpublish` returns it to draft.
- AC-001-5: any form not owned by the caller → 404 `FORM_NOT_FOUND`.

### FR-FORM-002 — Questions CRUD + reorder
**Code**: `question_handler.go`, `question_service.go`, `question_repo.go`

- AC-002-1: `POST /forms/:id/questions` appends a question; per-type metadata validated (choice needs options, rating scale 2–10, number min≤max, statement forced not-required).
- AC-002-2: `PATCH`/`DELETE /forms/:id/questions/:qid` edit/remove.
- AC-002-3: `PUT /forms/:id/questions/reorder` rewrites positions; `ordered_ids` must be a permutation of the form's questions (else 422).

### FR-FORM-003 — Logic rules CRUD
**Code**: `logic_handler.go`, `logic_service.go`, `logic_repo.go`

- AC-003-1: `GET/POST /forms/:id/logic` and `PATCH/DELETE /forms/:id/logic/:rid`.
- AC-003-2: rule validation — source/target in form, no self-target, `jump_to` must be forward (`target.position > source.position`), `end_form` has no target (else 422 `INVALID_LOGIC_RULE`).

---

## 3. API Surface

| Method | Path | Auth |
|---|---|---|
| GET / POST | `/api/v1/forms` | session |
| GET / PATCH / DELETE | `/api/v1/forms/:id` | session |
| POST | `/api/v1/forms/:id/{publish,unpublish}` | session |
| POST | `/api/v1/forms/:id/questions` | session |
| PUT | `/api/v1/forms/:id/questions/reorder` | session |
| PATCH / DELETE | `/api/v1/forms/:id/questions/:qid` | session |
| GET / POST | `/api/v1/forms/:id/logic` | session |
| PATCH / DELETE | `/api/v1/forms/:id/logic/:rid` | session |

### Error Codes

| Code | HTTP | Trigger |
|---|---|---|
| `FORM_NOT_FOUND` | 404 | Not owned / absent |
| `PUBLISH_EMPTY` | 422 | Publish with no answerable question |
| `INVALID_LOGIC_RULE` | 422 | Rule graph violation |
| `VALIDATION_ERROR` | 422 | Body / metadata invalid |

---

## 4. Data Model Touch Points

| Table | Read | Write | Notes |
|---|:---:|:---:|---|
| `forms` | ✓ | ✓ | partial-unique slug, soft delete |
| `questions` | ✓ | ✓ | metadata JSONB, contiguous positions |
| `logic_rules` | ✓ | ✓ | operator/action enums |

---

## 5. Cross-links

- ADR: `docs/architecture/adr/001-dual-logic-engine.md`
- Sister: `docs/fr/survey/active/fr-runner.md`

---

## 6. Contract (machine-readable)

> Drift-detector source. Schema: `docs/fr/_contract-schema.json`.

```yaml
fr_file: docs/fr/survey/active/fr-forms.md
covers:
  - FR-FORM-001
  - FR-FORM-002
  - FR-FORM-003

endpoints:
  - id: AC-FORM-LIST
    method: GET
    path: /api/v1/forms
    auth: { mode: session }
  - id: AC-FORM-CREATE
    method: POST
    path: /api/v1/forms
    auth: { mode: session }
  - id: AC-FORM-GET
    method: GET
    path: /api/v1/forms/{id}
    auth: { mode: session }
  - id: AC-FORM-UPDATE
    method: PATCH
    path: /api/v1/forms/{id}
    auth: { mode: session }
  - id: AC-FORM-DELETE
    method: DELETE
    path: /api/v1/forms/{id}
    auth: { mode: session }
  - id: AC-FORM-PUBLISH
    method: POST
    path: /api/v1/forms/{id}/publish
    auth: { mode: session }
  - id: AC-FORM-UNPUBLISH
    method: POST
    path: /api/v1/forms/{id}/unpublish
    auth: { mode: session }
  - id: AC-QUESTION-CREATE
    method: POST
    path: /api/v1/forms/{id}/questions
    auth: { mode: session }
  - id: AC-QUESTION-REORDER
    method: PUT
    path: /api/v1/forms/{id}/questions/reorder
    auth: { mode: session }
  - id: AC-QUESTION-UPDATE
    method: PATCH
    path: /api/v1/forms/{id}/questions/{qid}
    auth: { mode: session }
  - id: AC-QUESTION-DELETE
    method: DELETE
    path: /api/v1/forms/{id}/questions/{qid}
    auth: { mode: session }
  - id: AC-LOGIC-LIST
    method: GET
    path: /api/v1/forms/{id}/logic
    auth: { mode: session }
  - id: AC-LOGIC-CREATE
    method: POST
    path: /api/v1/forms/{id}/logic
    auth: { mode: session }
  - id: AC-LOGIC-UPDATE
    method: PATCH
    path: /api/v1/forms/{id}/logic/{rid}
    auth: { mode: session }
  - id: AC-LOGIC-DELETE
    method: DELETE
    path: /api/v1/forms/{id}/logic/{rid}
    auth: { mode: session }

db:
  writes:
    - table: forms
      columns_declared: [id, owner_id, title, description, slug, status, settings, published_at, created_at, updated_at, deleted_at]
    - table: questions
      columns_declared: [id, form_id, type, title, description, position, required, metadata, created_at, updated_at]
    - table: logic_rules
      columns_declared: [id, form_id, source_question_id, operator, compare_value, action, target_question_id, priority, created_at]

cross_links:
  adr_refs:
    - docs/architecture/adr/001-dual-logic-engine.md
  sisters:
    - docs/fr/survey/active/fr-runner.md
    - docs/fr/survey/active/fr-results.md
```
