# txsurvey — Model-Readiness Assessment (baseline)

> This is txsurvey's own filled-in scorecard. The portable, stack-agnostic method +
> rubric it applies lives in [`docs/model-readiness-kit.md`](model-readiness-kit.md) —
> read that to understand what each dimension means and how to run another round.

**Question it answers:** can a *lower-grade model* (Sonnet) pick up this repo and produce
Opus-quality changes without a human narrating the invariants? "Ready" has an operational
definition: *a Sonnet agent, given only the repo + docs, completes a representative task
correctly and to house style.* The score is a cheap predictor of that outcome — treat the
**per-dimension evidence** as the map, the total as a compass.

This baseline was calibrated by **2 real unaided Sonnet subagent runs** (clean checkout,
plain user-voice task, zero invariant hints) plus a direct audit of the gates/docs/CI.
Method + findings below. Re-runnable any time.

---

## Score: **66 → 78 / 100** (Round 1 — gates landed)

> Baseline was 66. **Round 1** (an unattended run executing [`docs/plan.md`](plan.md))
> converted the un-gated invariants into red/green — see the **Round 1** section below the
> table. Per-dimension scores show `baseline → after`.

| # | Dimension (weight) | Score | Why (evidence) |
|---|---|---:|---|
| 1 | **Invariant encapsulation & local exemplars** (25) | 82 | **Strong — measured, not guessed.** 2/2 unaided runs followed the encapsulated/exemplar invariants perfectly *without being told*: run A wired a new logic operator (`starts_with`) across **both** engines + model + builder + migration + FR; run B wired a new question type (`phone`) across the full chain. Both picked the *correct* structural exemplar (`contains` for the operator; `email`, not the PM-suggested `rating`, for phone) and matched house style (Indonesian labels, migration-header template). Capped at 82 because the conditional-logic engine is *duplicated* (`logic_engine.go` ↔ `logicEngine.ts`) — an anti-encapsulation that held only because a doc + adjacent exemplar reminded the model, not because the design makes divergence impossible. |
| 2 | **Machine-checkable enforcement (gates)** (25) | 52 → **75** | Was **the main gap**; Round 1 closed most of it. Baseline had CI (lint, `go test -race`, FE build, Playwright E2E) + `spec-validate` + `docs-check`, but the #1 invariant (dual engine) and coverage were prose-only. **R1 landed:** the dual-engine **parity gate** (shared fixture run through *both* engines — proven to bite), a **coverage ratchet** (`make cover-check`), an **enforcing route→FR** gate (`make route-check`, was advisory), the **vitest** harness (was zero FE tests), and an umbrella **`make check`**. Still ungated (why not 90): `verify-deploy` is authored-not-wired (probes live infra), coverage floors are low (regression-only, not high-absolute), E2E is one path. |
| 3 | **Self-verification affordances** (15) | 60 → **78** | **R1:** one umbrella **`make check`** now runs lint + unit tests + coverage + routes + docs + vitest + FE build, DB-free and deterministic — a worker verifies with a single command instead of four. Coverage is surfaced via `cover-check`. Cap: the DB integration tier still needs a manually-supplied `*_test` `DATABASE_URL` (not folded into `make check`). |
| 4 | **Definition-of-done clarity** (10) | 62 → **68** | **R1:** "done" now points at a *command* (`make check` green) rather than only prose, and CLAUDE.md's Commands list it. Still mostly prose for the FR/ADR artifacts, hence a modest bump. |
| 5 | **Navigation / discoverability** (10) | 88 | **Strongest dimension — measured.** Both unaided runs *found* the right files fast, picked the correct analog, and **reused rather than reinvented**. Names map to features (`features/builder`, `features/runner`, `logic_engine.go` ↔ `logicEngine.ts`); CLAUDE.md's "things that will bite you" + the add-a-type/operator map keep "where does X live" ≤1 hop. (Unchanged in R1.) |
| 6 | **Task runbooks** (deploy/ship) (10) | 60 → **76** | **R1:** `docs/runbooks/deploy.md` linearizes the deploy sequence (gate → build → verify) and `scripts/verify-deploy.sh` + `make verify-deploy` add a 3-layer post-deploy smoke check, each red line naming the culprit. Cap: `verify-deploy` is authored, not auto-run (it probes live infra + the public edge). |
| 7 | **Failure-recovery docs** (5) | 50 → **76** | **R1:** `docs/runbooks/recovery.md` — a symptom→cause→fix table seeded with the two *real* incidents this exercise hit (duplicate migration number; shared `*_test` DB drifted to a reverted migration version → non-destructive `schema_migrations` reset), plus the known gotchas. |

**Weighted total: 66.0 → 77.95 ≈ 78.**

Grade read: **66 → 78 = "B, Sonnet-ready and now mostly gated."** The baseline's biggest
risk — the dual logic engine diverging silently — is now **red/green** (a shared fixture
through both engines). What still caps the score: coverage floors are low (the ratchet
forbids regression but doesn't demand high absolute coverage — deliberate), the deploy smoke
gate is authored-not-wired, and Dim 1's duplication remains a *design* smell a gate mitigates
but doesn't remove.

---

## Round 1 — gates (autonomous run from `docs/plan.md`)

An unattended run implemented the plan's levers on branch `feat/readiness-gates`, each gate
proven to actually bite, whole run green (`make check` + full `go test ./...` incl.
integration). What landed:

1. **Frontend vitest harness** — there were zero FE unit tests; now `npm --prefix frontend
   run test` runs, covering the otherwise-untested TS logic engine.
2. **Dual-engine parity gate** (the biggest win) — one shared fixture
   `internal/service/testdata/logic_parity_cases.json` is run through **both**
   `logic_engine.go` (Go test) and `logicEngine.ts` (vitest). Divergence → one side red.
   Proven: breaking the TS substring match failed the TS parity suite.
3. **Umbrella `make check`** — lint + go unit tests + coverage ratchet + route-check +
   docs-check + vitest + FE build, DB-free, one red/green.
4. **Coverage ratchet** (`make cover-check` / `cover-update`) — floors = today (service 35.1%,
   auth 84.6%, handler 2.1%); fails on regression. Proven to bite.
5. **Enforcing route→FR** (`make route-check`) — promoted from advisory; one legacy route
   waivered in `docs/fr/_route_waivers.txt`. Proven to bite on a new unspecced route.
6. **Runbooks + recovery** — `docs/runbooks/{deploy,recovery}.md`.
7. **Post-deploy smoke** — `scripts/verify-deploy.sh` + `make verify-deploy` (authored; not
   wired into `make check` because it probes live infra).

Honest remaining scope (next round): raise DB-bound coverage floors via a counted integration
run; fold a throwaway-Postgres `make test-integration` into a CI stage; consider collapsing the
two logic engines to one source of truth (the parity gate makes that refactor safe). The score
moved only where a real gate landed — no number-chasing.

---

## What the spikes actually showed

**Run A — new logic operator `starts_with` (the dual-engine lockstep test).** An unaided
Sonnet, no hints, implemented it across **both** engines in lockstep (`logic_engine.go`
`case model.OpStartsWith` + `logicEngine.ts` `case "starts_with"`), the model const +
`ValidLogicOperator`, the `LogicEditor` operator list, the `LogicOperator` type union,
migration `008` (correctly renumbered after colliding with run B's `007` — a real judgment
call), Go test rows, **and** synced `fr-forms.md` (AC-003-4) unprompted. Verified against a
live Postgres running all migrations. Expert-level.

**Run B — new question type `phone`.** Full chain hit correctly: migration `007`,
`model.QuestionType` const + `ValidQuestionType`, `validateAnswer` case (strip-then-count
regex), runner `QuestionScreen`, `questionTypes.ts`, type union — **plus unit tests it
added on its own** by extending the adjacent table-driven exemplar. Chose `email` over the
PM's suggested `rating` as the structural match. Correctly judged that no FR needed updating
(the contract is route-based, not type-based).

**The load-bearing conclusion:** the invariants held because they are **encapsulated or
exemplified** (a `contains` twin sitting in both engines; a `validateAnswer` table right
there to copy) and because **CLAUDE.md's checklist told the model to touch both engines** —
*not* because a gate would have caught a miss. That is precisely the kit's thesis: this repo
raises the floor with docs + exemplars, and the way to raise it further is to convert the
un-gated invariants into red/green.

> Method note: these two spikes were throwaway *measurements* — their code (`phone`,
> `starts_with`) was reverted after scoring; only this scorecard was kept. Re-running them is
> how you re-measure after a lever lands.

---

## Next levers (highest ROI first) — for a future improvement round

1. **Dual-engine parity gate (dim 2, biggest single win).** A test that runs a shared table
   of `(answer, operator, compareValue) → expected` through *both* `conditionMatches`
   implementations and fails on any divergence. Options: a Go golden file the TS side also
   consumes via a tiny node runner in CI, or port `logic_engine_test.go`'s table into a
   vitest that imports `logicEngine.ts`. Turns ADR-001's prose invariant into red/green.
2. **Stand up frontend unit tests (dim 2/3).** There is no vitest at all today; add one and
   the parity test above has a home. Immediately covers the otherwise-untested TS engine.
3. **Umbrella `make check` (dim 3).** One target wrapping lint + test + FE build + docs-check
   (+ the parity + coverage gates below) — one red/green a weak worker can't miss.
4. **Coverage ratchet (dim 2).** Baseline = today's numbers (service ~35%, auth ~85%,
   handler ~2%), fail on regression. Neither spike was *forced* to add tests; this forces it.
5. **Make `spec-drift` enforcing (dim 2).** Promote route→FR from advisory (exit 0) to a hard
   gate with a waiver baseline for anything legacy.
6. **`docs/runbooks/` + post-deploy smoke gate + `recovery.md` (dims 6, 7).** Linearize the
   deploy prose; add a `verify-deploy` that checks container-up → `/health` → public edge;
   write the symptom→fix table (starting with the migration-collision run A hit).

Re-run the two spikes (or new ones) after each lever lands and diff the score — per the kit,
the number moves only where you added real enforcement.

---

## Method (reproducible, not vibes)

1. **Spike** — spawn unaided Sonnet agents on real, representative tasks, plain user-voice,
   **zero invariant hints**; ask each for decisions / reused-vs-new / what it had to guess.
2. **Verify** — Opus checks each *diff* against the invariant yardstick (below). Trust nothing
   self-reported. (Both spikes self-reported honestly here, but the diffs were still read.)
3. **Map** — a miss *with a local exemplar present* = encapsulation gap; a miss *with no gate*
   = enforcement gap; a "couldn't find it" = navigation gap.
4. **Improve** — prefer a machine gate over a doc for every real invariant.
5. **Re-score** — re-run and diff.

**Invariant yardstick (txsurvey):** conditional-logic change lands in **both** engines +
both test suites · `owner_id = $ AND deleted_at IS NULL` on every form-scoped query ·
jumps forward-only · submit validation stays logic-aware (required only on reachable, reject
answers to skipped) · new question type/operator walks the full CLAUDE.md chain · new
migration takes the next number, forward-only · owning FR contract updated in the same change
· `make lint` + `make test` + FE build + `make docs-check` green.

**Calibrated rule of thumb:** *a weak model follows an invariant when it is encapsulated in a
component it can't bypass, OR present as a working local exemplar next to the edit; it skips
an invariant that is "first-of-its-kind" here with no adjacent exemplar and no gate.* Design
for the model you're handing to: encapsulate, exemplify, or gate.

---

## Gate inventory (today)

| Gate | Command | Fails when | Status |
|---|---|---|---|
| Lint | `make lint` | `go vet` error or non-gofmt'd Go | ✅ enforced (CI) |
| Tests (+race, +DB) | `make test` / CI | any Go test fails | ✅ enforced (CI) |
| FE typecheck + build | `npm run build` | `tsc --noEmit` or vite build error | ✅ enforced (CI) |
| E2E create-survey | Playwright (CI) | the boot→create→submit happy path breaks | ✅ enforced (CI) |
| FR contract schema | `make spec-validate` | an FR contract block violates the schema | ✅ enforced |
| Docs coherence | `make docs-check` | broken cross-link / ADR-index incoherence | ✅ enforced |
| Route → FR (advisory) | `make spec-drift` | a route has no FR endpoint | ℹ️ advisory overview |
| **Route → FR (enforcing)** | `make route-check` | a route has no FR endpoint and no waiver | ✅ **enforced (R1)** |
| **Dual-engine parity** | `go test ./internal/service/ -run Parity` + `npm run test` | the two logic engines diverge on the shared fixture | ✅ **enforced (R1)** |
| **Coverage ratchet** | `make cover-check` | a tracked package's coverage regresses below its floor | ✅ **enforced (R1)** |
| **Frontend unit** | `make fe-test` / `npm run test` | a vitest (incl. parity) fails | ✅ **enforced (R1)** |
| **Umbrella** | `make check` | any of lint/test/coverage/routes/docs/FE | ✅ **enforced (R1)** |
| **Post-deploy smoke** | `make verify-deploy` | container down / local `/health` / public edge fails | 🟡 **authored (R1), run manually** |
