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

// FunnelStep is one question's retention in the drop-off funnel.
type FunnelStep struct {
	QuestionID string `json:"question_id"`
	Title      string `json:"title"`
	Position   int    `json:"position"`
	Reached    int    `json:"reached"` // respondents who got at least this far
}

// FormFunnel is the response drop-off funnel: how far respondents get before
// abandoning. `Reached` is a position-based approximation — paradata stores only
// the furthest question position a session saw, not its exact (logic-branched)
// path — and is monotonically non-increasing down the form. `Starts` counts every
// opened session (completed + abandoned); `Completed` counts finishers.
type FormFunnel struct {
	Starts    int          `json:"starts"`
	Completed int          `json:"completed"`
	Steps     []FunnelStep `json:"steps"`
}
