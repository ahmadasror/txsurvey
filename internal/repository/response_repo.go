package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		`INSERT INTO responses (form_id, completed, meta, started_at, completed_at)
		 VALUES ($1, $2, $3, now(), CASE WHEN $2 THEN now() ELSE NULL END) RETURNING id`,
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

// FinalizeSession turns an in-progress paradata row (opened by StartSession) into
// a completed submission: it stamps completed/submitted/completed_at, refreshes
// meta, and attaches the answers — all in one transaction. This reconciles the
// runner's start-row with its submission so a completed respondent leaves ONE row,
// not a completed row PLUS an orphaned in-progress ghost (which would inflate the
// drop-off funnel by counting every completer as an abandoner too).
//
// Returns finalized=false (no error) when responseID does not match a still-in-
// progress row of THIS form — a bogus/stale id, an already-completed row, a cross-
// form id, or a malformed (non-UUID) id — so the caller falls back to a plain
// Insert of a fresh completed row (the no-paradata path is unchanged).
func (r *ResponseRepo) FinalizeSession(ctx context.Context, responseID, formID string, meta model.ResponseMeta, answers []model.Answer) (bool, error) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return false, fmt.Errorf("encode response meta: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin finalize tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE responses
		    SET completed = true, meta = $3,
		        submitted_at = now(), completed_at = now(), last_seen_at = now()
		  WHERE id = $1 AND form_id = $2 AND NOT completed`,
		responseID, formID, metaJSON)
	if err != nil {
		// A malformed (non-UUID) id can't match any row; treat it as "not
		// finalized" so the caller inserts a fresh completed row instead of 500ing.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return false, nil
		}
		return false, fmt.Errorf("finalize response: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	for _, a := range answers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO answers (response_id, question_id, value) VALUES ($1, $2, $3)`,
			responseID, a.QuestionID, []byte(a.Value),
		); err != nil {
			return false, fmt.Errorf("insert answer: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit finalize: %w", err)
	}
	return true, nil
}

// FunnelData is the raw aggregate the drop-off funnel is built from: total
// sessions started, how many completed, and — among the still-in-progress
// (abandoned) sessions — a histogram of the furthest question position reached.
type FunnelData struct {
	Starts    int
	Completed int
	// AbandonedAt[p] = in-progress sessions whose furthest_position is exactly p.
	AbandonedAt map[int]int
}

// FunnelByForm returns the drop-off aggregates for a form: every response row
// counts as a "start", completed rows are the finishers, and the non-completed
// rows are bucketed by furthest_position (the abandonment histogram). Aggregation
// into per-question retention happens in Go (ResultsService.buildFunnel), matching
// the analytics convention of not doing this in JSONB SQL.
func (r *ResponseRepo) FunnelByForm(ctx context.Context, formID string) (FunnelData, error) {
	d := FunnelData{AbandonedAt: map[int]int{}}
	if err := r.pool.QueryRow(ctx,
		`SELECT count(*), count(*) FILTER (WHERE completed)
		   FROM responses WHERE form_id = $1`, formID).Scan(&d.Starts, &d.Completed); err != nil {
		return d, fmt.Errorf("funnel totals: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT furthest_position, count(*) FROM responses
		  WHERE form_id = $1 AND NOT completed GROUP BY furthest_position`, formID)
	if err != nil {
		return d, fmt.Errorf("funnel histogram: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var pos, n int
		if err := rows.Scan(&pos, &n); err != nil {
			return d, fmt.Errorf("scan funnel bucket: %w", err)
		}
		d.AbandonedAt[pos] = n
	}
	return d, rows.Err()
}

// CountByForm returns the number of COMPLETED responses for a form. In-progress
// paradata rows (completed=false, written by StartSession) are excluded so this
// count — used as the owner's response total and the completion-rate denominator
// — keeps its pre-paradata meaning.
func (r *ResponseRepo) CountByForm(ctx context.Context, formID string) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM responses WHERE form_id = $1 AND completed`, formID).Scan(&n); err != nil {
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

// DeleteByForm removes every response of a form (answers cascade via FK) and
// returns how many were deleted.
func (r *ResponseRepo) DeleteByForm(ctx context.Context, formID string) (int, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM responses WHERE form_id = $1`, formID)
	if err != nil {
		return 0, fmt.Errorf("delete responses: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// ListByForm returns a page of responses (newest first) with their answers
// attached, plus the total count.
func (r *ResponseRepo) ListByForm(ctx context.Context, formID string, limit, offset int) ([]model.Response, int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, form_id, completed, meta, submitted_at
		FROM responses WHERE form_id = $1 AND completed
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
		WHERE r.form_id = $1 AND r.completed`, formID)
	if err != nil {
		return nil, fmt.Errorf("all answers: %w", err)
	}
	defer rows.Close()
	return scanAnswers(rows)
}

// StartSession creates an in-progress (completed=false) response for a form and
// returns its id — the client holds it and pings AdvanceProgress as it navigates.
// It carries no answers and, being completed=false, is invisible to every
// owner-facing surface (all scoped to completed) until a future funnel view.
func (r *ResponseRepo) StartSession(ctx context.Context, formID string, meta model.ResponseMeta) (string, error) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("encode response meta: %w", err)
	}
	var id string
	if err := r.pool.QueryRow(ctx,
		`INSERT INTO responses (form_id, completed, meta, started_at, last_seen_at, furthest_position)
		 VALUES ($1, false, $2, now(), now(), 0) RETURNING id`,
		formID, metaJSON,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("start session: %w", err)
	}
	return id, nil
}

// AdvanceProgress bumps an in-progress response's furthest_position (monotonic —
// GREATEST so out-of-order/concurrent pings can't regress it) and last_seen_at.
// Returns (matched, exists): matched=false means the row is already completed or
// absent; exists separates "already completed" (true) from "no such id / malformed
// id" (false), so the caller can 404 the latter and no-op the former.
func (r *ResponseRepo) AdvanceProgress(ctx context.Context, responseID string, position int) (matched, exists bool, err error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE responses SET last_seen_at = now(), furthest_position = GREATEST(furthest_position, $2)
		 WHERE id = $1 AND NOT completed`, responseID, position)
	if err != nil {
		// A malformed (non-UUID) id is a client error, not a server fault.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return false, false, nil
		}
		return false, false, fmt.Errorf("advance progress: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return true, true, nil
	}
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM responses WHERE id = $1)`, responseID).Scan(&exists); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return false, false, nil
		}
		return false, false, fmt.Errorf("check response exists: %w", err)
	}
	return false, exists, nil
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
