# Model-Readiness Kit — a portable method + rubric

> **Copy this one file into any repo.** It is stack-agnostic. It tells you whether your
> codebase is ready to hand to a *lower-capability worker* — a cheaper/smaller model, a
> future agent, or a junior developer — and gives you a repeatable loop to raise that
> readiness. A worked, filled-in example lives at `docs/sonnet-readiness.md` (this repo's
> own assessment) — use it as a reference for what "done" looks like per dimension.

## The one-line thesis

**Encode quality as machine-checkable gates (red/green), not prose a worker must
remember.** A smaller model reliably runs `test`/`lint`/`build` and reacts to red — it
does *not* reliably read, retain, and apply rules scattered across a long doc.

## The calibrated insight (why this works)

Observed across real runs (spawn a cheaper model on real tasks, no hints, then check its
diff): **a weak worker follows an invariant when it is (a) encapsulated in a component it
can't bypass, or (b) present as a working local exemplar next to the edit. It skips an
invariant that is "first-of-its-kind here" with no adjacent exemplar and no gate.**

Design consequence — for every invariant that matters, do ONE of:
1. **Encapsulate** it (a helper/component the worker must go through), or
2. **Exemplify** it (a working local example beside where the next edit happens), or
3. **Gate** it (a test/check that goes red if violated).

Prose in a big onboarding doc is the *weakest* lever. Use docs for *orientation*
(where things live), not for *enforcing correctness*.

## The rubric (7 dimensions — score each 0–100)

| # | Dimension (weight) | The question a weak worker fails on |
|---|---|---|
| 1 | Invariant encapsulation & local exemplars (25) | Is each invariant a component you go through, or a working example beside the edit — not a rule to recall? |
| 2 | **Machine-checkable enforcement (gates)** (25) | Are the important invariants red/green (test/lint/CI), or just written down? |
| 3 | Self-verification affordances (15) | Can the worker verify its own change with ONE command, deterministically? |
| 4 | Definition-of-Done clarity (10) | Is "done" an explicit checklist pointing at commands, or implicit? |
| 5 | Navigation / discoverability (10) | From the entry doc, is "where does X live" ≤2 hops? Do names match features? |
| 6 | Task runbooks (deploy/ship/release) (10) | Are recurring ops linear step-by-step, or ordering-sensitive prose? |
| 7 | Failure-recovery docs (5) | Are common failures → fixes written down (ideally embedded in a gate's error message)? |

Weights are a default — retune for your context. Total = weighted average.
Grade read: **<60** = worker-dependent (good bones, no enforcement) · **~85+** = ready
(invariants are red/green, not remembered).

## The loop (one evidence-backed round at a time)

1. **Spike** — spawn 2–3 cheaper-model agents on *real, representative* tasks, in clean
   worktrees, **user-voice prompt only, zero invariant hints**. Ask each to report its
   decisions + "what did you reuse vs. write new, and what did you have to guess".
2. **Verify** — you (or a stronger model) check each diff against the real invariants.
   *Trust nothing self-reported.* Where the worker guessed or reinvented = a gap.
3. **Map** — a miss *with a local exemplar present* = an encapsulation gap; a miss *with
   no gate* = an enforcement gap; a "couldn't find it" = a navigation gap.
4. **Close** — prefer a **gate** over a doc. Encapsulate/exemplify where a gate can't.
5. **Re-measure** — re-run; the score moves only where you added real enforcement.

### Two rules that keep it honest
- **No theater.** If a package/surface has no meaningful test, *skip it and say why* —
  never add a vacuous test to move a number. A gate you can't make bite isn't a gate.
- **Ratchet, don't demand.** Set every baseline to *today's* state, so gates catch
  *regressions* without a giant retroactive backlog. Raise floors over time.

## Gate patterns (stack-agnostic — adapt the tool, keep the shape)

| You want to guarantee… | Gate shape | Adapt with |
|---|---|---|
| New behaviour is tested | **coverage ratchet** — fail if a tracked unit's coverage drops below its recorded floor | any coverage tool; store a baseline JSON, `--update` to re-bless |
| New endpoint is documented | **route ↔ spec** — fail if a route has no spec/contract entry (waiver baseline for legacy) | parse your router + your API spec |
| No private data leaks to an external surface | **whitelist test** — feed a full entity with sentinel private fields; assert the outgoing payload's keys ⊆ allowed set | any serializer boundary to a 3rd party |
| A change is "done" | **umbrella command** — one target running build+lint+test+coverage+spec | wrap your gates in one `make check`-style entry |
| DB/integration behaviour works | **throwaway-env harness** — spin up disposable infra, run tagged tests, tear down | docker + your migration path |
| The deploy actually worked | **post-deploy smoke gate** — process up → local health (+ its deps) → public edge; each red line names the culprit | your process manager + health endpoint + edge |

## Anti-patterns (learned the hard way)

- **Chasing the number.** The score is a *predictor*, not the goal. Each round must be a
  real gate/exemplar, documented, or the number is a lie.
- **Docs as enforcement.** Adding a paragraph to the onboarding wall does *not* change a
  weak worker's behaviour. A red test does. (Verified: identical tasks — the worker
  added the missing test/spec only once a *gate* would go red.)
- **Trusting a green you never run.** A tier of tests behind a flag that no gate runs
  *will* bit-rot silently and give false confidence. **Running the gate is what finds
  the rot.** Wire every tier into a one-command gate.
- **Silent scope caps.** If a gate only covers part of the repo (e.g. tested packages,
  one surface), *say so* — partial coverage reads as "covered everything" when it isn't.

## Blank scorecard (copy me)

```
# <repo> — Model-Readiness Assessment   (score: __ → __ / 100)

| # | Dimension (weight)                 | Before | After | Evidence |
|---|------------------------------------|-------:|------:|----------|
| 1 | Encapsulation & exemplars (25)     |        |       |          |
| 2 | Machine-checkable enforcement (25) |        |       |          |
| 3 | Self-verification (15)             |        |       |          |
| 4 | Definition-of-Done (10)            |        |       |          |
| 5 | Navigation (10)                    |        |       |          |
| 6 | Task runbooks (10)                 |        |       |          |
| 7 | Failure-recovery (5)               |        |       |          |

Gates added this round: ...
No-theater skips (and why): ...
Honest remaining scope: ...
```

## Making the score deterministic (so two evaluators agree)

A holistic 0–100 per dimension is **not reproducible**: two careful evaluators land
10–17 points apart per dimension, and a matching *total* is just compensating averaging,
not agreement. If you need a number you can trust across runs, repos, and evaluators,
turn each dimension into **binary, outcome-phrased criteria** and let a script sum them.

**The portability split (this is the whole trick):**

| Layer | What | Where |
|---|---|---|
| **Generic** | the outcome-phrased criteria (+ points) and the scorer engine | this kit + `scripts/readiness_score.py` — copy to any repo |
| **Per-repo** | how each criterion is verified: a `probe` (shell one-liner, exit 0 = met) or a `manual` met/not-met | `readiness.yaml` in the repo root — the only stack-specific file |

Criteria name **outcomes, never tools** — "the top invariant has a check that fails on
violation," not "a Go test named parity." So the same criterion is answered by
`grep … ci.yml` in one repo and `grep … Jenkinsfile` in another; the engine and criteria
don't change. Points across all criteria sum to 100; each dimension's criteria sum to its
weight. Score = Σ met points — a pure function of the probe outcomes, so **same repo state →
identical output** (the engine emits no timestamps/randomness; two runs or two agents match).

**Run it:** `python3 scripts/readiness_score.py readiness.yaml` (or `make readiness-score`).
It prints each criterion `[x]/[ ]`, per-dimension subtotals, the total, and — crucially — the
**determinism boundary**: how many points are `probe` (machine-verified, reproducible) vs
`manual` (the evaluator-dependent residue). Push the manual fraction down over time by
mechanizing fuzzy criteria into probes.

**Four honest limits — state them whenever you report a deterministic score:**
1. **Judgment doesn't vanish; it relocates** to the one-time design of the criteria + weights.
   That's a shared, reviewable decision, not re-litigated every evaluation.
2. **Fuzzy dimensions keep a residue.** Navigation ("≤2 hops"), encapsulation quality — leave
   these `manual` and report them as the variance surface, not as precision.
3. **Deterministic ≠ valid.** A binary rubric can miss what a holistic read catches, and can be
   gamed (a probe that always passes is theater — keep the probe output as the audit trail).
   Always pair the number with one narrative read.
4. **Binary tends to read higher than holistic** (existence earns full points), so add criteria
   that split *existence* from *strength* where it matters — e.g. "gate exists" AND "gate runs
   in CI" as two lines, so local-only enforcement is docked deterministically.

A worked instance — the criteria, probes, and a filled `readiness.yaml` — ships alongside this
kit in the txsurvey repo (`readiness.yaml` + `scripts/readiness_score.py`).

---

*This kit is itself just orientation. The value is the gates you build from it — and, if you
want a number you can defend, a score that is computed from those gates rather than felt.*
