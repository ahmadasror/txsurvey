# CODEX HANDOFF — txsurvey (Codex migration pilot)

> **READ THIS FIRST, then execute it.** You (Codex) are the pilot for migrating
> Ahmad's dev workflow **off Claude Code onto Codex CLI**. This repo, `txsurvey`,
> is the low-risk proving ground. This file is a **one-time mission brief**: run the
> Setup checklist, do the Exit Test, write the report. Do **not** delete this file
> until the pilot passes (§3). It is authored from Ahmad's Obsidian vault, not by you.

## 0. Why this exists

- Ahmad plans to **drop the Claude subscription → ChatGPT/Codex**. Coexistence
  (Claude + Codex both working) is only a *transition bridge*, not the goal.
- txsurvey is picked because it is the **non-canonical, lower-stakes** repo (the
  canonical ones are kucingcat/txcrm). If Codex misbehaves here, blast radius is small.
- **Pass condition (the whole point):** complete one full feature cycle on txsurvey
  **Codex-only, without ever opening Claude** (see §3).

## 1. Orient fast (what this repo is)

- **Go (Gin) API-first backend + React SPA.** Typeform-like survey SaaS, 5–10 creators.
- **Authoritative deep-dive = `CLAUDE.md`** (repo root, ~10 KB, tool-agnostic content).
  Read it fully before editing — it lists the dev commands and the invariants that
  *"will bite you"* (dual logic engine, forward-only migrations, cookie-session auth,
  slug lifecycle, hex→HSL theming, build-tagged SPA embedding).
- **Dev loop** (from `CLAUDE.md §Commands`): `make run` · `make check` (umbrella gate:
  lint+unit+coverage+routes+docs+FE) · `make test` (integration; DB name **must** contain
  `test`, harness TRUNCATEs) · `go test ./... -short` (unit only) · `make build` · `make lint`.
- **Spec-driven via leanSDD** (`~/code/leanSDD`, workflow in `docs/sdd-workflow.md`):
  FR → `docs/fr/survey/active/fr-<feature>.md` (ends in a `## Contract` YAML block),
  ADR → `docs/architecture/adr/NNN-*.md`, gates `make docs-check` / `make spec-drift`.

## 2. Setup checklist (do in order; commit after each step)

- [ ] **2.1 Shared source of truth.** `AGENTS.md` (interim bridge) already exists at root.
      Finalize it: make `AGENTS.md` the canonical project-instruction doc for Codex, and
      keep `CLAUDE.md` working for Claude during coexistence. Cleanest: `AGENTS.md` mirrors
      the tool-agnostic content of `CLAUDE.md`; if you consolidate, keep `AGENTS.md` **< 32 KiB**
      (Codex truncation limit) and leave `CLAUDE.md` a thin `@AGENTS.md`-style pointer so both
      tools read one source. **Verify Claude still loads CLAUDE.md afterward.**
- [ ] **2.2 Port the kept agents (only 2).** Ahmad curated the global fleet down to
      `security-reviewer` + `night-builder` (all others deleted — SDD flow is now leanSDD,
      the rest were unused duplicates). Translate these two from `~/.claude/agents/*.md`
      (YAML frontmatter) → `.codex/agents/*.toml`. Map the model alias; set `sandbox_mode`.
      Do **not** recreate any other legacy agent.
- [ ] **2.3 Custom commands → Codex.** `.claude/commands/{spec,cgeser}.md` are Claude slash
      commands. Re-express as Codex **skills** (`.codex/skills/<name>/SKILL.md`) or plain prompts.
      ⚠️ `CLAUDE.md` says *"/spec → fr-writer authors the FR"* — **stale**: `fr-writer` is deleted.
      FR authoring is now a **leanSDD** step; wire `/spec` to that, not to an agent.
- [ ] **2.4 MCP.** Mirror any MCP servers Ahmad uses into `.codex/config.toml [mcp_servers.*]`.
      (Cross-check his `~/.claude` / claude.ai connectors — Gmail/Calendar/Drive/GitHub.)
- [ ] **2.5 Cost tuning.** Set a cheap model on `security-reviewer` and the worker pool; a
      capable model on `night-builder`. Configure `[agents] max_threads` / `max_depth` to mirror
      the old "orchestrator + cheap subagents" pattern (token-optimize delegation).
- [ ] **2.6 Sanity.** Make a trivial Codex-driven edit and confirm `make check` stays green.

## 3. Exit test — the pass/fail

Pick **one small, real** change on txsurvey and run the **full leanSDD cycle, Codex-only**:

1. FR + `## Contract` block in `docs/fr/survey/active/fr-<feature>.md`.
2. ADR if the decision is non-obvious (`docs/architecture/adr/`).
3. Implement it — respecting the invariants (esp. the **dual logic engine lockstep** and
   **forward-only, next-number-only migrations**; see `CLAUDE.md`).
4. Tests green: `go test ./...` **and** `cd frontend && npm run build`.
5. Gates green: `make check` **and** `make docs-check` (`make spec-drift` clean).
6. One conventional commit, single concern.

**Pass = all 6 done without touching Claude, gates green.**

Good small candidates (self-contained, exercise the extension recipes in `CLAUDE.md`):
- Add a new **logic operator** (e.g. `is_not_empty`) — touches both engines + builder + a migration.
- Add a **theme preset** in `frontend/src/lib/themes.ts` (+ mirror `:root` in `index.css`).
- A small **runner/results** UI fix. Avoid anything touching deployment or auth on the first run.

## 4. Report back

Write findings to `docs/codex-pilot-report.md`:
- What ported **1:1** vs what needed **manual translation**.
- Any Claude primitive with **no Codex equivalent** (real gaps — e.g. auto-memory).
- **Quality delta** vs Claude on the 2 kept agents (subjective is fine).
- **Cost/limits**: which ChatGPT tier (Plus $20 / Pro $100) the fleet actually needs.
- **Recommendation:** cut Claude, or keep permanent coexistence (Codex routine + Claude for
  corebanking/financial-critical work)?

## 5. Guardrails (do not violate)

- **Do NOT break Claude** during the pilot — keep `CLAUDE.md` valid and loadable.
- **Do NOT deploy / touch production.** There is no CD; deploy is a manual
  `docker compose -f docker-compose.prod.yml` on the VPS (`CLAUDE.md §Deployment`). Live at
  brainzap.net/txsurvey. Never commit `deploy/production.env`.
- **Integration tests TRUNCATE** — only ever point `DATABASE_URL` at a `*_test` DB.
- **Migrations are forward-only**, next number only. Don't back-fill a lower number.
- Conventional commits, one concern per commit; don't force-push `main`.
