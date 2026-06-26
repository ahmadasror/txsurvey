package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
)

func TestCsvSafe_NeutralisesFormulaTriggers(t *testing.T) {
	cases := map[string]string{
		"=1+1":            "'=1+1",
		"+cmd":            "'+cmd",
		"-2":              "'-2",
		"@SUM(A1)":        "'@SUM(A1)",
		"\tlead-tab":      "'\tlead-tab",
		"normal text":     "normal text",
		"a=still-literal": "a=still-literal",
		"":                "",
	}
	for in, want := range cases {
		if got := csvSafe(in); got != want {
			t.Errorf("csvSafe(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateAnswer_HardTextCap(t *testing.T) {
	q := model.Question{Type: model.QShortText} // no MaxLength set
	long, _ := json.Marshal(strings.Repeat("a", maxAnswerRunes+1))
	if _, _, err := validateAnswer(q, long); err == nil {
		t.Fatal("expected an over-cap text answer to be rejected even without MaxLength")
	}
	ok, _ := json.Marshal(strings.Repeat("a", maxAnswerRunes))
	if _, _, err := validateAnswer(q, ok); err != nil {
		t.Fatalf("answer at the cap should be accepted: %v", err)
	}
}

func TestValidateFormSettings_URLFields(t *testing.T) {
	bad := []model.FormSettings{
		{RedirectURL: "javascript:alert(1)"},
		{RedirectURL: "ftp://evil"},
		{BannerURL: "https://evil.example/x.png"},
		{LogoURL: "../../etc/passwd"},
	}
	for _, s := range bad {
		if err := validateFormSettings(s); err == nil {
			t.Errorf("expected %+v to be rejected", s)
		}
	}

	good := []model.FormSettings{
		{},
		{RedirectURL: "https://example.com/thanks"},
		{BannerURL: "uploads/abc123.png", LogoURL: "uploads/def456.webp"},
	}
	for _, s := range good {
		if err := validateFormSettings(s); err != nil {
			t.Errorf("expected %+v to be accepted, got %v", s, err)
		}
	}
}
