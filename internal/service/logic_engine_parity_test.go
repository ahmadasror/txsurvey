package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// TestLogicEngineParity is the Go half of the DUAL-ENGINE PARITY GATE. It runs the
// authoritative Go conditionMatches over the SAME fixture the TS engine test reads
// (frontend/src/lib/logicEngine.parity.test.ts). If the two engines drift, one of
// the two sides goes red. See docs/architecture/adr/001-dual-logic-engine.md.
func TestLogicEngineParity(t *testing.T) {
	raw, err := os.ReadFile("testdata/logic_parity_cases.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture struct {
		Cases []struct {
			Name         string              `json:"name"`
			Answer       json.RawMessage     `json:"answer"`
			Operator     model.LogicOperator `json:"operator"`
			CompareValue json.RawMessage     `json:"compareValue"`
			Expected     bool                `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(fixture.Cases) == 0 {
		t.Fatal("fixture has no cases")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if !model.ValidLogicOperator(c.Operator) {
				t.Fatalf("fixture uses unknown operator %q", c.Operator)
			}
			if got := conditionMatches(c.Answer, c.Operator, c.CompareValue); got != c.Expected {
				t.Fatalf("conditionMatches(%s, %s, %s) = %v, want %v",
					c.Answer, c.Operator, c.CompareValue, got, c.Expected)
			}
		})
	}
}
