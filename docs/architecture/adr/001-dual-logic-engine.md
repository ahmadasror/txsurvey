# ADR-001 — Dual logic engine (Go authoritative + TS runner)

> **Status**: accepted
> **Date**: 2026-06-25

## Context

The conditional-logic feature (jump / show / hide / end) needs to drive runner
navigation in the browser (interactive UX) AND be enforced on the server at submit
time (a client can tamper with the payload). A single implementation can't live in
both places.

## Decision

Implement the evaluator **twice**, with identical semantics, and treat the Go side
as authoritative:

- `internal/service/logic_engine.go` — `reachablePath()` replays the path the
  submitted answers imply; submit validation enforces required-ness only on
  reachable questions and rejects answers to skipped questions.
- `frontend/src/lib/logicEngine.ts` — runner navigation only (jump/show/hide/end,
  progress over reachable path). Never trusted.

Rules: per-source first-match-by-priority for navigation; `show`/`hide` adjust
visibility; a visited-set + step cap guarantee termination even after a reorder
leaves a backward jump.

## Alternatives considered

- **Server-only (no client engine)** — every "next" is a round-trip; kills the
  smooth one-question-per-screen feel. Rejected.
- **Client-only (trust the payload)** — anyone can POST answers to skipped/required
  questions. Rejected (data integrity).
- **Compile one engine to both targets (GopherJS/WASM)** — heavy toolchain for a
  ~150-line pure function. Rejected as over-engineering at this scale.

## Consequences

- Positive: smooth UX + server is the source of truth.
- Negative: two implementations must be kept in lockstep.
- Mitigation: `logic_engine_test.go` is the spec; CLAUDE.md flags the lockstep
  requirement; the TS file header points back to the Go file.

## Related

- FR: `docs/fr/survey/active/fr-runner.md`, `docs/fr/survey/active/fr-forms.md`
