package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// FormRepo persists forms. Every owner-scoped query carries
// `owner_id = $ AND deleted_at IS NULL` so one creator can never read or mutate
// another's forms, and soft-deleted forms are invisible.
type FormRepo struct {
	pool *pgxpool.Pool
}

func NewFormRepo(pool *pgxpool.Pool) *FormRepo {
	return &FormRepo{pool: pool}
}

const formCols = `id, owner_id, title, description, slug, status, settings, published_at, created_at, updated_at`

func scanForm(row pgx.Row) (*model.Form, error) {
	var f model.Form
	var settings []byte
	var publishedAt *time.Time
	if err := row.Scan(&f.ID, &f.OwnerID, &f.Title, &f.Description, &f.Slug,
		&f.Status, &settings, &publishedAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, err
	}
	if len(settings) > 0 {
		if err := json.Unmarshal(settings, &f.Settings); err != nil {
			return nil, fmt.Errorf("decode form settings: %w", err)
		}
	}
	f.PublishedAt = publishedAt
	return &f, nil
}

// Create inserts a new form and fills server-assigned fields.
func (r *FormRepo) Create(ctx context.Context, f *model.Form) error {
	settings, err := json.Marshal(f.Settings)
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	const q = `
		INSERT INTO forms (owner_id, title, description, slug, status, settings)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + formCols
	created, err := scanForm(r.pool.QueryRow(ctx, q,
		f.OwnerID, f.Title, f.Description, f.Slug, f.Status, settings))
	if err != nil {
		return fmt.Errorf("create form: %w", err)
	}
	*f = *created
	return nil
}

// GetByIDOwned returns the owner's form (nil, nil when absent or not owned).
func (r *FormRepo) GetByIDOwned(ctx context.Context, id, ownerID string) (*model.Form, error) {
	const q = `SELECT ` + formCols + ` FROM forms
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`
	f, err := scanForm(r.pool.QueryRow(ctx, q, id, ownerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get form: %w", err)
	}
	return f, nil
}

// ListByOwner returns a page of the owner's forms with question counts, plus the
// total count for pagination.
func (r *FormRepo) ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]model.FormListItem, int, error) {
	listQ := `
		SELECT ` + prefixCols("f", formCols) + `,
		       COUNT(q.id) AS question_count,
		       (SELECT COUNT(*) FROM responses r WHERE r.form_id = f.id) AS response_count
		FROM forms f
		LEFT JOIN questions q ON q.form_id = f.id
		WHERE f.owner_id = $1 AND f.deleted_at IS NULL
		GROUP BY f.id
		ORDER BY f.updated_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, listQ, ownerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list forms: %w", err)
	}
	defer rows.Close()

	var items []model.FormListItem
	for rows.Next() {
		var it model.FormListItem
		var settings []byte
		var publishedAt *time.Time
		if err := rows.Scan(&it.ID, &it.OwnerID, &it.Title, &it.Description, &it.Slug,
			&it.Status, &settings, &publishedAt, &it.CreatedAt, &it.UpdatedAt, &it.QuestionCount, &it.ResponseCount); err != nil {
			return nil, 0, fmt.Errorf("scan form list: %w", err)
		}
		if len(settings) > 0 {
			_ = json.Unmarshal(settings, &it.Settings)
		}
		it.PublishedAt = publishedAt
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM forms WHERE owner_id = $1 AND deleted_at IS NULL`, ownerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count forms: %w", err)
	}
	return items, total, nil
}

// UpdateMeta updates editable fields and returns the new row (nil when not
// owned / absent).
func (r *FormRepo) UpdateMeta(ctx context.Context, ownerID, id, title, description, slug string, settings model.FormSettings) (*model.Form, error) {
	s, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("encode settings: %w", err)
	}
	const q = `
		UPDATE forms SET title = $3, description = $4, slug = $5, settings = $6, updated_at = now()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
		RETURNING ` + formCols
	f, err := scanForm(r.pool.QueryRow(ctx, q, id, ownerID, title, description, slug, s))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update form: %w", err)
	}
	return f, nil
}

// SetStatus transitions a form's status (publish/unpublish/close).
func (r *FormRepo) SetStatus(ctx context.Context, ownerID, id string, status model.FormStatus, publishedAt *time.Time) (*model.Form, error) {
	const q = `
		UPDATE forms SET status = $3, published_at = COALESCE($4, published_at), updated_at = now()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
		RETURNING ` + formCols
	f, err := scanForm(r.pool.QueryRow(ctx, q, id, ownerID, status, publishedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("set form status: %w", err)
	}
	return f, nil
}

// SoftDelete marks a form deleted; returns false when nothing matched.
func (r *FormRepo) SoftDelete(ctx context.Context, ownerID, id string) (bool, error) {
	const q = `UPDATE forms SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`
	tag, err := r.pool.Exec(ctx, q, id, ownerID)
	if err != nil {
		return false, fmt.Errorf("soft delete form: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// GetPublishedBySlug resolves a published, non-deleted form by slug for the
// public runner (nil, nil when absent or not published).
func (r *FormRepo) GetPublishedBySlug(ctx context.Context, slug string) (*model.Form, error) {
	const q = `SELECT ` + formCols + ` FROM forms
		WHERE slug = $1 AND status = 'published' AND deleted_at IS NULL`
	f, err := scanForm(r.pool.QueryRow(ctx, q, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get form by slug: %w", err)
	}
	return f, nil
}

// SlugExists reports whether a live form already uses slug.
func (r *FormRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM forms WHERE slug = $1 AND deleted_at IS NULL)`, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("slug exists: %w", err)
	}
	return exists, nil
}
