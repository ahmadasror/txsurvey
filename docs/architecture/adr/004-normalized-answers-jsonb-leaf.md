# ADR-004 — Normalized answers table with a JSONB value leaf

> **Status**: accepted
> **Date**: 2026-06-25

## Context

Survey answers are heterogeneous by question type (string, number, boolean, array
of option ids, rating). They must be queryable for per-question analytics and CSV
export (one column per question), and integrity-checked (one answer per question
per response).

## Decision

Store answers **normalized** — one `answers` row per answered question — with a
single JSONB `value` leaf for the type-varying payload:

```
answers(id, response_id → responses, question_id → questions, value JSONB,
        created_at, UNIQUE(response_id, question_id))
```

Analytics are aggregated in Go (`ResultsService.computeSummary`) over rows, not via
JSONB SQL. Indexes: `answers(question_id)` for per-question rollups.

## Alternatives considered

- **One JSONB blob per response** (`responses.answers = {qid: value}`) — per-question
  `GROUP BY` and one-column-per-question CSV become fragile JSONB extraction across a
  heterogeneous map; no FK on question_id. Rejected.
- **N typed columns / table-per-type** — schema churn on every new question type;
  joins explode. Rejected.

## Consequences

- Positive: clean `GROUP BY question_id`, FK integrity, `UNIQUE` guard, flexible leaf.
- Negative: a submission writes N rows (fine at this scale — done in one tx).

## Related

- FR: `docs/fr/survey/active/fr-runner.md`, `docs/fr/survey/active/fr-results.md`
