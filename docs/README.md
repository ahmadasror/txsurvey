# txsurvey — Documentation Map

Lean SDD docs. Start at **[sdd-workflow.md](sdd-workflow.md)** for the process.

## Where is X?

| I want to… | Go to |
|---|---|
| Understand the dev workflow | [`sdd-workflow.md`](sdd-workflow.md) |
| See what a feature must satisfy (the contract) | `fr/survey/active/fr-*.md` |
| Know why a decision was made | `architecture/adr/` (index: `architecture/adr/README.md`) |
| Spec a new feature | `/spec survey` (fr-writer agent) → writes an FR here |
| Validate FR contract blocks | `make spec-validate` |
| Check FR↔code drift (advisory) | `make spec-drift` |
| **Check everything is aligned** | **`make docs-check`** (schema + cross-links + ADR index + drift) |
| Build/run/deploy the app | repo root `README.md` and `CLAUDE.md` |

## Layout

```
docs/
  sdd-workflow.md            # the lean process
  README.md                  # this map
  fr/
    _contract-schema.json    # contract block schema
    survey/{active,done,todo}/fr-*.md
  architecture/
    adr/README.md            # decision log
    adr/NNN-*.md
  templates/{fr.md,adr.md}
```

## Current FRs (survey)

- `fr/survey/active/fr-auth.md` — Google sign-in + session
- `fr/survey/active/fr-forms.md` — forms + questions + logic-rule CRUD + publish
- `fr/survey/active/fr-runner.md` — public runner: fetch + logic-aware submit
- `fr/survey/active/fr-results.md` — responses, analytics, CSV export

These were backfilled to seed traceability for the already-built code (code-first is the norm here).
