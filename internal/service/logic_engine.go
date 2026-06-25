package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// logic_engine.go is the AUTHORITATIVE evaluator. The TS runner
// (frontend/src/lib/logicEngine.ts) mirrors these exact semantics — keep them in
// lockstep:
//
//   - Rules for a source question are tried by ascending priority; the FIRST
//     matching navigation action (jump_to / end_form) decides the next question.
//   - show/hide rules adjust a target question's visibility: a question with any
//     show rule is hidden unless a show rule matches; a hide rule that matches
//     hides it. visible = (no show rules OR some show matches) AND no hide matches.
//   - Navigation walks questions by position, skipping non-visible ones; jump_to
//     lands on the target (then skips forward to the next visible question).
//   - A visited-set + step cap guarantee termination even if a reorder left a
//     backward jump in place.

// reachablePath replays the path a respondent with the given answers should have
// taken, returning the ordered set of reachable (answerable) question ids.
func reachablePath(questions []model.Question, rules []model.LogicRule, answers map[string]json.RawMessage) []string {
	ordered := sortedByPosition(questions)
	if len(ordered) == 0 {
		return nil
	}
	posByID := make(map[string]int, len(ordered))
	for i, q := range ordered {
		posByID[q.ID] = i
	}
	qByID := make(map[string]model.Question, len(ordered))
	for _, q := range ordered {
		qByID[q.ID] = q
	}

	jumpRules := rulesBySource(rules, model.ActionJumpTo, model.ActionEndForm)
	showRules := rulesByTarget(rules, model.ActionShow)
	hideRules := rulesByTarget(rules, model.ActionHide)

	visible := func(qid string) bool {
		hidden := anyMatch(hideRules[qid], qByID, answers)
		shows := showRules[qid]
		base := true
		if len(shows) > 0 {
			base = anyMatch(shows, qByID, answers)
		}
		return base && !hidden
	}

	firstVisibleFrom := func(idx int) int {
		for i := idx; i < len(ordered); i++ {
			if visible(ordered[i].ID) {
				return i
			}
		}
		return -1
	}

	var path []string
	seen := make(map[string]bool, len(ordered))
	cur := firstVisibleFrom(0)
	for steps := 0; cur >= 0 && steps <= len(ordered); steps++ {
		q := ordered[cur]
		if seen[q.ID] {
			break // loop guard
		}
		seen[q.ID] = true
		path = append(path, q.ID)

		// Navigation: first matching jump/end rule on this question wins.
		next := -1
		jumped := false
		for _, rule := range jumpRules[q.ID] {
			if !ruleMatches(rule, qByID, answers) {
				continue
			}
			if rule.Action == model.ActionEndForm {
				next = -1
				jumped = true
				break
			}
			if rule.TargetQuestionID != nil {
				if tp, ok := posByID[*rule.TargetQuestionID]; ok {
					next = firstVisibleFrom(tp)
					jumped = true
					break
				}
			}
		}
		if !jumped {
			next = firstVisibleFrom(cur + 1)
		}
		cur = next
	}
	return path
}

// ruleMatches evaluates a rule's condition against the source question's answer.
func ruleMatches(rule model.LogicRule, qByID map[string]model.Question, answers map[string]json.RawMessage) bool {
	return conditionMatches(answers[rule.SourceQuestionID], rule.Operator, rule.CompareValue)
}

func anyMatch(rules []model.LogicRule, qByID map[string]model.Question, answers map[string]json.RawMessage) bool {
	for _, r := range rules {
		if ruleMatches(r, qByID, answers) {
			return true
		}
	}
	return false
}

// conditionMatches is the per-operator predicate. Unanswered sources match only
// is_empty.
func conditionMatches(answer json.RawMessage, op model.LogicOperator, compare json.RawMessage) bool {
	empty := isJSONEmpty(answer)
	switch op {
	case model.OpIsEmpty:
		return empty
	case model.OpIsNotEmpty:
		return !empty
	}
	if empty {
		return false
	}
	switch op {
	case model.OpEq:
		return jsonEqual(answer, compare)
	case model.OpNeq:
		return !jsonEqual(answer, compare)
	case model.OpGt, model.OpGte, model.OpLt, model.OpLte:
		a, ok1 := jsonNumber(answer)
		b, ok2 := jsonNumber(compare)
		if !ok1 || !ok2 {
			return false
		}
		switch op {
		case model.OpGt:
			return a > b
		case model.OpGte:
			return a >= b
		case model.OpLt:
			return a < b
		default:
			return a <= b
		}
	case model.OpContains:
		return jsonContains(answer, compare)
	case model.OpNotContains:
		return !jsonContains(answer, compare)
	}
	return false
}

func sortedByPosition(qs []model.Question) []model.Question {
	out := make([]model.Question, len(qs))
	copy(out, qs)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out
}

func rulesBySource(rules []model.LogicRule, actions ...model.LogicAction) map[string][]model.LogicRule {
	want := map[model.LogicAction]bool{}
	for _, a := range actions {
		want[a] = true
	}
	m := map[string][]model.LogicRule{}
	for _, r := range rules {
		if want[r.Action] {
			m[r.SourceQuestionID] = append(m[r.SourceQuestionID], r)
		}
	}
	for k := range m {
		sort.SliceStable(m[k], func(i, j int) bool { return m[k][i].Priority < m[k][j].Priority })
	}
	return m
}

func rulesByTarget(rules []model.LogicRule, action model.LogicAction) map[string][]model.LogicRule {
	m := map[string][]model.LogicRule{}
	for _, r := range rules {
		if r.Action == action && r.TargetQuestionID != nil {
			m[*r.TargetQuestionID] = append(m[*r.TargetQuestionID], r)
		}
	}
	return m
}

// --- value comparison helpers (canonicalize via decode→encode) ---

func canonicalJSON(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return strings.TrimSpace(string(raw))
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func jsonEqual(a, b json.RawMessage) bool {
	return canonicalJSON(a) == canonicalJSON(b)
}

func jsonNumber(raw json.RawMessage) (float64, bool) {
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return f, true
	}
	// Allow a numeric string ("4").
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if _, err := fmt.Sscanf(strings.TrimSpace(s), "%g", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

func jsonContains(answer, compare json.RawMessage) bool {
	// Array answer (checkboxes): membership.
	var arr []json.RawMessage
	if err := json.Unmarshal(answer, &arr); err == nil {
		cc := canonicalJSON(compare)
		for _, el := range arr {
			if canonicalJSON(el) == cc {
				return true
			}
		}
		return false
	}
	// String answer: substring.
	if as, err := decodeString(answer); err == nil {
		if cs, err := decodeString(compare); err == nil {
			return strings.Contains(as, cs)
		}
	}
	return false
}
