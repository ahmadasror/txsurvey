package service

import (
	"encoding/json"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
)

func TestValidateRule(t *testing.T) {
	questions := []model.Question{
		q("q1", 0, model.QYesNo),
		q("q2", 1, model.QShortText),
		q("q3", 2, model.QShortText),
	}

	tests := []struct {
		name    string
		in      dto.LogicRuleInput
		wantErr bool
	}{
		{
			name: "valid forward jump",
			in:   dto.LogicRuleInput{SourceQuestionID: "q1", Operator: model.OpEq, CompareValue: json.RawMessage("true"), Action: model.ActionJumpTo, TargetQuestionID: strptr("q3")},
		},
		{
			name:    "backward jump rejected",
			in:      dto.LogicRuleInput{SourceQuestionID: "q3", Operator: model.OpEq, CompareValue: json.RawMessage(`"x"`), Action: model.ActionJumpTo, TargetQuestionID: strptr("q1")},
			wantErr: true,
		},
		{
			name:    "self target rejected",
			in:      dto.LogicRuleInput{SourceQuestionID: "q2", Operator: model.OpEq, CompareValue: json.RawMessage(`"x"`), Action: model.ActionShow, TargetQuestionID: strptr("q2")},
			wantErr: true,
		},
		{
			name:    "unknown source rejected",
			in:      dto.LogicRuleInput{SourceQuestionID: "zzz", Operator: model.OpEq, CompareValue: json.RawMessage(`"x"`), Action: model.ActionJumpTo, TargetQuestionID: strptr("q3")},
			wantErr: true,
		},
		{
			name:    "jump without target rejected",
			in:      dto.LogicRuleInput{SourceQuestionID: "q1", Operator: model.OpEq, CompareValue: json.RawMessage("true"), Action: model.ActionJumpTo},
			wantErr: true,
		},
		{
			name:    "end_form with target rejected",
			in:      dto.LogicRuleInput{SourceQuestionID: "q1", Operator: model.OpEq, CompareValue: json.RawMessage("true"), Action: model.ActionEndForm, TargetQuestionID: strptr("q3")},
			wantErr: true,
		},
		{
			name: "is_empty needs no compare value",
			in:   dto.LogicRuleInput{SourceQuestionID: "q1", Operator: model.OpIsEmpty, Action: model.ActionEndForm},
		},
		{
			name:    "eq needs compare value",
			in:      dto.LogicRuleInput{SourceQuestionID: "q1", Operator: model.OpEq, Action: model.ActionEndForm},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRule(tt.in, questions)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
