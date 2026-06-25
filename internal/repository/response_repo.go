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

// ResponseRepo persists submissions and their answers.
type ResponseRepo struct {
	pool *pgxpool.Pool
}

func NewResponseRepo(pool *pgxpool.Pool) *ResponseRepo {
	return &ResponseRepo{pool: pool}
}

// Insert writes a response and all its answers in one transaction, returning the
// new response id.
func (r *ResponseRepo) Insert(ctx context.Context, formID string, completed bool, meta model.ResponseMeta, answers []model.Answer) (string, error) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("encode response meta: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin response tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var responseID string
	if err := tx.QueryRow(ctx,
		`INSERT INTO responses (form_id, completed, meta) VALUES ($1, $2, $3) RETURNING id`,
		formID, completed, metaJSON,
	).Scan(&responseID); err != nil {
		return "", fmt.Errorf("insert response: %w", err)
	}

	for _, a := range answers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO answers (response_id, question_id, value) VALUES ($1, $2, $3)`,
			responseID, a.QuestionID, []byte(a.Value),
		); err != nil {
			return "", fmt.Errorf("insert answer: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit response: %w", err)
	}
	return responseID, nil
}

// CountByForm returns the total number of responses for a form.
func (r *ResponseRepo) CountByForm(ctx context.Context, formID string) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM responses WHERE form_id = $1`, formID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count responses: %w", err)
	}
	return n, nil
}

// CompletedCountByForm returns how many responses are marked completed.
func (r *ResponseRepo) CompletedCountByForm(ctx context.Context, formID string) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM responses WHERE form_id = $1 AND completed`, formID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count completed responses: %w", err)
	}
	return n, nil
}

// ListByForm returns a page of responses (newest first) with their answers
// attached, plus the total count.
func (r *ResponseRepo) ListByForm(ctx context.Context, formID string, limit, offset int) ([]model.Response, int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, form_id, completed, meta, submitted_at
		FROM responses WHERE form_id = $1
		ORDER BY submitted_at DESC LIMIT $2 OFFSET $3`, formID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list responses: %w", err)
	}
	defer rows.Close()

	var responses []model.Response
	ids := make([]string, 0)
	for rows.Next() {
		resp, err := scanResponse(rows)
		if err != nil {
			return nil, 0, err
		}
		responses = append(responses, *resp)
		ids = append(ids, resp.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.attachAnswers(ctx, responses, ids); err != nil {
		return nil, 0, err
	}

	total, err := r.CountByForm(ctx, formID)
	if err != nil {
		return nil, 0, err
	}
	return responses, total, nil
}

// GetByID returns one response (scoped to its form) with answers, or nil.
func (r *ResponseRepo) GetByID(ctx context.Context, formID, responseID string) (*model.Response, error) {
	resp, err := scanResponse(r.pool.QueryRow(ctx, `
		SELECT id, form_id, completed, meta, submitted_at
		FROM responses WHERE id = $1 AND form_id = $2`, responseID, formID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get response: %w", err)
	}
	list := []model.Response{*resp}
	if err := r.attachAnswers(ctx, list, []string{resp.ID}); err != nil {
		return nil, err
	}
	return &list[0], nil
}

// AllAnswers returns every answer across all responses of a form (for analytics
// aggregation). At this scale loading all rows is cheap and avoids fragile
// JSONB SQL.
func (r *ResponseRepo) AllAnswers(ctx context.Context, formID string) ([]model.Answer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.response_id, a.question_id, a.value, a.created_at
		FROM answers a JOIN responses r ON r.id = a.response_id
		WHERE r.form_id = $1`, formID)
	if err != nil {
		return nil, fmt.Errorf("all answers: %w", err)
	}
	defer rows.Close()
	return scanAnswers(rows)
}

// attachAnswers loads answers for the given response ids and attaches them.
func (r *ResponseRepo) attachAnswers(ctx context.Context, responses []model.Response, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, response_id, question_id, value, created_at
		FROM answers WHERE response_id = ANY($1) ORDER BY created_at`, ids)
	if err != nil {
		return fmt.Errorf("load answers: %w", err)
	}
	defer rows.Close()
	answers, err := scanAnswers(rows)
	if err != nil {
		return err
	}
	byResp := make(map[string][]model.Answer, len(responses))
	for _, a := range answers {
		byResp[a.ResponseID] = append(byResp[a.ResponseID], a)
	}
	for i := range responses {
		responses[i].Answers = byResp[responses[i].ID]
	}
	return nil
}

func scanResponse(row pgx.Row) (*model.Response, error) {
	var resp model.Response
	var meta []byte
	if err := row.Scan(&resp.ID, &resp.FormID, &resp.Completed, &meta, &resp.SubmittedAt); err != nil {
		return nil, err
	}
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &resp.Meta)
	}
	return &resp, nil
}

func scanAnswers(rows pgx.Rows) ([]model.Answer, error) {
	var out []model.Answer
	for rows.Next() {
		var a model.Answer
		var value []byte
		if err := rows.Scan(&a.ID, &a.ResponseID, &a.QuestionID, &value, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		a.Value = json.RawMessage(value)
		out = append(out, a)
	}
	return out, rows.Err()
}
