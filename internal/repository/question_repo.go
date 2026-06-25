package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// QuestionRepo persists a form's questions. Ownership is enforced one level up
// (the service loads the owning form first), so these queries scope by form_id.
type QuestionRepo struct {
	pool *pgxpool.Pool
}

func NewQuestionRepo(pool *pgxpool.Pool) *QuestionRepo {
	return &QuestionRepo{pool: pool}
}

const questionCols = `id, form_id, type, title, description, position, required, metadata, created_at, updated_at`

func scanQuestion(row pgx.Row) (*model.Question, error) {
	var q model.Question
	var meta []byte
	if err := row.Scan(&q.ID, &q.FormID, &q.Type, &q.Title, &q.Description,
		&q.Position, &q.Required, &meta, &q.CreatedAt, &q.UpdatedAt); err != nil {
		return nil, err
	}
	if len(meta) > 0 {
		if err := json.Unmarshal(meta, &q.Metadata); err != nil {
			return nil, fmt.Errorf("decode question metadata: %w", err)
		}
	}
	return &q, nil
}

// ListByForm returns a form's questions in stable presentation order.
func (r *QuestionRepo) ListByForm(ctx context.Context, formID string) ([]model.Question, error) {
	const q = `SELECT ` + questionCols + ` FROM questions
		WHERE form_id = $1 ORDER BY position, created_at, id`
	rows, err := r.pool.Query(ctx, q, formID)
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	defer rows.Close()

	var out []model.Question
	for rows.Next() {
		qq, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *qq)
	}
	return out, rows.Err()
}

// Insert appends a question at the next position (computed atomically).
func (r *QuestionRepo) Insert(ctx context.Context, q *model.Question) error {
	meta, err := json.Marshal(q.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	const ins = `
		INSERT INTO questions (form_id, type, title, description, position, required, metadata)
		VALUES ($1, $2, $3, $4,
		        COALESCE((SELECT MAX(position) + 1 FROM questions WHERE form_id = $1), 0),
		        $5, $6)
		RETURNING ` + questionCols
	created, err := scanQuestion(r.pool.QueryRow(ctx, ins,
		q.FormID, q.Type, q.Title, q.Description, q.Required, meta))
	if err != nil {
		return fmt.Errorf("insert question: %w", err)
	}
	*q = *created
	return nil
}

// Update mutates a question (scoped to its form) and returns the new row
// (nil when not found within that form).
func (r *QuestionRepo) Update(ctx context.Context, formID, id string, q *model.Question) (*model.Question, error) {
	meta, err := json.Marshal(q.Metadata)
	if err != nil {
		return nil, fmt.Errorf("encode metadata: %w", err)
	}
	const upd = `
		UPDATE questions SET type = $3, title = $4, description = $5, required = $6,
		       metadata = $7, updated_at = now()
		WHERE id = $1 AND form_id = $2
		RETURNING ` + questionCols
	updated, err := scanQuestion(r.pool.QueryRow(ctx, upd,
		id, formID, q.Type, q.Title, q.Description, q.Required, meta))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}
	return updated, nil
}

// Delete removes a question from its form; returns false when nothing matched.
func (r *QuestionRepo) Delete(ctx context.Context, formID, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM questions WHERE id = $1 AND form_id = $2`, id, formID)
	if err != nil {
		return false, fmt.Errorf("delete question: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// Reorder rewrites positions to match orderedIDs (index = new position) in a
// single transaction. Returns an error if any id does not belong to the form.
func (r *QuestionRepo) Reorder(ctx context.Context, formID string, orderedIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reorder: %w", err)
	}
	defer tx.Rollback(ctx)

	for pos, id := range orderedIDs {
		tag, err := tx.Exec(ctx,
			`UPDATE questions SET position = $3, updated_at = now() WHERE id = $1 AND form_id = $2`,
			id, formID, pos)
		if err != nil {
			return fmt.Errorf("reorder update: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("question %s not in form %s", id, formID)
		}
	}
	return tx.Commit(ctx)
}

// CountByForm returns how many questions a form has.
func (r *QuestionRepo) CountByForm(ctx context.Context, formID string) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM questions WHERE form_id = $1`, formID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count questions: %w", err)
	}
	return n, nil
}
