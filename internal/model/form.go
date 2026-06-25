package model

import "time"

// FormStatus mirrors the form_status Postgres enum (migration 002).
type FormStatus string

const (
	FormDraft     FormStatus = "draft"
	FormPublished FormStatus = "published"
	FormClosed    FormStatus = "closed"
)

// ThemeSettings selects the runner's color theme. Preset is one of the named
// presets (corporate/fun/comical/girl/boy) resolved by the frontend; Accent is
// a legacy single-color override (hex/hsl) kept for backward compatibility.
type ThemeSettings struct {
	Preset string `json:"preset,omitempty"`
	Accent string `json:"accent,omitempty"`
}

// FormSettings is the JSONB blob of presentation/behaviour options.
type FormSettings struct {
	WelcomeTitle        string        `json:"welcome_title,omitempty"`
	WelcomeDescription  string        `json:"welcome_description,omitempty"`
	StartButtonText     string        `json:"start_button_text,omitempty"`
	ThankYouTitle       string        `json:"thank_you_title,omitempty"`
	ThankYouDescription string        `json:"thank_you_description,omitempty"`
	RedirectURL         string        `json:"redirect_url,omitempty"`
	BannerURL           string        `json:"banner_url,omitempty"`
	LogoURL             string        `json:"logo_url,omitempty"`
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
	LogicRules  []LogicRule  `json:"logic_rules,omitempty"`
}

// FormListItem is a form summary row for the dashboard list.
type FormListItem struct {
	Form
	QuestionCount int `json:"question_count"`
	ResponseCount int `json:"response_count"`
}
