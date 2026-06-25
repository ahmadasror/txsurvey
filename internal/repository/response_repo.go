package repository

import (
	"context"
	"encoding/json"
	"fmt"

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
