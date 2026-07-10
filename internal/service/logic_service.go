package service

import (
	"context"
	"net/http"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// LogicService manages a form's logic rules with ownership + graph validation.
type LogicService struct {
	forms     *repository.FormRepo
	questions *repository.QuestionRepo
	logic     *repository.LogicRepo
}

func NewLogicService(forms *repository.FormRepo, questions *repository.QuestionRepo, logic *repository.LogicRepo) *LogicService {
	return &LogicService{forms: forms, questions: questions, logic: logic}
}

func (s *LogicService) requireForm(ctx context.Context, ownerID, formID string) error {
	form, err := s.forms.GetByIDOwned(ctx, formID, ownerID)
	if err != nil {
		return err
	}
	if form == nil {
		return errFormNotFound
	}
	return nil
}

// List returns a form's rules.
func (s *LogicService) List(ctx context.Context, ownerID, formID string) ([]model.LogicRule, error) {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	return s.logic.ListByForm(ctx, formID)
}

// Create validates and inserts a rule.
func (s *LogicService) Create(ctx context.Context, ownerID, formID string, in dto.LogicRuleInput) (*model.LogicRule, error) {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	questions, err := s.questions.ListByForm(ctx, formID)
	if err != nil {
		return nil, err
	}
	if err := validateRule(in, questions); err != nil {
		return nil, err
	}
	rule := &model.LogicRule{
		FormID:           formID,
		SourceQuestionID: in.SourceQuestionID,
		Operator:         in.Operator,
		CompareValue:     in.CompareValue,
		Action:           in.Action,
		TargetQuestionID: in.TargetQuestionID,
		Priority:         in.Priority,
	}
	if err := s.logic.Insert(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// Update validates and mutates a rule.
func (s *LogicService) Update(ctx context.Context, ownerID, formID, ruleID string, in dto.LogicRuleInput) (*model.LogicRule, error) {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	questions, err := s.questions.ListByForm(ctx, formID)
	if err != nil {
		return nil, err
	}
	if err := validateRule(in, questions); err != nil {
		return nil, err
	}
	rule := &model.LogicRule{
		Operator:         in.Operator,
		CompareValue:     in.CompareValue,
		Action:           in.Action,
		TargetQuestionID: in.TargetQuestionID,
		Priority:         in.Priority,
	}
	updated, err := s.logic.Update(ctx, formID, ruleID, rule)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperror.New(http.StatusNotFound, "RULE_NOT_FOUND", "logic rule not found")
	}
	return updated, nil
}

// Delete removes a rule.
func (s *LogicService) Delete(ctx context.Context, ownerID, formID, ruleID string) error {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return err
	}
	ok, err := s.logic.Delete(ctx, formID, ruleID)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.New(http.StatusNotFound, "RULE_NOT_FOUND", "logic rule not found")
	}
	return nil
}

// validateRule enforces the rule's structural invariants. Forward-only jumps
// keep the navigation graph a DAG (no infinite loops); the runtime engine also
// caps steps as a belt-and-suspenders against post-reorder backward jumps.
func validateRule(in dto.LogicRuleInput, questions []model.Question) error {
	byID := make(map[string]model.Question, len(questions))
	for _, q := range questions {
		byID[q.ID] = q
	}

	source, ok := byID[in.SourceQuestionID]
	if !ok {
		return logicErr("source question is not in this form")
	}
	if !model.ValidLogicOperator(in.Operator) {
		return logicErr("unknown operator")
	}
	if !model.ValidLogicAction(in.Action) {
		return logicErr("unknown action")
	}
	needsValue := in.Operator != model.OpIsEmpty && in.Operator != model.OpIsNotEmpty && in.Operator != model.OpAlways
	if needsValue && len(in.CompareValue) == 0 {
		return logicErr("this operator requires a comparison value")
	}

	switch in.Action {
	case model.ActionEndForm:
		if in.TargetQuestionID != nil {
			return logicErr("end_form must not have a target")
		}
	case model.ActionJumpTo, model.ActionShow, model.ActionHide:
		if in.TargetQuestionID == nil {
			return logicErr("this action requires a target question")
		}
		target, ok := byID[*in.TargetQuestionID]
		if !ok {
			return logicErr("target question is not in this form")
		}
		if target.ID == source.ID {
			return logicErr("a rule cannot target its own question")
		}
		if in.Action == model.ActionJumpTo && target.Position <= source.Position {
			return logicErr("jump_to must target a later question (forward only)")
		}
	}
	return nil
}

func logicErr(msg string) error {
	return apperror.New(http.StatusUnprocessableEntity, "INVALID_LOGIC_RULE", msg)
}
