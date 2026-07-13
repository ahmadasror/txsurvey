---
description: /cgeser — hand a one-line task to the right model tier. Opus classifies + dispatches, Sonnet implements, `make check` verifies, Opus reviews only the judgment layer. Warm-context + judgment-dense tasks stay full-Opus.
argument-hint: <task in one sentence>
---

You are the orchestrator. The user handed you a one-line task:

> `$ARGUMENTS`

**Turning that sentence into a proper dispatch is YOUR job — never ask the user to write a
detailed prompt.** Run the loop below, and spend Opus tokens only where they pay off (classify
+ judgment review); push the token-heavy implementation down to Sonnet. The repo's gate
(`make check`) is what makes this cheap — it verifies the mechanical invariants so you don't
have to re-read the repo to check them.

## 1. Classify `$ARGUMENTS`

- **Trivial** (single-file, mechanical: copy/typo, rename, dep bump, one obvious edit) →
  **do it yourself now**, run `make check`, report. STOP. Do NOT dispatch — the 3-stage
  overhead costs more than it saves on a small task.
- **Ambiguous** (scope unclear, needs a decision, or touches architecture/data-model
  non-obviously) → ask the user **1–2 sharp questions** (AskUserQuestion), or if it's a real
  feature, run `/spec <name>` first to pin the FR + contract. Do NOT dispatch blind. Re-classify
  after.
- **Clear + substantial** (multi-file but a known pattern the repo already teaches — see the
  CLAUDE.md recipes) → **dispatch to Sonnet** (step 2).

## 1b. Warm-context override — check BEFORE dispatching or /spec-ing

The classifier above judges task *shape*. Two levers it ignores can flip the answer to
**full-Opus in this session** — do the whole thing yourself, no dispatch, no Sonnet:

Take the full-Opus path when **both** hold:
- **Warm context** — the files this task touches are **already loaded/warm in this session**
  (you read or reasoned over them earlier in this conversation). Cached input bills at ~0.1×,
  so Opus's input here is *cheaper* than a fresh Sonnet subagent that must re-read the repo cold.
- **Judgment-dense / high cost-of-miss** — a Sonnet miss would be expensive: cross-cutting
  semantic changes, shared contracts other modules read, security, money/ledger, or correctness
  where "gate green" ≠ "no latent change" (e.g. a new row type quietly changing an existing
  count/metric).

Why: warm cache removes Opus's cost penalty, and one coherent judgment across design +
implementation avoids the **handoff seam** where a subtle cross-module impact gets missed (the
exact failure mode a cold Sonnet dispatch is prone to). This override beats both the `/spec`
and the dispatch routing for that specific case — it is the *only* case where a substantial
task stays full-Opus.

If only ONE holds — warm but mechanical, or judgment-dense but on a **cold** subsystem — do
NOT full-Opus: fall back to §1 (dispatch, or design-first then dispatch). A cold subsystem gives
Opus no cache edge; a mechanical warm task doesn't need Opus judgment.

## 2. Dispatch to Sonnet

Spawn an implementer with the Agent tool (`model: sonnet`). Expand `$ARGUMENTS` into the
dispatch below, filling specifics from your classification (which files/recipe; does it touch
an endpoint or table?):

> Kerjakan di repo txsurvey. Buat branch `feat/<slug>` dari branch aktif; commit konvensional
> kecil-kecil; JANGAN push ke main, JANGAN merge.
> Task: `<expanded task + kriteria selesai>`.
> - Kalau ini mengextend konsep yang ada (question type / logic operator), ikuti resep di
>   CLAUDE.md "Adding a question type or operator" — dan ingat logika kondisional WAJIB di
>   KEDUA engine (`make check` menegakkan lockstep-nya).
> - Kalau menambah/ubah endpoint atau tabel: update FR contract pemiliknya di `docs/fr/...`
>   di commit yang sama.
> - Acceptance: `make check` HARUS hijau sebelum lapor. Kalau tak bisa hijau, revert bagian
>   itu dan laporkan — jangan ship merah.
> - Lapor: file yang berubah, output `make check`, dan asumsi yang diambil.

For parallel/isolated work use `isolation: worktree` — but only from a **fresh session**, since
a worktree forks from the session-start commit (mid-session it won't contain recent commits).

## 3. Verify — trust the gate, don't re-verify mechanics

When Sonnet reports, confirm `make check` is green (run it if unsure). It already checked the
mechanical invariants — parity (both engines), coverage ratchet, route→FR, docs, tests. **Do
NOT re-read the whole repo to re-verify them; that's the expensive mistake this workflow
exists to avoid.** If `make check` is RED, bounce it back to the same Sonnet agent
(SendMessage) with the failure — once or twice, not an open loop.

## 4. Review the JUDGMENT layer only (short, cheap Opus pass)

Read the **diff**, not the exploration. Check what the gate can't: is this the right approach?
naming/UX sane? any owner-scoping / auth / security smell? **Does a new row/state quietly change
an existing count, metric, list, or export other code relies on?** Keep it brief. Then report to
the user: one-line what-changed, the branch name, `make check` status, and follow-ups (merge and
deploy are the user's call — see `docs/runbooks/deploy.md`).

## Cost discipline (the whole point)

- Trivial → do it directly, no dispatch.
- **Warm + judgment-dense → full-Opus in this session** (§1b) — cache makes it cheap, coherence
  avoids the miss.
- Otherwise trust `make check`; never hand-verify mechanics. Keep your Opus touches small
  (classify + judgment review). Sonnet absorbs the token-heavy implementation at its cheaper
  rate — that's why this is cheaper than Opus-does-everything for non-trivial, cold, mechanical
  work, and why you must NOT pipeline a trivial task.
