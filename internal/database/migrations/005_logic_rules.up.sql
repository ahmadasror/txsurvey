-- 005 — logic rules (conditional branching / show-hide / early end).
--
-- Predecessor audit (Rule 1): none affected; additive table.
-- A rule is (source_question, operator, compare_value) -> (action, target).
-- Rules for a given source are evaluated by ascending priority; first matching
-- navigation action (jump_to / end_form) wins. show/hide rules adjust a target's
-- visibility. target_question_id is NULL only for end_form.

CREATE TYPE logic_operator AS ENUM (
    'eq', 'neq', 'gt', 'gte', 'lt', 'lte', 'contains', 'not_contains', 'is_empty', 'is_not_empty'
);
CREATE TYPE logic_action AS ENUM ('jump_to', 'show', 'hide', 'end_form');

CREATE TABLE logic_rules (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id            UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    source_question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    operator           logic_operator NOT NULL,
    compare_value      JSONB,
    action             logic_action NOT NULL,
    target_question_id UUID REFERENCES questions(id) ON DELETE CASCADE,
    priority           INTEGER NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX logic_rules_form_idx ON logic_rules (form_id);
CREATE INDEX logic_rules_source_idx ON logic_rules (source_question_id, priority);
