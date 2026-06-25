package model

import (
	"encoding/json"
	"time"
)

// Answer is one respondent's answer to one question. Value is the raw JSON leaf
// whose shape depends on the question type (string, number, []optionID, bool).
type Answer struct {
	ID         string          `json:"id"`
	ResponseID string          `json:"response_id"`
	QuestionID string          `json:"question_id"`
	Value      json.RawMessage `json:"value"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ResponseMeta is the small JSONB context captured per submission.
type ResponseMeta struct {
	UserAgent string `json:"user_agent,omitempty"`
	Referrer  string `json:"referrer,omitempty"`
}

// Response is one (anonymous) submission of a form.
type Response struct {
	ID          string       `json:"id"`
	FormID      string       `json:"form_id"`
	Completed   bool         `json:"completed"`
	Meta        ResponseMeta `json:"meta"`
	SubmittedAt time.Time    `json:"submitted_at"`
	Answers     []Answer     `json:"answers,omitempty"`
}
