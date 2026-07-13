-- 007 — response paradata (progress-only lifecycle capture).
--
-- Predecessor audit (Rule 1):
--   - V004 responses: `completed BOOLEAN NOT NULL DEFAULT false`, no ON CONFLICT
--     / unique to preserve. Submit has ALWAYS written completed=true, so
--     completed=false has never occurred — we REUSE `completed` as the lifecycle
--     flag (in_progress = false) rather than add a redundant status enum.
--
-- Behaviour change:
--   - A new create-on-start + update-on-progress path writes in_progress rows
--     (completed=false). Existing owner-facing surfaces (results list, response
--     count, completion-rate denominator, CSV, analytics) are scoped to
--     `completed` in the repo layer, so in_progress rows are INERT paradata —
--     they change none of those numbers until a future funnel view queries them.
--
-- Idempotency: Yes (ADD COLUMN IF NOT EXISTS + guarded backfill).
-- Rollback: 007_response_paradata.down.sql drops the four columns.

ALTER TABLE responses
    ADD COLUMN IF NOT EXISTS started_at        TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_seen_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS furthest_position INTEGER NOT NULL DEFAULT 0;

-- Backfill: every existing row is a completed submission — stamp completed_at
-- from submitted_at so the column is meaningful for historical rows.
UPDATE responses SET completed_at = submitted_at WHERE completed AND completed_at IS NULL;

-- Live in-progress sessions are the funnel's working set; a partial index keeps
-- them cheap to scan without touching the (large) completed set.
CREATE INDEX IF NOT EXISTS responses_in_progress_idx
    ON responses (form_id, last_seen_at DESC) WHERE NOT completed;
