package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// errFormNotFound is returned when a form is absent or not owned by the caller.
var errFormNotFound = apperror.New(http.StatusNotFound, "FORM_NOT_FOUND", "form not found")

// FormStore is the form-ownership dependency of question/logic services.
type FormStore interface {
	GetByIDOwned(ctx context.Context, id, ownerID string) (*model.Form, error)
}

// QuestionStore is the question persistence dependency.
type QuestionStore interface {
	ListByForm(ctx context.Context, formID string) ([]model.Question, error)
	Insert(ctx context.Context, q *model.Question) error
	Update(ctx context.Context, formID, id string, q *model.Question) (*model.Question, error)
	Delete(ctx context.Context, formID, id string) (bool, error)
	Reorder(ctx context.Context, formID string, orderedIDs []string) error
}

// QuestionService manages a form's questions, enforcing ownership and per-type
// metadata validation.
type QuestionService struct {
	forms     FormStore
	questions QuestionStore
}

func NewQuestionService(forms FormStore, questions QuestionStore) *QuestionService {
	return &QuestionService{forms: forms, questions: questions}
}

func (s *QuestionService) requireForm(ctx context.Context, ownerID, formID string) error {
	form, err := s.forms.GetByIDOwned(ctx, formID, ownerID)
	if err != nil {
		return err
	}
	if form == nil {
		return errFormNotFound
	}
	return nil
}

// Add validates and appends a question to an owned form.
func (s *QuestionService) Add(ctx context.Context, ownerID, formID string, in dto.QuestionInput) (*model.Question, error) {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	q := &model.Question{
		FormID:      formID,
		Type:        in.Type,
		Title:       in.Title,
		Description: in.Description,
		Required:    in.Required,
		Metadata:    in.Metadata,
	}
	if err := validateQuestion(q); err != nil {
		return nil, err
	}
	if err := s.questions.Insert(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

// Update validates and mutates an existing question of an owned form.
func (s *QuestionService) Update(ctx context.Context, ownerID, formID, qid string, in dto.QuestionInput) (*model.Question, error) {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	q := &model.Question{
		Type:        in.Type,
		Title:       in.Title,
		Description: in.Description,
		Required:    in.Required,
		Metadata:    in.Metadata,
	}
	if err := validateQuestion(q); err != nil {
		return nil, err
	}
	updated, err := s.questions.Update(ctx, formID, qid, q)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperror.New(http.StatusNotFound, "QUESTION_NOT_FOUND", "question not found")
	}
	return updated, nil
}

// Delete removes a question from an owned form.
func (s *QuestionService) Delete(ctx context.Context, ownerID, formID, qid string) error {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return err
	}
	ok, err := s.questions.Delete(ctx, formID, qid)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.New(http.StatusNotFound, "QUESTION_NOT_FOUND", "question not found")
	}
	return nil
}

// Reorder sets question positions to match orderedIDs, which must be a
// permutation of the form's questions.
func (s *QuestionService) Reorder(ctx context.Context, ownerID, formID string, orderedIDs []string) error {
	if err := s.requireForm(ctx, ownerID, formID); err != nil {
		return err
	}
	existing, err := s.questions.ListByForm(ctx, formID)
	if err != nil {
		return err
	}
	if len(orderedIDs) != len(existing) {
		return apperror.New(http.StatusUnprocessableEntity, "REORDER_MISMATCH",
			"ordered_ids must list every question exactly once")
	}
	seen := make(map[string]bool, len(orderedIDs))
	for _, id := range orderedIDs {
		if seen[id] {
			return apperror.New(http.StatusUnprocessableEntity, "REORDER_DUPLICATE", "duplicate question id in reorder")
		}
		seen[id] = true
	}
	return s.questions.Reorder(ctx, formID, orderedIDs)
}

// validateQuestion normalizes and validates per-type metadata. It mutates q in
// place (filling defaults, generating option ids, clearing irrelevant fields)
// and returns a ClientError on invalid input. Pure aside from q — unit-tested.
func validateQuestion(q *model.Question) error {
	if !model.ValidQuestionType(q.Type) {
		return apperror.New(http.StatusUnprocessableEntity, "BAD_QUESTION_TYPE", "unknown question type")
	}
	if strings.TrimSpace(q.Title) == "" {
		return apperror.New(http.StatusUnprocessableEntity, "QUESTION_TITLE", "question title is required")
	}

	switch {
	case q.Type.IsChoice():
		if len(q.Metadata.Options) == 0 {
			return apperror.New(http.StatusUnprocessableEntity, "QUESTION_OPTIONS", "choice question needs at least one option")
		}
		seen := make(map[string]bool, len(q.Metadata.Options))
		for i := range q.Metadata.Options {
			opt := &q.Metadata.Options[i]
			opt.Label = strings.TrimSpace(opt.Label)
			if opt.Label == "" {
				return apperror.New(http.StatusUnprocessableEntity, "QUESTION_OPTIONS", "option label is required")
			}
			if opt.ID == "" {
				opt.ID = "opt_" + randSuffix(8)
			}
			if seen[opt.ID] {
				return apperror.New(http.StatusUnprocessableEntity, "QUESTION_OPTIONS", "duplicate option id")
			}
			seen[opt.ID] = true
		}
		q.Metadata.Scale = 0

	case q.Type == model.QRating:
		if q.Metadata.Scale == 0 {
			q.Metadata.Scale = 5
		}
		if q.Metadata.Scale < 2 || q.Metadata.Scale > 10 {
			return apperror.New(http.StatusUnprocessableEntity, "QUESTION_SCALE", "rating scale must be between 2 and 10")
		}
		q.Metadata.Options = nil

	case q.Type == model.QNumber:
		if q.Metadata.Min != nil && q.Metadata.Max != nil && *q.Metadata.Min > *q.Metadata.Max {
			return apperror.New(http.StatusUnprocessableEntity, "QUESTION_RANGE", "min must be <= max")
		}
		q.Metadata.Options = nil

	case q.Type == model.QStatement:
		// Display-only: never answered, never required, no metadata.
		q.Required = false
		q.Metadata = model.QuestionMetadata{}

	default:
		q.Metadata.Options = nil
		q.Metadata.Scale = 0
	}
	return nil
}
