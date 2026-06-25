package dto

import (
	"encoding/json"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// LogicRuleInput is the create/update payload for a logic rule.
type LogicRuleInput struct {
	SourceQuestionID string              `json:"source_question_id" binding:"required"`
	Operator         model.LogicOperator `json:"operator" binding:"required"`
	CompareValue     json.RawMessage     `json:"compare_value"`
	Action           model.LogicAction   `json:"action" binding:"required"`
	TargetQuestionID *string             `json:"target_question_id"`
	Priority         int                 `json:"priority"`
}
