package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// LogicRepo persists a form's logic rules. Ownership is enforced one level up
// (the service loads the owning form first), so these scope by form_id.
type LogicRepo struct {
	pool *pgxpool.Pool
}

func NewLogicRepo(pool *pgxpool.Pool) *LogicRepo {
	return &LogicRepo{pool: pool}
}

const logicCols = `id, form_id, source_question_id, operator, compare_value, action, target_question_id, priority, created_at`

func scanRule(row pgx.Row) (*model.LogicRule, error) {
	var r model.LogicRule
	var compare []byte
	var target *string
	if err := row.Scan(&r.ID, &r.FormID, &r.SourceQuestionID, &r.Operator, &compare,
		&r.Action, &target, &r.Priority, &r.CreatedAt); err != nil {
		return nil, err
	}
	if len(compare) > 0 {
		r.CompareValue = append([]byte(nil), compare...)
	}
	r.TargetQuestionID = target
	return &r, nil
}

// ListByForm returns a form's rules, ordered for deterministic evaluation.
func (r *LogicRepo) ListByForm(ctx context.Context, formID string) ([]model.LogicRule, error) {
	const q = `SELECT ` + logicCols + ` FROM logic_rules
		WHERE form_id = $1 ORDER BY source_question_id, priority, created_at`
	rows, err := r.pool.Query(ctx, q, formID)
	if err != nil {
		return nil, fmt.Errorf("list logic rules: %w", err)
	}
	defer rows.Close()

	var out []model.LogicRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rule)
	}
	return out, rows.Err()
}

// Insert creates a rule.
func (r *LogicRepo) Insert(ctx context.Context, rule *model.LogicRule) error {
	const q = `
		INSERT INTO logic_rules (form_id, source_question_id, operator, compare_value, action, target_question_id, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + logicCols
	created, err := scanRule(r.pool.QueryRow(ctx, q,
		rule.FormID, rule.SourceQuestionID, rule.Operator, nullableJSON(rule.CompareValue),
		rule.Action, rule.TargetQuestionID, rule.Priority))
	if err != nil {
		return fmt.Errorf("insert logic rule: %w", err)
	}
	*rule = *created
	return nil
}

// Update mutates a rule (scoped to its form); nil when not found.
func (r *LogicRepo) Update(ctx context.Context, formID, id string, rule *model.LogicRule) (*model.LogicRule, error) {
	const q = `
		UPDATE logic_rules
		SET operator = $3, compare_value = $4, action = $5, target_question_id = $6, priority = $7
		WHERE id = $1 AND form_id = $2
		RETURNING ` + logicCols
	updated, err := scanRule(r.pool.QueryRow(ctx, q,
		id, formID, rule.Operator, nullableJSON(rule.CompareValue), rule.Action, rule.TargetQuestionID, rule.Priority))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update logic rule: %w", err)
	}
	return updated, nil
}

// Delete removes a rule from its form; false when nothing matched.
func (r *LogicRepo) Delete(ctx context.Context, formID, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM logic_rules WHERE id = $1 AND form_id = $2`, id, formID)
	if err != nil {
		return false, fmt.Errorf("delete logic rule: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// nullableJSON converts an empty RawMessage to a SQL NULL.
func nullableJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return []byte(b)
}
