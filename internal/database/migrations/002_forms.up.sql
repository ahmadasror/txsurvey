-- 002 — forms (a creator's survey).
--
-- Predecessor audit (Rule 1): no ON CONFLICT depends on forms yet.
-- slug uniqueness is PARTIAL (live rows only) so a soft-deleted form frees its
-- slug for reuse. Behaviour: public runner resolves a form by slug among
-- non-deleted, published rows (Phase 3).

CREATE TYPE form_status AS ENUM ('draft', 'published', 'closed');

CREATE TABLE forms (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    slug         TEXT NOT NULL,
    status       form_status NOT NULL DEFAULT 'draft',
    settings     JSONB NOT NULL DEFAULT '{}',
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX forms_slug_unique ON forms (slug) WHERE deleted_at IS NULL;
CREATE INDEX forms_owner_idx ON forms (owner_id) WHERE deleted_at IS NULL;
