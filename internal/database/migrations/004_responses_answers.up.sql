-- 004 — responses + answers (anonymous submissions).
--
-- Predecessor audit (Rule 1): none affected.
-- Answers are NORMALIZED (one row per answered question) with a JSONB `value`
-- leaf for the type-varying payload (string | number | [option_ids] | bool).
-- UNIQUE(response_id, question_id) prevents a double-answer; the question index
-- backs per-question analytics (Phase 4) via GROUP BY question_id.

CREATE TABLE responses (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id      UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    completed    BOOLEAN NOT NULL DEFAULT false,
    meta         JSONB NOT NULL DEFAULT '{}',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX responses_form_idx ON responses (form_id, submitted_at DESC);

CREATE TABLE answers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    response_id UUID NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    value       JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (response_id, question_id)
);
CREATE INDEX answers_question_idx ON answers (question_id);
