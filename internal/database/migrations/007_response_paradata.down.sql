-- down for 007_response_paradata.
DROP INDEX IF EXISTS responses_in_progress_idx;
ALTER TABLE responses
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS last_seen_at,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS furthest_position;
