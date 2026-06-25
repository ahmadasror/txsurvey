# Lean SDD — txsurvey

Adapted from the txhcs "SDD Lean Mode". The heavyweight 7-layer / 8-agent
pipeline is intentionally **not** here — this is the solo-dev lean cut.

## Mantra

**Document as contract · code as truth · tests as the drift detector.**

## The 3 mandatory artifacts (per non-trivial feature)

1. **FR with a machine-readable contract block** — `docs/fr/survey/<partition>/fr-<feature>.md`,
   ending in a `## Contract (machine-readable)` YAML block (schema: `docs/fr/_contract-schema.json`).
2. **ADR** for any non-obvious decision — `docs/architecture/adr/NNN-<title>.md`.
3. **Tests green** — `go test ./...` (incl. the integration E2E) + `cd frontend && npm run build`.

That's it. PRD, formal architecture docs, API specs, and exhaustive test-scenario
docs are **on-demand**, not gated.

## Authoring direction: code-first is fine

Write the code, then land the **FR + contract block in the same commit (or the
immediate follow-up)**. Drift verifies *that pairing* on the API surface; it does
**not** demand spec-before-code. (Every phase in this repo's git history was built
code-first; the FRs in `docs/fr/survey/active/` were backfilled to seed traceability.)

## Two commands

- **`/spec survey`** — invoke the `fr-writer` agent to produce/refresh an FR + contract block
  (and an ADR if a decision is non-obvious). Primary entry point.
- **`make spec-drift`** — advisory FR↔code check: every endpoint declared in an FR
  contract block must exist in `internal/router/routes.go`. Exit 0 (advisory).

Plus the gates:

- **`make spec-validate`** — every FR contract block must be schema-valid (hard gate).
- **`make docs-check`** — one-shot "are the docs aligned?": runs the schema gate, a
  coherence sentinel (FR `fr_file` matches its path; `adr_refs`/`sisters` resolve; the
  ADR index lists exactly the ADR files that exist; warns on FRs missing from the docs
  README and on ADRs no FR references), then the advisory drift check. Run this before
  committing a docs change. `make docs-status` lists FRs by partition.

## Artifact layout

```
docs/
  sdd-workflow.md              # this file
  README.md                    # documentation map
  fr/
    _contract-schema.json      # contract block schema (the L3 source of truth)
    survey/
      active/                  # ready / in-flight FRs  (drift strict)
      done/                    # shipped + stable        (drift informational)
      todo/                    # drafted, not ready      (drift warn-only)
  architecture/
    adr/
      README.md                # decision log index
      NNN-<title>.md           # one decision per file
  templates/
    fr.md                      # FR template (with contract block)
    adr.md                     # ADR template
```

### Partitions (lean compaction)

- **active/** — the FR is the contract for live code. Endpoint/db drift here matters.
- **done/** — shipped and stable; kept for traceability, low-priority on drift.
- **todo/** — a stub or draft; not yet a contract. Move to `active/` when ready.

Move a file between partitions with a plain `git mv` — no tooling required.

## Definition of done (non-trivial feature)

- [ ] `go test ./...` green (integration E2E covers the happy path)
- [ ] FR exists in `docs/fr/survey/active/` with a `## Contract` block
- [ ] `make spec-validate` passes (schema-valid contract block)
- [ ] `make spec-drift` shows the new endpoints resolved (advisory)
- [ ] ADR written if a non-obvious decision was made

## When to reach for the heavy artifacts (on-demand only)

| Moment | Artifact | How |
|---|---|---|
| Fuzzy requirements / need user-flow discovery | PRD | `docs/prd/...` (free-form) or the `requirement-gatherer` agent |
| A real architecture fork (storage, topology, security) | ADR | `docs/templates/adr.md` |
| Many test cases worth enumerating before coding | Test scenarios | `docs/test-scenarios/...` or the `tester-explorer` agent |

If a feature is trivial (a copy tweak, a refactor, a dependency bump), skip the FR —
the lean rule only applies to *non-trivial* features that change behavior or the API surface.
