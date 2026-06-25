package service

import (
	"encoding/json"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
)

func raw(s string) json.RawMessage { return json.RawMessage(s) }

func TestValidateAnswer(t *testing.T) {
	choice := model.Question{
		Type:     model.QMultipleChoice,
		Metadata: model.QuestionMetadata{Options: []model.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}},
	}
	checks := model.Question{
		Type:     model.QCheckboxes,
		Metadata: model.QuestionMetadata{Options: []model.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}},
	}
	num := model.Question{Type: model.QNumber, Metadata: model.QuestionMetadata{Min: f64(1), Max: f64(10)}}
	rating := model.Question{Type: model.QRating, Metadata: model.QuestionMetadata{Scale: 5}}

	tests := []struct {
		name      string
		q         model.Question
		in        json.RawMessage
		wantEmpty bool
		wantErr   bool
	}{
		{"text ok", model.Question{Type: model.QShortText}, raw(`"hi"`), false, false},
		{"text blank is empty", model.Question{Type: model.QShortText}, raw(`"  "`), true, false},
		{"null is empty", model.Question{Type: model.QShortText}, raw(`null`), true, false},
		{"email ok", model.Question{Type: model.QEmail}, raw(`"a@b.com"`), false, false},
		{"email invalid", model.Question{Type: model.QEmail}, raw(`"nope"`), false, true},
		{"number in range", num, raw(`5`), false, false},
		{"number below min", num, raw(`0`), false, true},
		{"number above max", num, raw(`11`), false, true},
		{"number not a number", num, raw(`"x"`), false, true},
		{"rating ok", rating, raw(`4`), false, false},
		{"rating out of range", rating, raw(`6`), false, true},
		{"rating non-integer", rating, raw(`3.5`), false, true},
		{"yes_no bool", model.Question{Type: model.QYesNo}, raw(`true`), false, false},
		{"yes_no string", model.Question{Type: model.QYesNo}, raw(`"yes"`), false, false},
		{"choice valid", choice, raw(`"a"`), false, false},
		{"choice invalid option", choice, raw(`"zzz"`), false, true},
		{"checkboxes valid", checks, raw(`["a","b"]`), false, false},
		{"checkboxes empty is empty", checks, raw(`[]`), true, false},
		{"checkboxes invalid option", checks, raw(`["a","x"]`), false, true},
		{"checkboxes duplicate", checks, raw(`["a","a"]`), false, true},
		{"date ok", model.Question{Type: model.QDate}, raw(`"2026-06-25"`), false, false},
		{"date bad", model.Question{Type: model.QDate}, raw(`"25/06/2026"`), false, true},
		{"statement not answerable", model.Question{Type: model.QStatement}, raw(`"x"`), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, empty, err := validateAnswer(tt.q, tt.in)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && empty != tt.wantEmpty {
				t.Fatalf("empty = %v, want %v", empty, tt.wantEmpty)
			}
		})
	}
}
