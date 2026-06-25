-- 003 — questions (fields of a form).
--
-- Predecessor audit (Rule 1): none affected.
-- Ordering: (form_id, position) is indexed but NOT unique — reorder rewrites
-- positions contiguously in a single transaction, and a non-unique index keeps
-- that operation collision-free without deferrable-constraint gymnastics.
-- Reads always ORDER BY position, created_at, id for a stable sequence.
-- metadata JSON holds per-type config (options, min/max, rating scale, ...).

CREATE TYPE question_type AS ENUM (
    'short_text', 'long_text', 'multiple_choice', 'checkboxes', 'dropdown',
    'rating', 'number', 'email', 'date', 'yes_no', 'statement'
);

CREATE TABLE questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id     UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    type        question_type NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL,
    required    BOOLEAN NOT NULL DEFAULT false,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX questions_form_pos_idx ON questions (form_id, position);
