package service

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
)

func q(id string, pos int, typ model.QuestionType) model.Question {
	return model.Question{ID: id, Position: pos, Type: typ}
}

func strptr(s string) *string { return &s }

func rule(src string, op model.LogicOperator, compare string, action model.LogicAction, target *string, priority int) model.LogicRule {
	var cv json.RawMessage
	if compare != "" {
		cv = json.RawMessage(compare)
	}
	return model.LogicRule{SourceQuestionID: src, Operator: op, CompareValue: cv, Action: action, TargetQuestionID: target, Priority: priority}
}

func TestReachablePath(t *testing.T) {
	// q1 -> q2 -> q3 -> q4 by position.
	qs := []model.Question{
		q("q1", 0, model.QYesNo),
		q("q2", 1, model.QShortText),
		q("q3", 2, model.QShortText),
		q("q4", 3, model.QShortText),
	}

	tests := []struct {
		name    string
		rules   []model.LogicRule
		answers map[string]json.RawMessage
		want    []string
	}{
		{
			name:    "linear, no rules",
			answers: map[string]json.RawMessage{},
			want:    []string{"q1", "q2", "q3", "q4"},
		},
		{
			name:    "jump skips q2 when q1=yes",
			rules:   []model.LogicRule{rule("q1", model.OpEq, "true", model.ActionJumpTo, strptr("q3"), 0)},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("true")},
			want:    []string{"q1", "q3", "q4"},
		},
		{
			name:    "jump not taken when q1=no",
			rules:   []model.LogicRule{rule("q1", model.OpEq, "true", model.ActionJumpTo, strptr("q3"), 0)},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("false")},
			want:    []string{"q1", "q2", "q3", "q4"},
		},
		{
			name:    "end_form stops early",
			rules:   []model.LogicRule{rule("q1", model.OpEq, "true", model.ActionEndForm, nil, 0)},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("true")},
			want:    []string{"q1"},
		},
		{
			name:    "hide q3 when q1=yes",
			rules:   []model.LogicRule{rule("q1", model.OpEq, "true", model.ActionHide, strptr("q3"), 0)},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("true")},
			want:    []string{"q1", "q2", "q4"},
		},
		{
			name:    "show q3 only when q1=yes (default hidden)",
			rules:   []model.LogicRule{rule("q1", model.OpEq, "true", model.ActionShow, strptr("q3"), 0)},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("false")},
			want:    []string{"q1", "q2", "q4"},
		},
		{
			name:    "show q3 included when q1=yes",
			rules:   []model.LogicRule{rule("q1", model.OpEq, "true", model.ActionShow, strptr("q3"), 0)},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("true")},
			want:    []string{"q1", "q2", "q3", "q4"},
		},
		{
			name:    "always jump skips q2/q3 regardless of answer",
			rules:   []model.LogicRule{rule("q1", model.OpAlways, "", model.ActionJumpTo, strptr("q4"), 0)},
			answers: map[string]json.RawMessage{},
			want:    []string{"q1", "q4"},
		},
		{
			name:    "always jump fires even when source unanswered",
			rules:   []model.LogicRule{rule("q2", model.OpAlways, "", model.ActionJumpTo, strptr("q4"), 0)},
			answers: map[string]json.RawMessage{},
			want:    []string{"q1", "q2", "q4"},
		},
		{
			name: "first matching rule by priority wins",
			rules: []model.LogicRule{
				rule("q1", model.OpEq, "true", model.ActionJumpTo, strptr("q4"), 10),
				rule("q1", model.OpEq, "true", model.ActionEndForm, nil, 1),
			},
			answers: map[string]json.RawMessage{"q1": json.RawMessage("true")},
			want:    []string{"q1"}, // priority 1 (end_form) beats priority 10 (jump)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reachablePath(qs, tt.rules, tt.answers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("path = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionMatches(t *testing.T) {
	r := func(s string) json.RawMessage { return json.RawMessage(s) }
	cases := []struct {
		name    string
		answer  json.RawMessage
		op      model.LogicOperator
		compare json.RawMessage
		want    bool
	}{
		{"eq string", r(`"a"`), model.OpEq, r(`"a"`), true},
		{"eq number", r(`4`), model.OpEq, r(`4`), true},
		{"neq", r(`"a"`), model.OpNeq, r(`"b"`), true},
		{"gt true", r(`5`), model.OpGt, r(`3`), true},
		{"gt false", r(`2`), model.OpGt, r(`3`), false},
		{"lte equal", r(`3`), model.OpLte, r(`3`), true},
		{"contains array", r(`["a","b"]`), model.OpContains, r(`"b"`), true},
		{"not_contains array", r(`["a"]`), model.OpNotContains, r(`"b"`), true},
		{"contains substring", r(`"hello world"`), model.OpContains, r(`"world"`), true},
		{"is_empty on missing", nil, model.OpIsEmpty, nil, true},
		{"is_not_empty on value", r(`"x"`), model.OpIsNotEmpty, nil, true},
		{"always on value", r(`"x"`), model.OpAlways, nil, true},
		{"always on missing", nil, model.OpAlways, nil, true},
		{"unanswered eq is false", nil, model.OpEq, r(`"x"`), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := conditionMatches(c.answer, c.op, c.compare); got != c.want {
				t.Fatalf("conditionMatches = %v, want %v", got, c.want)
			}
		})
	}
}
