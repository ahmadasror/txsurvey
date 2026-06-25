package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// ResponseService serves the public runner: form fetch by slug and submission
// with server-side validation (required + per-type). Phase 3 is LINEAR — every
// answerable question is reachable; Phase 5 makes reachability logic-aware.
type ResponseService struct {
	forms     *repository.FormRepo
	questions *repository.QuestionRepo
	responses *repository.ResponseRepo
	logic     *repository.LogicRepo
}

func NewResponseService(forms *repository.FormRepo, questions *repository.QuestionRepo, responses *repository.ResponseRepo, logic *repository.LogicRepo) *ResponseService {
	return &ResponseService{forms: forms, questions: questions, responses: responses, logic: logic}
}

// GetPublicForm returns the runner contract for a published form by slug.
func (s *ResponseService) GetPublicForm(ctx context.Context, slug string) (*dto.PublicForm, error) {
	form, questions, rules, err := s.loadPublished(ctx, slug)
	if err != nil {
		return nil, err
	}
	if rules == nil {
		rules = []model.LogicRule{}
	}
	return &dto.PublicForm{
		ID:          form.ID,
		Title:       form.Title,
		Description: form.Description,
		Slug:        form.Slug,
		Settings:    form.Settings,
		Questions:   questions,
		LogicRules:  rules,
	}, nil
}

// Submit validates and persists a completed submission. Returns the new
// response id. Validation is LOGIC-AWARE: the server replays the reachable path
// the submitted answers imply, enforces required-ness only on reachable
// questions, and rejects answers to questions that weren't reached (anti-tamper).
func (s *ResponseService) Submit(ctx context.Context, slug string, req dto.SubmitResponseRequest, meta model.ResponseMeta) (string, error) {
	form, questions, rules, err := s.loadPublished(ctx, slug)
	if err != nil {
		return "", err
	}

	byID := make(map[string]model.Question, len(questions))
	for _, q := range questions {
		byID[q.ID] = q
	}

	// De-dupe submitted answers (last value wins) and reject unknown questions.
	submitted := make(map[string]json.RawMessage, len(req.Answers))
	for _, a := range req.Answers {
		if _, ok := byID[a.QuestionID]; !ok {
			return "", answerErr("answer references a question not in this form")
		}
		submitted[a.QuestionID] = a.Value
	}

	reachable := make(map[string]bool)
	for _, id := range reachablePath(questions, rules, submitted) {
		reachable[id] = true
	}

	stored, err := s.validateSubmission(questions, submitted, reachable)
	if err != nil {
		return "", err
	}

	return s.responses.Insert(ctx, form.ID, true, meta, stored)
}

// validateSubmission walks the form's questions in order. A reachable +
// answerable question that is required must have a non-empty answer; an answer
// to an unreachable question is rejected.
func (s *ResponseService) validateSubmission(questions []model.Question, submitted map[string]json.RawMessage, reachable map[string]bool) ([]model.Answer, error) {
	stored := make([]model.Answer, 0, len(submitted))
	for _, q := range questions {
		raw, present := submitted[q.ID]

		if q.Type == model.QStatement {
			if present && !isJSONEmpty(raw) {
				return nil, answerErr("statement questions cannot be answered")
			}
			continue
		}

		if !reachable[q.ID] {
			if present && !isJSONEmpty(raw) {
				return nil, answerErr("answer for a question that wasn't reached on this path")
			}
			continue
		}

		if !present {
			if q.Required {
				return nil, requiredErr(q)
			}
			continue
		}

		val, empty, err := validateAnswer(q, raw)
		if err != nil {
			return nil, err
		}
		if empty {
			if q.Required {
				return nil, requiredErr(q)
			}
			continue
		}
		stored = append(stored, model.Answer{QuestionID: q.ID, Value: val})
	}
	return stored, nil
}

func (s *ResponseService) loadPublished(ctx context.Context, slug string) (*model.Form, []model.Question, []model.LogicRule, error) {
	form, err := s.forms.GetPublishedBySlug(ctx, slug)
	if err != nil {
		return nil, nil, nil, err
	}
	if form == nil {
		return nil, nil, nil, apperror.New(http.StatusNotFound, "FORM_NOT_FOUND", "form not found or not published")
	}
	questions, err := s.questions.ListByForm(ctx, form.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	rules, err := s.logic.ListByForm(ctx, form.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return form, questions, rules, nil
}

func requiredErr(q model.Question) error {
	title := q.Title
	if title == "" {
		title = "This question"
	}
	return apperror.New(http.StatusUnprocessableEntity, "REQUIRED", title+" is required")
}
