package model

import "time"

// FormStatus mirrors the form_status Postgres enum (migration 002).
type FormStatus string

const (
	FormDraft     FormStatus = "draft"
	FormPublished FormStatus = "published"
	FormClosed    FormStatus = "closed"
)

// ThemeSettings carries the per-form accent the runner applies (as the
// --primary CSS variable). Empty falls back to the default theme.
type ThemeSettings struct {
	Accent string `json:"accent,omitempty"` // hex (#2563eb) or hsl triple
}

// FormSettings is the JSONB blob of presentation/behaviour options.
type FormSettings struct {
	WelcomeTitle        string        `json:"welcome_title,omitempty"`
	WelcomeDescription  string        `json:"welcome_description,omitempty"`
	ThankYouTitle       string        `json:"thank_you_title,omitempty"`
	ThankYouDescription string        `json:"thank_you_description,omitempty"`
	RedirectURL         string        `json:"redirect_url,omitempty"`
	ShowProgress        bool          `json:"show_progress"`
	Theme               ThemeSettings `json:"theme"`
}

// Form is a creator's survey.
type Form struct {
	ID          string       `json:"id"`
	OwnerID     string       `json:"owner_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Slug        string       `json:"slug"`
	Status      FormStatus   `json:"status"`
	Settings    FormSettings `json:"settings"`
	PublishedAt *time.Time   `json:"published_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Questions   []Question   `json:"questions,omitempty"`
}

// FormListItem is a form summary row for the dashboard list.
type FormListItem struct {
	Form
	QuestionCount int `json:"question_count"`
	ResponseCount int `json:"response_count"`
}
