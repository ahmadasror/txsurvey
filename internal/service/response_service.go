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

// publicFormStore / responseWriter are the narrow slices of the form/response
// repos the runner needs — interfaces so StartSession/UpdateProgress are unit-
// testable without a DB (the concrete repos satisfy them, so wiring is unchanged).
type publicFormStore interface {
	GetPublishedBySlug(ctx context.Context, slug string) (*model.Form, error)
}

type responseWriter interface {
	Insert(ctx context.Context, formID string, completed bool, meta model.ResponseMeta, answers []model.Answer) (string, error)
	StartSession(ctx context.Context, formID string, meta model.ResponseMeta) (string, error)
	AdvanceProgress(ctx context.Context, responseID string, position int) (matched, exists bool, err error)
	FinalizeSession(ctx context.Context, responseID, formID string, meta model.ResponseMeta, answers []model.Answer) (bool, error)
}

// ResponseService serves the public runner: form fetch by slug and submission
// with server-side validation (required + per-type). Phase 3 is LINEAR — every
// answerable question is reachable; Phase 5 makes reachability logic-aware.
type ResponseService struct {
	forms     publicFormStore
	questions *repository.QuestionRepo
	responses responseWriter
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

	// If the runner opened a paradata session, finalize that in-progress row so
	// the completed respondent leaves ONE row (not a completed row + an orphaned
	// in-progress ghost that would skew the drop-off funnel). A mismatched/stale
	// id finalizes nothing → fall back to inserting a fresh completed row.
	if req.ResponseID != "" {
		finalized, err := s.responses.FinalizeSession(ctx, req.ResponseID, form.ID, meta, stored)
		if err != nil {
			return "", err
		}
		if finalized {
			return req.ResponseID, nil
		}
	}

	return s.responses.Insert(ctx, form.ID, true, meta, stored)
}

// StartSession opens an in-progress response for a published form (paradata
// capture) and returns its id. The runner calls this on load and echoes the id
// to UpdateProgress as the respondent advances. The row is completed=false, so
// it is invisible to every owner-facing surface until a future funnel view.
func (s *ResponseService) StartSession(ctx context.Context, slug string, meta model.ResponseMeta) (string, error) {
	form, err := s.forms.GetPublishedBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	if form == nil {
		return "", apperror.New(http.StatusNotFound, "FORM_NOT_FOUND", "form not found or not published")
	}
	return s.responses.StartSession(ctx, form.ID, meta)
}

// UpdateProgress advances an in-progress response's furthest-reached position.
// Best-effort telemetry: a ping to an already-completed response is a silent
// no-op (a final ping legitimately races Submit); a ping to a non-existent or
// malformed id is 404. A negative position clamps to 0.
func (s *ResponseService) UpdateProgress(ctx context.Context, responseID string, position int) error {
	if position < 0 {
		position = 0
	}
	_, exists, err := s.responses.AdvanceProgress(ctx, responseID, position)
	if err != nil {
		return err
	}
	if !exists {
		return apperror.New(http.StatusNotFound, "RESPONSE_NOT_FOUND", "response not found")
	}
	return nil
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
