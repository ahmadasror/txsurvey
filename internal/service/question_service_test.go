package service

import (
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
)

func f64(v float64) *float64 { return &v }

func TestValidateQuestion(t *testing.T) {
	tests := []struct {
		name    string
		in      model.Question
		wantErr bool
		check   func(t *testing.T, q model.Question)
	}{
		{
			name: "short text ok",
			in:   model.Question{Type: model.QShortText, Title: "Name"},
		},
		{
			name:    "empty title rejected",
			in:      model.Question{Type: model.QShortText, Title: "   "},
			wantErr: true,
		},
		{
			name:    "unknown type rejected",
			in:      model.Question{Type: "bogus", Title: "x"},
			wantErr: true,
		},
		{
			name:    "choice without options rejected",
			in:      model.Question{Type: model.QMultipleChoice, Title: "Pick"},
			wantErr: true,
		},
		{
			name: "choice generates option ids and clears scale",
			in: model.Question{
				Type:  model.QMultipleChoice,
				Title: "Pick",
				Metadata: model.QuestionMetadata{
					Scale:   7,
					Options: []model.Option{{Label: "A"}, {Label: "B"}},
				},
			},
			check: func(t *testing.T, q model.Question) {
				if q.Metadata.Scale != 0 {
					t.Errorf("scale should be cleared, got %d", q.Metadata.Scale)
				}
				for _, o := range q.Metadata.Options {
					if o.ID == "" {
						t.Errorf("option id not generated for %q", o.Label)
					}
				}
			},
		},
		{
			name:    "choice with blank label rejected",
			in:      model.Question{Type: model.QDropdown, Title: "Pick", Metadata: model.QuestionMetadata{Options: []model.Option{{Label: " "}}}},
			wantErr: true,
		},
		{
			name: "rating defaults scale to 5",
			in:   model.Question{Type: model.QRating, Title: "Rate"},
			check: func(t *testing.T, q model.Question) {
				if q.Metadata.Scale != 5 {
					t.Errorf("want default scale 5, got %d", q.Metadata.Scale)
				}
			},
		},
		{
			name:    "rating scale out of range rejected",
			in:      model.Question{Type: model.QRating, Title: "Rate", Metadata: model.QuestionMetadata{Scale: 11}},
			wantErr: true,
		},
		{
			name:    "number min greater than max rejected",
			in:      model.Question{Type: model.QNumber, Title: "Age", Metadata: model.QuestionMetadata{Min: f64(10), Max: f64(5)}},
			wantErr: true,
		},
		{
			name: "statement forces not-required and clears metadata",
			in: model.Question{
				Type:     model.QStatement,
				Title:    "Welcome",
				Required: true,
				Metadata: model.QuestionMetadata{Placeholder: "x"},
			},
			check: func(t *testing.T, q model.Question) {
				if q.Required {
					t.Error("statement must not be required")
				}
				if q.Metadata.Placeholder != "" {
					t.Error("statement metadata must be cleared")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.in
			err := validateQuestion(&q)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil && err == nil {
				tt.check(t, q)
			}
		})
	}
}
