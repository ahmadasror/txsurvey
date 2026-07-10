-- 006 — add the 'always' logic operator (unconditional jump).
--
-- Predecessor audit (Rule 1):
--   - 005: logic_operator ENUM(eq,neq,gt,gte,lt,lte,contains,not_contains,
--     is_empty,is_not_empty) — EXTENDED, not replaced. No ON CONFLICT / UNIQUE /
--     CHECK depends on the exact value set, so appending a value is safe.
--
-- Behaviour change:
--   - Creators can attach an *unconditional* jump: operator 'always' with
--     action 'jump_to' routes a question straight to a chosen later question with
--     no condition to reason about. The engines treat 'always' as a match
--     regardless of the answer (even when the source is unanswered).
--
-- Idempotency: Yes (IF NOT EXISTS). PG12+ permits ADD VALUE inside a txn as long
--   as the new value is not USED in the same txn — this migration only adds it.
-- Rollback: forward-only — dropping a single enum value needs a type rebuild +
--   column rewrite; see the .down.sql note.

ALTER TYPE logic_operator ADD VALUE IF NOT EXISTS 'always';
