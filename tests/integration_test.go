// Package tests holds in-process integration/E2E tests that hit a real Postgres
// (DATABASE_URL). They skip in -short mode or when DATABASE_URL is unset.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/database"
	"github.com/ahmadasror/txsurvey/internal/handler"
	"github.com/ahmadasror/txsurvey/internal/middleware"
	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/internal/router"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/auth"
)

const testSecret = "integration-secret-at-least-32-characters!!"

// harness bundles everything a test needs to drive the API in-process.
type harness struct {
	t      *testing.T
	pool   *pgxpool.Pool
	engine http.Handler
	jwt    *auth.JWTManager
	userID string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if testing.Short() || dbURL == "" {
		t.Skip("integration test requires DATABASE_URL and no -short")
	}
	if err := database.RunMigrations(dbURL); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	pool, err := database.NewPool(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	// Clean slate (CASCADE clears forms/questions/responses/answers/logic).
	if _, err := pool.Exec(context.Background(), `TRUNCATE users CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	var userID string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO users (google_sub, email, name) VALUES ('sub-1','creator@test.dev','Creator') RETURNING id`,
	).Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cfg := &config.Config{Env: "test", AppBaseURL: "http://localhost", SessionTTL: time.Hour,
		CORSAllowedOrigins: []string{"http://localhost"}}
	jwtMgr := auth.NewJWTManager(testSecret, time.Hour)

	userRepo := repository.NewUserRepo(pool)
	formRepo := repository.NewFormRepo(pool)
	questionRepo := repository.NewQuestionRepo(pool)

	h := &router.Handlers{
		Auth:     handler.NewAuthHandler(service.NewAuthService(cfg, userRepo), jwtMgr, cfg),
		Form:     handler.NewFormHandler(service.NewFormService(formRepo, questionRepo)),
		Question: handler.NewQuestionHandler(service.NewQuestionService(formRepo, questionRepo)),
	}
	return &harness{t: t, pool: pool, engine: router.Setup(cfg, h, jwtMgr), jwt: jwtMgr, userID: userID}
}

// do issues an authenticated JSON request and decodes the {data:...} envelope.
func (h *harness) do(method, path string, body any, out any) int {
	h.t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	tok, _ := h.jwt.GenerateSessionToken(h.userID, "creator@test.dev", "Creator")
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: tok})

	rec := httptest.NewRecorder()
	h.engine.ServeHTTP(rec, req)

	if out != nil && rec.Body.Len() > 0 {
		var env struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err == nil && len(env.Data) > 0 {
			_ = json.Unmarshal(env.Data, out)
		}
	}
	return rec.Code
}

func TestFormAndQuestionFlow(t *testing.T) {
	h := newHarness(t)

	// Create a form.
	var form struct {
		ID     string `json:"id"`
		Slug   string `json:"slug"`
		Status string `json:"status"`
	}
	if code := h.do(http.MethodPost, "/api/v1/forms",
		map[string]string{"title": "Customer Satisfaction"}, &form); code != http.StatusCreated {
		t.Fatalf("create form: want 201, got %d", code)
	}
	if form.ID == "" || form.Slug == "" || form.Status != "draft" {
		t.Fatalf("unexpected form: %+v", form)
	}

	// Publishing an empty form must fail.
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("publish empty: want 422, got %d", code)
	}

	// Add two questions.
	var q1, q2 struct {
		ID       string `json:"id"`
		Position int    `json:"position"`
	}
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Your name?", "required": true}, &q1)
	h.mustCreateQuestion(form.ID, map[string]any{
		"type": "multiple_choice", "title": "Plan?",
		"metadata": map[string]any{"options": []map[string]string{{"label": "Free"}, {"label": "Pro"}}},
	}, &q2)
	if q1.Position != 0 || q2.Position != 1 {
		t.Fatalf("positions: q1=%d q2=%d", q1.Position, q2.Position)
	}

	// Reorder: q2 first.
	if code := h.do(http.MethodPut, "/api/v1/forms/"+form.ID+"/questions/reorder",
		map[string]any{"ordered_ids": []string{q2.ID, q1.ID}}, nil); code != http.StatusOK {
		t.Fatalf("reorder: want 200, got %d", code)
	}

	// Publish now succeeds.
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: want 200, got %d", code)
	}

	// Get detail: published, 2 questions, reordered.
	var detail struct {
		Status    string `json:"status"`
		Questions []struct {
			ID       string `json:"id"`
			Position int    `json:"position"`
		} `json:"questions"`
	}
	if code := h.do(http.MethodGet, "/api/v1/forms/"+form.ID, nil, &detail); code != http.StatusOK {
		t.Fatalf("get: want 200, got %d", code)
	}
	if detail.Status != "published" || len(detail.Questions) != 2 {
		t.Fatalf("unexpected detail: %+v", detail)
	}
	if detail.Questions[0].ID != q2.ID {
		t.Fatalf("reorder not reflected: first=%s want %s", detail.Questions[0].ID, q2.ID)
	}

	// List shows question_count = 2.
	var list []struct {
		ID            string `json:"id"`
		QuestionCount int    `json:"question_count"`
	}
	if code := h.do(http.MethodGet, "/api/v1/forms", nil, &list); code != http.StatusOK {
		t.Fatalf("list: want 200, got %d", code)
	}
	if len(list) != 1 || list[0].QuestionCount != 2 {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func (h *harness) mustCreateQuestion(formID string, body map[string]any, out any) {
	h.t.Helper()
	if code := h.do(http.MethodPost, "/api/v1/forms/"+formID+"/questions", body, out); code != http.StatusCreated {
		h.t.Fatalf("create question %v: want 201, got %d", body["type"], code)
	}
}
