---
description: Lean SDD /spec — author/refresh an FR + machine-readable contract block for txsurvey
argument-hint: <feature-name>
---

Invoke the `fr-writer` agent to author the L3 functional requirements for the
survey feature `$ARGUMENTS` in this repo (txsurvey).

This is the **primary entry point** of the lean SDD workflow (see
`docs/sdd-workflow.md`). Code-first is fine — produce/refresh the FR + contract
block to pair with the code (same commit or immediate follow-up).

**What to produce**:
- `docs/fr/survey/active/fr-$ARGUMENTS.md` from the template `docs/templates/fr.md`:
  - FR ids + AC (one section per requirement, code paths referenced)
  - API surface table + error codes
  - data-model touch points
  - a `## Contract (machine-readable)` YAML block (schema: `docs/fr/_contract-schema.json`)
- An ADR in `docs/architecture/adr/NNN-*.md` (template `docs/templates/adr.md`) **only if** a non-obvious decision was made.

**Contract block rules** (lean):
- `endpoints[]` carry `{ id, method, path, auth: { mode: session|public } }`. Use real `/api/v1/...` paths with `{param}` placeholders; they must resolve in `internal/router/routes.go`.
- `db.writes[]` list the tables + columns the feature writes; they must exist in `internal/database/migrations/`.
- `covers[]` use ids matching `^FR-[A-Z0-9]+-[0-9]+$`.

**After writing**: run `make spec-validate` (schema gate) and `make spec-drift`
(advisory FR↔code check). Both should be clean for a money/critical path.

Heavy upstream artifacts (PRD, formal design, test scenarios) are **optional** —
pull them in only when the feature genuinely needs them.

Use the Agent tool with `subagent_type: fr-writer`, passing feature `$ARGUMENTS`
and this repo's conventions as context.
