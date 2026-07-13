package dto

import (
	"encoding/json"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// SubmitAnswer is one answer in a submission. Value is preserved as raw JSON so
// the server validates it against the question's type.
type SubmitAnswer struct {
	QuestionID string          `json:"question_id" binding:"required"`
	Value      json.RawMessage `json:"value"`
}

// SubmitResponseRequest is the body for POST /public/forms/:slug/responses.
// ResponseID is optional: when the runner opened a paradata session (POST
// /start), it echoes that id here so Submit finalizes that in-progress row instead
// of inserting a new one — keeping one row per respondent. Absent (no-JS / old
// client / start failed) → a fresh completed row is inserted, exactly as before.
type SubmitResponseRequest struct {
	ResponseID string         `json:"response_id"`
	Answers    []SubmitAnswer `json:"answers" binding:"required"`
}

// ProgressRequest is the body for POST /public/forms/:slug/progress — a runner
// beacon advancing an in-progress response's furthest-reached question position.
type ProgressRequest struct {
	ResponseID string `json:"response_id" binding:"required"`
	Position   int    `json:"position"`
}

// PublicForm is the runner contract: everything the respondent UI needs and
// nothing about the owner.
type PublicForm struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Slug        string             `json:"slug"`
	Settings    model.FormSettings `json:"settings"`
	Questions   []model.Question   `json:"questions"`
	LogicRules  []model.LogicRule  `json:"logic_rules"`
}
