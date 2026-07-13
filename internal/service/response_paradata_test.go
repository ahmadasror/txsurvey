package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// fakeForms / fakeResponses are DB-free stand-ins for the runner's repo slices,
// exercising StartSession/UpdateProgress without Postgres.
type fakeForms struct {
	form *model.Form
	err  error
}

func (f fakeForms) GetPublishedBySlug(_ context.Context, _ string) (*model.Form, error) {
	return f.form, f.err
}

type fakeResponses struct {
	startID  string
	startErr error

	advMatched bool
	advExists  bool
	advErr     error
	gotPos     int
}

func (f *fakeResponses) Insert(context.Context, string, bool, model.ResponseMeta, []model.Answer) (string, error) {
	return "", nil
}
func (f *fakeResponses) StartSession(context.Context, string, model.ResponseMeta) (string, error) {
	return f.startID, f.startErr
}
func (f *fakeResponses) AdvanceProgress(_ context.Context, _ string, position int) (bool, bool, error) {
	f.gotPos = position
	return f.advMatched, f.advExists, f.advErr
}
func (f *fakeResponses) FinalizeSession(context.Context, string, string, model.ResponseMeta, []model.Answer) (bool, error) {
	return false, nil
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	var ce *apperror.ClientError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *apperror.ClientError, got %v", err)
	}
	return ce.Code
}

func TestStartSession(t *testing.T) {
	t.Run("published form -> returns id", func(t *testing.T) {
		svc := &ResponseService{
			forms:     fakeForms{form: &model.Form{ID: "form-1"}},
			responses: &fakeResponses{startID: "resp-1"},
		}
		id, err := svc.StartSession(context.Background(), "slug", model.ResponseMeta{})
		if err != nil || id != "resp-1" {
			t.Fatalf("StartSession = (%q, %v), want (resp-1, nil)", id, err)
		}
	})

	t.Run("unpublished/unknown slug -> 404 FORM_NOT_FOUND", func(t *testing.T) {
		svc := &ResponseService{forms: fakeForms{form: nil}, responses: &fakeResponses{}}
		_, err := svc.StartSession(context.Background(), "nope", model.ResponseMeta{})
		if got := codeOf(t, err); got != "FORM_NOT_FOUND" {
			t.Fatalf("code = %q, want FORM_NOT_FOUND", got)
		}
	})

	t.Run("repo error propagates", func(t *testing.T) {
		boom := errors.New("db down")
		svc := &ResponseService{forms: fakeForms{err: boom}, responses: &fakeResponses{}}
		if _, err := svc.StartSession(context.Background(), "slug", model.ResponseMeta{}); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}

func TestUpdateProgress(t *testing.T) {
	t.Run("matched in-progress row -> ok", func(t *testing.T) {
		svc := &ResponseService{responses: &fakeResponses{advMatched: true, advExists: true}}
		if err := svc.UpdateProgress(context.Background(), "resp-1", 3); err != nil {
			t.Fatalf("UpdateProgress = %v, want nil", err)
		}
	})

	t.Run("already completed -> silent no-op (no error)", func(t *testing.T) {
		svc := &ResponseService{responses: &fakeResponses{advMatched: false, advExists: true}}
		if err := svc.UpdateProgress(context.Background(), "resp-1", 3); err != nil {
			t.Fatalf("UpdateProgress = %v, want nil (no-op)", err)
		}
	})

	t.Run("unknown/malformed id -> 404 RESPONSE_NOT_FOUND", func(t *testing.T) {
		svc := &ResponseService{responses: &fakeResponses{advMatched: false, advExists: false}}
		err := svc.UpdateProgress(context.Background(), "ghost", 1)
		if got := codeOf(t, err); got != "RESPONSE_NOT_FOUND" {
			t.Fatalf("code = %q, want RESPONSE_NOT_FOUND", got)
		}
	})

	t.Run("negative position clamps to 0", func(t *testing.T) {
		fr := &fakeResponses{advMatched: true, advExists: true}
		svc := &ResponseService{responses: fr}
		if err := svc.UpdateProgress(context.Background(), "resp-1", -5); err != nil {
			t.Fatalf("UpdateProgress = %v", err)
		}
		if fr.gotPos != 0 {
			t.Fatalf("clamped position = %d, want 0", fr.gotPos)
		}
	})

	t.Run("repo error propagates", func(t *testing.T) {
		boom := errors.New("db down")
		svc := &ResponseService{responses: &fakeResponses{advErr: boom}}
		if err := svc.UpdateProgress(context.Background(), "resp-1", 1); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}
