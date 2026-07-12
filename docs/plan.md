# Plan — raise Model-Readiness from 66 toward ~80+

**Owner intent:** execute this **unattended at 22:00 WIB** (limit-saving). Convert the
un-gated invariants the assessment flagged into red/green gates. Source of truth for the
baseline: [`docs/sonnet-readiness.md`](sonnet-readiness.md); method:
[`docs/model-readiness-kit.md`](model-readiness-kit.md).

The thesis (from the kit): a weak model follows an invariant when it's **encapsulated,
exemplified, or gated** — not when it's written in prose. Baseline 66 is capped by
Dimension 2 (enforcement = 52): the #1 invariant (dual logic engine lockstep) and test
coverage are **prose-only**. This plan gates them.

---

## Ground rules for the autonomous run (READ FIRST)

1. **Branch, don't touch main.** Create `feat/readiness-gates` from `main`. Commit in small
   conventional commits (`test:`, `feat:`, `chore:`, `docs:`). **Do NOT push to main.** At
   the end, either leave the branch or open a PR — do not merge.
2. **No outward-facing actions.** Do **NOT** deploy, do **NOT** rebuild the prod image, do
   **NOT** touch the live `txsurvey-app-1` container or the shared `txsurvey` dev DB.
3. **DB safety:** integration tests need a DB whose name contains `test`. Use a **throwaway**
   DB (e.g. `txsurvey_plan_test`) that you create and DROP within the run — never reuse or
   mutate the shared `txsurvey_test` beyond what the harness does. If you can't get Postgres,
   skip DB-backed steps and say so; unit-level work does not need a DB.
4. **Every task must end green.** Acceptance gate for the whole run:
   `make lint` ✓ · `make test` ✓ · `cd frontend && npm run build` ✓ · `make docs-check` ✓
   plus any new gate this plan adds. If a task can't go green, revert that task's commits and
   record it in the report rather than shipping red.
5. **No theater** (kit rule): if a gate can't be made to actually bite, skip it and say why —
   never add a vacuous test/check to move a number.
6. **Take assumptions on blockers, log them.** Don't stop for questions; pick the reasonable
   default, note it in the final report.
7. **Keep the two engines in lockstep** for any logic change (`internal/service/logic_engine.go`
   ↔ `frontend/src/lib/logicEngine.ts`) — this plan is partly *about* that invariant.

---

## Tasks (priority order — do 1–4 first; 5–7 if time remains)

### 1. Frontend unit-test harness (vitest) — enables everything below
- Add `vitest` (+ `@vitest/coverage-v8`) to `frontend` devDeps; add `"test": "vitest run"` and
  `"test:watch": "vitest"` to `frontend/package.json`. Minimal `vitest.config.ts` (jsdom not
  needed for pure `lib/` logic; node env is fine).
- Add one real smoke test importing `src/lib/logicEngine.ts` to prove the harness runs.
- **Accept:** `cd frontend && npm run test` runs and passes; `npm run build` still green.

### 2. Dual-engine parity gate — THE biggest win (Dim 2)
- Create a **shared fixture** of conditional-logic cases as JSON, consumed by *both* engines:
  `internal/service/testdata/logic_parity_cases.json` — array of
  `{ "answer": <json>, "operator": "...", "compareValue": <json>, "expected": true|false }`
  covering every operator (`eq neq gt gte lt lte contains not_contains is_empty is_not_empty
  always`) incl. edge cases (arrays for checkboxes, missing answer, type mismatch).
- **Go side:** a test in `internal/service/` that loads the JSON and asserts
  `conditionMatches(...)` equals `expected` for each row.
- **TS side:** a vitest that loads the *same* JSON file (via a relative import / fs read) and
  asserts the TS engine's `conditionMatches`/evaluation equals `expected` for each row.
- Because both read one fixture, a divergence between the engines makes one side go red.
- Seed the fixture from the existing `logic_engine_test.go` `TestConditionMatches` table so it
  starts complete. **Accept:** both Go and TS parity tests pass on the same JSON; deliberately
  breaking one engine (locally, then revert) makes its side fail.

### 3. Umbrella `make check` (Dim 3)
- New Makefile target `check` running, in order: `make lint`, `go test ./...`,
  `cd frontend && npm run test && npm run build`, `make docs-check`, and the coverage gate
  from task 4. One red/green a weak worker can't miss. Update `CLAUDE.md` Commands to mention it.
- **Accept:** `make check` passes clean on the branch and returns non-zero if any sub-gate fails.

### 4. Coverage ratchet (Dim 2)
- `scripts/cover_check.py` (or Go) storing a per-package floor in
  `scripts/coverage-baseline.json`; `make cover-check` fails if any tracked package drops below
  its floor; `make cover-update` re-blesses. **Baseline = today's numbers** (measure them in the
  run; today they were ~`service 34.8%`, `pkg/auth 84.6%`, `handler 2.1%` — do not demand more,
  only forbid regression). Fold `cover-check` into `make check`.
- **Accept:** `make cover-check` green at baseline; lowering a covered path (locally, revert)
  trips it red.

### 5. Make `spec-drift` enforcing (Dim 2)
- Promote route→FR from advisory (exit 0) to a hard gate: `scripts/spec_drift.py` exits non-zero
  on an undocumented route, with a **waiver file** (`docs/fr/_route_waivers.txt` or similar)
  listing any legacy routes so the baseline starts green. Add `make route-check` and fold into
  `make check`. Keep the advisory `spec-drift` target too if useful.
- **Accept:** `make route-check` green today; adding a fake unspecced route trips it red.

### 6. Runbooks + recovery docs (Dims 6, 7)
- `docs/runbooks/deploy.md` — linearize the deploy sequence from `CLAUDE.md` (gate → build →
  `docker compose -f docker-compose.prod.yml -p txsurvey up -d --build` → verify).
- `docs/runbooks/recovery.md` — symptom→cause→fix table. Seed it with the two real incidents
  from this exercise: (a) **duplicate migration number** under concurrent work → renumber to the
  next free number; (b) **shared `*_test` DB drifted** to a migration version whose file was
  reverted → non-destructively reset `schema_migrations.version` to the max committed migration
  (do NOT drop the shared DB). Plus the known gotchas (port 8080 occupied, `*_test` truncation
  guard, forward-only migrations, slug-lock).
- **Accept:** `make docs-check` still green; files cross-linked.

### 7. (If time) post-deploy smoke script — WRITE ONLY, do not run against prod
- `scripts/verify-deploy.sh` + `make verify-deploy`: container up → local `/health` → public
  edge `200`, each red line naming the likely culprit. **Do not execute it against live infra in
  this run** — just author it and reference it from `docs/runbooks/deploy.md`.

---

## Finalize (end of run)

1. Run the full acceptance gate: `make check` green.
2. **Update the scorecard** `docs/sonnet-readiness.md`: add a "Round 1 — gates" section; move
   the now-gated items in the Dim 2 table and gate inventory from ❌ to ✅; re-derive the score
   (expected Dim 2 ~52→~72, Dim 3 ~60→~72 → total ~66→~72–75). Be honest about what's still
   ungated (e.g. smoke gate authored-not-wired).
3. **Regenerate the infographic** to match: edit `docs/assessment-infographic.html` numbers +
   the gate ✓/✕ column, re-render to `docs/assessment-infographic.png` at 800×600 via
   `cd frontend && npx playwright screenshot --viewport-size=800,600 file://<abs path> out.png`,
   and copy to `frontend/public/assessment/infographic.png`. (Do NOT redeploy — the live site
   updates on the owner's next manual redeploy.)
4. Write a report to `docs/plan-report.md`: what landed, what was skipped and why, gate status,
   the new score, and the exact follow-ups left for the owner (notably: **redeploy** to publish
   the updated page/infographic, and review/merge the `feat/readiness-gates` branch).

## Explicitly out of scope for the autonomous run
- Any deploy / prod image rebuild / live-infra probe.
- Re-running the Sonnet spike subagents (expensive; the owner can re-measure interactively).
- Merging to main.
