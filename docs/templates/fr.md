# FR — <Feature Title>

> **Module**: survey
> **Status**: Draft v0.1
> **Date**: YYYY-MM-DD
> **ADRs**: <docs/architecture/adr/NNN-*.md, if any>
> **Sisters**: <other FR files this relates to>

---

## 1. Background

One paragraph: what this feature is, who uses it, why it exists.

---

## 2. Functional Requirements

### FR-XXX-001 — <Requirement name>
**Code**: `internal/handler/<x>_handler.go`, `internal/service/<x>_service.go`

- AC-001-1: <condition> → <expected outcome>.
- AC-001-2: <error case> → <status + code>.

### FR-XXX-002 — <Requirement name>
- AC-002-1: …

---

## 3. API Surface

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/v1/...` | session | … |
| GET | `/api/v1/public/...` | public | … |

### Error Codes

| Code | HTTP | Trigger |
|---|---|---|
| `VALIDATION_ERROR` | 422 | Body fails binding |
| `NOT_FOUND` | 404 | … |

---

## 4. Data Model Touch Points

| Table | Read | Write | Notes |
|---|:---:|:---:|---|
| `…` | ✓ | ✓ | … |

---

## 5. Cross-links

- ADR: `docs/architecture/adr/NNN-*.md`
- Sister: `docs/fr/survey/active/fr-*.md`

---

## 6. Contract (machine-readable)

> Drift-detector source. Schema: `docs/fr/_contract-schema.json`.

```yaml
fr_file: docs/fr/survey/active/fr-<feature>.md
covers:
  - FR-XXX-001
  - FR-XXX-002

endpoints:
  - id: AC-XXX-CREATE
    method: POST
    path: /api/v1/...
    auth: { mode: session }

db:
  writes:
    - table: <table>
      columns_declared: [id, ...]

cross_links:
  adr_refs: []
  sisters: []
```
