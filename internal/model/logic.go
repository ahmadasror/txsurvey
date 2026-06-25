package model

import (
	"encoding/json"
	"time"
)

// LogicOperator mirrors the logic_operator Postgres enum (migration 005).
type LogicOperator string

const (
	OpEq          LogicOperator = "eq"
	OpNeq         LogicOperator = "neq"
	OpGt          LogicOperator = "gt"
	OpGte         LogicOperator = "gte"
	OpLt          LogicOperator = "lt"
	OpLte         LogicOperator = "lte"
	OpContains    LogicOperator = "contains"
	OpNotContains LogicOperator = "not_contains"
	OpIsEmpty     LogicOperator = "is_empty"
	OpIsNotEmpty  LogicOperator = "is_not_empty"
)

// ValidLogicOperator reports whether op is known.
func ValidLogicOperator(op LogicOperator) bool {
	switch op {
	case OpEq, OpNeq, OpGt, OpGte, OpLt, OpLte, OpContains, OpNotContains, OpIsEmpty, OpIsNotEmpty:
		return true
	}
	return false
}

// LogicAction mirrors the logic_action Postgres enum.
type LogicAction string

const (
	ActionJumpTo  LogicAction = "jump_to"
	ActionShow    LogicAction = "show"
	ActionHide    LogicAction = "hide"
	ActionEndForm LogicAction = "end_form"
)

// ValidLogicAction reports whether a is known.
func ValidLogicAction(a LogicAction) bool {
	switch a {
	case ActionJumpTo, ActionShow, ActionHide, ActionEndForm:
		return true
	}
	return false
}

// LogicRule is one conditional rule attached to a source question.
type LogicRule struct {
	ID               string          `json:"id"`
	FormID           string          `json:"form_id"`
	SourceQuestionID string          `json:"source_question_id"`
	Operator         LogicOperator   `json:"operator"`
	CompareValue     json.RawMessage `json:"compare_value,omitempty"`
	Action           LogicAction     `json:"action"`
	TargetQuestionID *string         `json:"target_question_id,omitempty"`
	Priority         int             `json:"priority"`
	CreatedAt        time.Time       `json:"created_at"`
}
