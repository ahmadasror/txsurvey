package service

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// assetURLRe matches the relative asset path the upload endpoint hands back
// ("uploads/<random>.<ext>"). Banner/logo settings must be either empty or this
// exact shape — never an arbitrary URL (open-redirect / SSRF / stored-XSS risk).
var assetURLRe = regexp.MustCompile(`^uploads/[A-Za-z0-9._-]+$`)

// FormService owns the form lifecycle: creation (with unique slug), editing,
// publish/unpublish (with validation), and detail assembly.
type FormService struct {
	forms     *repository.FormRepo
	questions *repository.QuestionRepo
	logic     *repository.LogicRepo
}

func NewFormService(forms *repository.FormRepo, questions *repository.QuestionRepo, logic *repository.LogicRepo) *FormService {
	return &FormService{forms: forms, questions: questions, logic: logic}
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
	rules, err := s.logic.ListByForm(ctx, form.ID)
	if err != nil {
		return nil, err
	}
	form.LogicRules = rules
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

// Update edits a form's title/description/settings, and optionally its slug
// (public URL). The slug can only change while the form is a draft — once
// published its URL is frozen so links already shared with respondents keep
// working.
func (s *FormService) Update(ctx context.Context, ownerID, id string, req dto.UpdateFormRequest) (*model.Form, error) {
	if err := validateFormSettings(req.Settings); err != nil {
		return nil, err
	}
	current, err := s.forms.GetByIDOwned(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errFormNotFound
	}
	slug, err := s.resolveSlug(ctx, current, req.Slug)
	if err != nil {
		return nil, err
	}
	form, err := s.forms.UpdateMeta(ctx, ownerID, id, req.Title, req.Description, slug, req.Settings)
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

// resolveSlug decides the slug to persist on update. An empty or unchanged
// requested slug keeps the current one. A changed slug is normalized, and is
// only accepted when the form is still a draft and the target is free.
func (s *FormService) resolveSlug(ctx context.Context, current *model.Form, requested string) (string, error) {
	if requested == "" {
		return current.Slug, nil
	}
	cand := slugify(requested)
	if cand == current.Slug {
		return current.Slug, nil
	}
	if current.Status == model.FormPublished {
		return "", apperror.New(http.StatusUnprocessableEntity, "SLUG_LOCKED",
			"the URL can only be changed while the form is a draft")
	}
	exists, err := s.forms.SlugExists(ctx, cand)
	if err != nil {
		return "", err
	}
	if exists {
		return "", apperror.New(http.StatusUnprocessableEntity, "SLUG_TAKEN",
			"that URL is already taken — pick another")
	}
	return cand, nil
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

// validateFormSettings guards the URL-bearing settings fields. RedirectURL must
// be a real http(s) URL (no javascript:/data: — open-redirect/XSS); banner/logo
// must be the relative asset path our own upload produces.
func validateFormSettings(s model.FormSettings) error {
	if s.RedirectURL != "" {
		u, err := url.Parse(s.RedirectURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return apperror.New(http.StatusUnprocessableEntity, "BAD_REDIRECT_URL",
				"redirect URL must be an http(s) URL")
		}
	}
	for _, v := range []string{s.BannerURL, s.LogoURL} {
		if v != "" && !assetURLRe.MatchString(v) {
			return apperror.New(http.StatusUnprocessableEntity, "BAD_ASSET_URL",
				"banner/logo must be an uploaded asset path")
		}
	}
	return nil
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
