package model

import "time"

// QuestionType enumerates the supported field types (must match the
// question_type Postgres enum in migration 003).
type QuestionType string

const (
	QShortText      QuestionType = "short_text"
	QLongText       QuestionType = "long_text"
	QMultipleChoice QuestionType = "multiple_choice" // single select
	QCheckboxes     QuestionType = "checkboxes"      // multi select
	QDropdown       QuestionType = "dropdown"
	QRating         QuestionType = "rating"
	QNumber         QuestionType = "number"
	QEmail          QuestionType = "email"
	QDate           QuestionType = "date"
	QYesNo          QuestionType = "yes_no"
	QStatement      QuestionType = "statement" // display-only, never answered
)

// ValidQuestionType reports whether t is a known question type.
func ValidQuestionType(t QuestionType) bool {
	switch t {
	case QShortText, QLongText, QMultipleChoice, QCheckboxes, QDropdown,
		QRating, QNumber, QEmail, QDate, QYesNo, QStatement:
		return true
	}
	return false
}

// IsChoice reports whether the type carries a fixed option set.
func (t QuestionType) IsChoice() bool {
	return t == QMultipleChoice || t == QCheckboxes || t == QDropdown
}

// Option is one selectable choice for a choice-type question.
type Option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// QuestionMetadata is the per-type configuration stored as JSONB. Only the
// fields relevant to a question's type are populated.
type QuestionMetadata struct {
	Options     []Option `json:"options,omitempty"`     // choice types
	Min         *float64 `json:"min,omitempty"`         // number
	Max         *float64 `json:"max,omitempty"`         // number
	Step        *float64 `json:"step,omitempty"`        // number
	Scale       int      `json:"scale,omitempty"`       // rating (e.g. 5 or 10)
	MaxLength   int      `json:"max_length,omitempty"`  // text
	Placeholder string   `json:"placeholder,omitempty"` // text/number
	AllowOther  bool     `json:"allow_other,omitempty"` // choice types
}

// Question is a single field of a form.
type Question struct {
	ID          string           `json:"id"`
	FormID      string           `json:"form_id"`
	Type        QuestionType     `json:"type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Position    int              `json:"position"`
	Required    bool             `json:"required"`
	Metadata    QuestionMetadata `json:"metadata"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
