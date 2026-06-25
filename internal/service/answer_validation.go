package service

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func answerErr(msg string) error {
	return apperror.New(http.StatusUnprocessableEntity, "INVALID_ANSWER", msg)
}

// validateAnswer validates a raw answer against its question's type, returning
// the canonical JSON value to store and whether the answer is effectively empty
// (blank/omitted). Statement questions are not answerable and must be filtered
// by the caller. Pure — unit-tested directly.
func validateAnswer(q model.Question, raw json.RawMessage) (json.RawMessage, bool, error) {
	if isJSONEmpty(raw) {
		return nil, true, nil
	}

	switch q.Type {
	case model.QShortText, model.QLongText:
		s, err := decodeString(raw)
		if err != nil {
			return nil, false, answerErr("expected a text value")
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, true, nil
		}
		if q.Metadata.MaxLength > 0 && len([]rune(s)) > q.Metadata.MaxLength {
			return nil, false, answerErr("answer exceeds the maximum length")
		}
		return mustJSON(s), false, nil

	case model.QDate:
		s, err := decodeString(raw)
		if err != nil {
			return nil, false, answerErr("expected a date")
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, true, nil
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return nil, false, answerErr("expected a date in YYYY-MM-DD format")
		}
		return mustJSON(s), false, nil

	case model.QEmail:
		s, err := decodeString(raw)
		if err != nil {
			return nil, false, answerErr("expected an email address")
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, true, nil
		}
		if !emailRe.MatchString(s) {
			return nil, false, answerErr("invalid email address")
		}
		return mustJSON(s), false, nil

	case model.QNumber:
		f, err := decodeNumber(raw)
		if err != nil {
			return nil, false, answerErr("expected a number")
		}
		if q.Metadata.Min != nil && f < *q.Metadata.Min {
			return nil, false, answerErr("number is below the minimum")
		}
		if q.Metadata.Max != nil && f > *q.Metadata.Max {
			return nil, false, answerErr("number is above the maximum")
		}
		return mustJSON(f), false, nil

	case model.QRating:
		f, err := decodeNumber(raw)
		if err != nil {
			return nil, false, answerErr("expected a rating")
		}
		scale := q.Metadata.Scale
		if scale == 0 {
			scale = 5
		}
		if f < 1 || f > float64(scale) || f != float64(int(f)) {
			return nil, false, answerErr("rating is out of range")
		}
		return mustJSON(int(f)), false, nil

	case model.QYesNo:
		b, err := decodeBool(raw)
		if err != nil {
			return nil, false, answerErr("expected yes or no")
		}
		return mustJSON(b), false, nil

	case model.QMultipleChoice, model.QDropdown:
		s, err := decodeString(raw)
		if err != nil {
			return nil, false, answerErr("expected a selected option")
		}
		if s == "" {
			return nil, true, nil
		}
		if !optionExists(q, s) {
			return nil, false, answerErr("selected option is not valid")
		}
		return mustJSON(s), false, nil

	case model.QCheckboxes:
		var arr []string
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, false, answerErr("expected a list of selected options")
		}
		if len(arr) == 0 {
			return nil, true, nil
		}
		seen := make(map[string]bool, len(arr))
		for _, id := range arr {
			if !optionExists(q, id) {
				return nil, false, answerErr("a selected option is not valid")
			}
			if seen[id] {
				return nil, false, answerErr("duplicate option selected")
			}
			seen[id] = true
		}
		return mustJSON(arr), false, nil

	default:
		return nil, false, answerErr("question is not answerable")
	}
}

func isJSONEmpty(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null" || s == `""` || s == "[]"
}

func decodeString(raw json.RawMessage) (string, error) {
	var s string
	err := json.Unmarshal(raw, &s)
	return s, err
}

func decodeNumber(raw json.RawMessage) (float64, error) {
	var f float64
	err := json.Unmarshal(raw, &f)
	return f, err
}

func decodeBool(raw json.RawMessage) (bool, error) {
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "yes", "true":
			return true, nil
		case "no", "false":
			return false, nil
		}
	}
	return false, answerErr("not a boolean")
}

func optionExists(q model.Question, id string) bool {
	for _, o := range q.Metadata.Options {
		if o.ID == id {
			return true
		}
	}
	return false
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
