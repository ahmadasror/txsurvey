package model

// OptionCount is a tally for one discrete value (an option, Yes/No, or a rating
// level) within a question summary.
type OptionCount struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// QuestionSummary aggregates answers for a single question.
type QuestionSummary struct {
	QuestionID string        `json:"question_id"`
	Title      string        `json:"title"`
	Type       QuestionType  `json:"type"`
	Answered   int           `json:"answered"`
	Options    []OptionCount `json:"options,omitempty"` // choice / yes_no / rating
	Average    *float64      `json:"average,omitempty"` // number / rating
}

// FormAnalytics is the per-form analytics payload.
type FormAnalytics struct {
	ResponseCount  int               `json:"response_count"`
	CompletionRate float64           `json:"completion_rate"` // 0..1 (completed / total)
	Questions      []QuestionSummary `json:"questions"`
}
