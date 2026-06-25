package service

import (
	"context"
	"net/http"
	"time"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// FormService owns the form lifecycle: creation (with unique slug), editing,
// publish/unpublish (with validation), and detail assembly.
type FormService struct {
	forms     *repository.FormRepo
	questions *repository.QuestionRepo
}

func NewFormService(forms *repository.FormRepo, questions *repository.QuestionRepo) *FormService {
	return &FormService{forms: forms, questions: questions}
}

// Create makes a new draft form with a globally-unique slug.
func (s *FormService) Create(ctx context.Context, ownerID string, req dto.CreateFormRequest) (*model.Form, error) {
	slug, err := s.uniqueSlug(ctx, req.Title)
	if err != nil {
		return nil, err
	}
	f := &model.Form{
		OwnerID:     ownerID,
		Title:       req.Title,
		Description: req.Description,
		Slug:        slug,
		Status:      model.FormDraft,
		Settings:    model.FormSettings{ShowProgress: true},
	}
	if err := s.forms.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

// Get returns an owned form with its questions attached.
func (s *FormService) Get(ctx context.Context, ownerID, id string) (*model.Form, error) {
	form, err := s.forms.GetByIDOwned(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, errFormNotFound
	}
	if err := s.attachQuestions(ctx, form); err != nil {
		return nil, err
	}
	return form, nil
}

// List returns a page of the owner's forms and the total count.
func (s *FormService) List(ctx context.Context, ownerID string, page, perPage int) ([]model.FormListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.forms.ListByOwner(ctx, ownerID, perPage, (page-1)*perPage)
}

// Update edits a form's title/description/settings.
func (s *FormService) Update(ctx context.Context, ownerID, id string, req dto.UpdateFormRequest) (*model.Form, error) {
	form, err := s.forms.UpdateMeta(ctx, ownerID, id, req.Title, req.Description, req.Settings)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, errFormNotFound
	}
	if err := s.attachQuestions(ctx, form); err != nil {
		return nil, err
	}
	return form, nil
}

// Delete soft-deletes a form.
func (s *FormService) Delete(ctx context.Context, ownerID, id string) error {
	ok, err := s.forms.SoftDelete(ctx, ownerID, id)
	if err != nil {
		return err
	}
	if !ok {
		return errFormNotFound
	}
	return nil
}

// Publish validates that the form has at least one answerable question, then
// flips it to published.
func (s *FormService) Publish(ctx context.Context, ownerID, id string) (*model.Form, error) {
	form, err := s.forms.GetByIDOwned(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, errFormNotFound
	}
	questions, err := s.questions.ListByForm(ctx, id)
	if err != nil {
		return nil, err
	}
	if answerableCount(questions) == 0 {
		return nil, apperror.New(http.StatusUnprocessableEntity, "PUBLISH_EMPTY",
			"a form needs at least one answerable question before publishing")
	}
	now := time.Now()
	published, err := s.forms.SetStatus(ctx, ownerID, id, model.FormPublished, &now)
	if err != nil {
		return nil, err
	}
	if published == nil {
		return nil, errFormNotFound
	}
	published.Questions = questions
	return published, nil
}

// Unpublish returns a form to draft (its slug and responses are preserved).
func (s *FormService) Unpublish(ctx context.Context, ownerID, id string) (*model.Form, error) {
	form, err := s.forms.SetStatus(ctx, ownerID, id, model.FormDraft, nil)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, errFormNotFound
	}
	if err := s.attachQuestions(ctx, form); err != nil {
		return nil, err
	}
	return form, nil
}

func (s *FormService) attachQuestions(ctx context.Context, form *model.Form) error {
	questions, err := s.questions.ListByForm(ctx, form.ID)
	if err != nil {
		return err
	}
	form.Questions = questions
	return nil
}

func (s *FormService) uniqueSlug(ctx context.Context, title string) (string, error) {
	base := slugify(title)
	for i := 0; i < 12; i++ {
		cand := base + "-" + randSuffix(6)
		exists, err := s.forms.SlugExists(ctx, cand)
		if err != nil {
			return "", err
		}
		if !exists {
			return cand, nil
		}
	}
	return "", apperror.New(http.StatusInternalServerError, "SLUG_ALLOC", "could not allocate a unique slug")
}

func answerableCount(questions []model.Question) int {
	n := 0
	for _, q := range questions {
		if q.Type != model.QStatement {
			n++
		}
	}
	return n
}
