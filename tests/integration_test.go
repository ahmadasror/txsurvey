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
	"strings"
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
	// SAFETY: these tests TRUNCATE — never let them run against a non-test DB
	// (e.g. the production database). Point DATABASE_URL at a *_test database.
	dbName := dbURL
	if i := strings.LastIndex(dbName, "/"); i >= 0 {
		dbName = dbName[i+1:]
	}
	if i := strings.IndexByte(dbName, '?'); i >= 0 {
		dbName = dbName[:i]
	}
	if !strings.Contains(dbName, "test") {
		t.Skipf("refusing to run TRUNCATE-ing tests against non-test database %q — use a *_test DB", dbName)
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
		CORSAllowedOrigins: []string{"http://localhost"}, UploadDir: t.TempDir()}
	jwtMgr := auth.NewJWTManager(testSecret, time.Hour)

	userRepo := repository.NewUserRepo(pool)
	formRepo := repository.NewFormRepo(pool)
	questionRepo := repository.NewQuestionRepo(pool)
	responseRepo := repository.NewResponseRepo(pool)
	logicRepo := repository.NewLogicRepo(pool)

	h := &router.Handlers{
		Auth:     handler.NewAuthHandler(service.NewAuthService(cfg, userRepo), jwtMgr, cfg),
		Form:     handler.NewFormHandler(service.NewFormService(formRepo, questionRepo, logicRepo)),
		Question: handler.NewQuestionHandler(service.NewQuestionService(formRepo, questionRepo)),
		Public:   handler.NewPublicHandler(service.NewResponseService(formRepo, questionRepo, responseRepo, logicRepo)),
		Results:  handler.NewResultsHandler(service.NewResultsService(formRepo, questionRepo, responseRepo)),
		Logic:    handler.NewLogicHandler(service.NewLogicService(formRepo, questionRepo, logicRepo)),
		Asset:    handler.NewAssetHandler(formRepo, cfg.UploadDir, cfg.StorageLimitBytes),
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

// TestE2EHappyPath is the locked end-to-end flow:
// create -> add questions -> publish -> submit 2 responses ->
// list responses -> analytics -> export.csv.
func TestE2EHappyPath(t *testing.T) {
	h := newHarness(t)

	var form struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "NPS Survey"}, &form)

	var name, rate struct {
		ID string `json:"id"`
	}
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Your name", "required": true}, &name)
	h.mustCreateQuestion(form.ID, map[string]any{"type": "rating", "title": "Rate us", "metadata": map[string]any{"scale": 5}}, &rate)

	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: %d", code)
	}

	// Two submissions: (Alice,5) and (Bob,3) -> avg rating 4.
	for _, sub := range []struct {
		who    string
		rating int
	}{{"Alice", 5}, {"Bob", 3}} {
		body := map[string]any{"answers": []map[string]any{
			{"question_id": name.ID, "value": sub.who},
			{"question_id": rate.ID, "value": sub.rating},
		}}
		if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses", body, nil); code != http.StatusCreated {
			t.Fatalf("submit %s: %d", sub.who, code)
		}
	}

	// List responses -> 2, each with 2 answers.
	var responses []struct {
		ID      string `json:"id"`
		Answers []struct {
			QuestionID string `json:"question_id"`
		} `json:"answers"`
	}
	if code := h.do(http.MethodGet, "/api/v1/forms/"+form.ID+"/responses", nil, &responses); code != http.StatusOK {
		t.Fatalf("list responses: %d", code)
	}
	if len(responses) != 2 {
		t.Fatalf("want 2 responses, got %d", len(responses))
	}
	for _, r := range responses {
		if len(r.Answers) != 2 {
			t.Fatalf("want 2 answers per response, got %d", len(r.Answers))
		}
	}

	// Analytics: 2 responses, completion 1.0, rating average 4.
	var analytics struct {
		ResponseCount  int     `json:"response_count"`
		CompletionRate float64 `json:"completion_rate"`
		Questions      []struct {
			QuestionID string   `json:"question_id"`
			Type       string   `json:"type"`
			Answered   int      `json:"answered"`
			Average    *float64 `json:"average"`
		} `json:"questions"`
	}
	if code := h.do(http.MethodGet, "/api/v1/forms/"+form.ID+"/analytics", nil, &analytics); code != http.StatusOK {
		t.Fatalf("analytics: %d", code)
	}
	if analytics.ResponseCount != 2 || analytics.CompletionRate != 1.0 {
		t.Fatalf("unexpected analytics totals: %+v", analytics)
	}
	var ratingAvg *float64
	for _, q := range analytics.Questions {
		if q.QuestionID == rate.ID {
			ratingAvg = q.Average
		}
	}
	if ratingAvg == nil || *ratingAvg != 4.0 {
		t.Fatalf("rating average = %v, want 4.0", ratingAvg)
	}

	// CSV export: header + 2 data rows, with the question titles as columns.
	code, csvBody := h.doAuthRaw(http.MethodGet, "/api/v1/forms/"+form.ID+"/export.csv", nil)
	if code != http.StatusOK {
		t.Fatalf("export: %d", code)
	}
	lines := strings.Split(strings.TrimSpace(csvBody), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 CSV lines (header+2), got %d: %q", len(lines), csvBody)
	}
	if !strings.Contains(lines[0], "Your name") || !strings.Contains(lines[0], "Rate us") {
		t.Fatalf("CSV header missing question columns: %q", lines[0])
	}
	if !strings.Contains(csvBody, "Alice") || !strings.Contains(csvBody, "Bob") {
		t.Fatalf("CSV missing answer values: %q", csvBody)
	}
}

// TestBranchingSubmission exercises Phase 5: a jump rule skips q2 when q1=yes,
// and the server validates required-ness against the reachable path + rejects
// answers to skipped questions.
func TestBranchingSubmission(t *testing.T) {
	h := newHarness(t)

	var form struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "Branching"}, &form)

	var q1, q2, q3 struct {
		ID string `json:"id"`
	}
	h.mustCreateQuestion(form.ID, map[string]any{"type": "yes_no", "title": "Skip the next one?"}, &q1)
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Why are you here?", "required": true}, &q2)
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Anything else?"}, &q3)

	// Rule: if q1 == true, jump to q3 (skipping the required q2).
	rule := map[string]any{
		"source_question_id": q1.ID,
		"operator":           "eq",
		"compare_value":      true,
		"action":             "jump_to",
		"target_question_id": q3.ID,
	}
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/logic", rule, nil); code != http.StatusCreated {
		t.Fatalf("create rule: want 201, got %d", code)
	}

	// A backward jump must be rejected.
	bad := map[string]any{"source_question_id": q3.ID, "operator": "eq", "compare_value": "x", "action": "jump_to", "target_question_id": q1.ID}
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/logic", bad, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("backward jump: want 422, got %d", code)
	}

	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: %d", code)
	}

	// Public contract carries the logic rule.
	var public struct {
		LogicRules []struct {
			Action string `json:"action"`
		} `json:"logic_rules"`
	}
	h.doAnon(http.MethodGet, "/api/v1/public/forms/"+form.Slug, nil, &public)
	if len(public.LogicRules) != 1 || public.LogicRules[0].Action != "jump_to" {
		t.Fatalf("public logic rules: %+v", public.LogicRules)
	}

	submit := func(answers []map[string]any) int {
		code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses",
			map[string]any{"answers": answers}, nil)
		return code
	}

	// q1=yes -> q2 skipped; answering only q1 & q3 is valid.
	if code := submit([]map[string]any{{"question_id": q1.ID, "value": true}, {"question_id": q3.ID, "value": "bye"}}); code != http.StatusCreated {
		t.Fatalf("branch-taken valid: want 201, got %d", code)
	}

	// q1=yes but also answering the skipped q2 -> rejected (unreachable).
	if code := submit([]map[string]any{{"question_id": q1.ID, "value": true}, {"question_id": q2.ID, "value": "sneaky"}}); code != http.StatusUnprocessableEntity {
		t.Fatalf("answer to skipped question: want 422, got %d", code)
	}

	// q1=no -> q2 is on the path and required; omitting it fails.
	if code := submit([]map[string]any{{"question_id": q1.ID, "value": false}}); code != http.StatusUnprocessableEntity {
		t.Fatalf("missing required on path: want 422, got %d", code)
	}

	// q1=no with q2 answered -> valid.
	if code := submit([]map[string]any{{"question_id": q1.ID, "value": false}, {"question_id": q2.ID, "value": "curious"}}); code != http.StatusCreated {
		t.Fatalf("branch-not-taken valid: want 201, got %d", code)
	}
}

// TestUnconditionalJump exercises the 'always' operator (migration 006): an
// unconditional jump routes q1 straight to q3, skipping q2 for every respondent
// regardless of q1's answer — the "Lompat langsung" builder feature.
func TestUnconditionalJump(t *testing.T) {
	h := newHarness(t)

	var form struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "Direct"}, &form)

	var q1, q2, q3 struct {
		ID string `json:"id"`
	}
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "First"}, &q1)
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Skipped", "required": true}, &q2)
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Last"}, &q3)

	// Unconditional jump: from q1, always go to q3 (no compare_value).
	rule := map[string]any{
		"source_question_id": q1.ID,
		"operator":           "always",
		"action":             "jump_to",
		"target_question_id": q3.ID,
	}
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/logic", rule, nil); code != http.StatusCreated {
		t.Fatalf("create always rule: want 201, got %d", code)
	}
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: %d", code)
	}

	submit := func(answers []map[string]any) int {
		code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses",
			map[string]any{"answers": answers}, nil)
		return code
	}

	// q2 is skipped for everyone: answering only q1 & q3 is valid even though q2 is required.
	if code := submit([]map[string]any{{"question_id": q1.ID, "value": "hi"}, {"question_id": q3.ID, "value": "bye"}}); code != http.StatusCreated {
		t.Fatalf("always-jump path: want 201, got %d", code)
	}
	// Answering the skipped q2 is rejected as unreachable.
	if code := submit([]map[string]any{{"question_id": q1.ID, "value": "hi"}, {"question_id": q2.ID, "value": "sneaky"}}); code != http.StatusUnprocessableEntity {
		t.Fatalf("answer to skipped question: want 422, got %d", code)
	}
}

// TestSlugEdit exercises the custom-slug feature: a draft form's URL can be
// retargeted, must stay unique, and is frozen once the form is published.
func TestSlugEdit(t *testing.T) {
	h := newHarness(t)

	var form struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "Draft One"}, &form)

	// Draft: retarget the slug (normalized from a messy input).
	var updated struct {
		Slug string `json:"slug"`
	}
	if code := h.do(http.MethodPatch, "/api/v1/forms/"+form.ID,
		map[string]any{"title": "Draft One", "slug": "Sprint Review!! Prospera"}, &updated); code != http.StatusOK {
		t.Fatalf("edit slug: want 200, got %d", code)
	}
	if updated.Slug != "sprint-review-prospera" {
		t.Fatalf("normalized slug = %q, want sprint-review-prospera", updated.Slug)
	}

	// A second form cannot claim the same slug.
	var form2 struct {
		ID string `json:"id"`
	}
	h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "Draft Two"}, &form2)
	if code := h.do(http.MethodPatch, "/api/v1/forms/"+form2.ID,
		map[string]any{"title": "Draft Two", "slug": "sprint-review-prospera"}, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("duplicate slug: want 422, got %d", code)
	}

	// Publish form one, then the slug is locked.
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Q"}, &struct {
		ID string `json:"id"`
	}{})
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: %d", code)
	}
	if code := h.do(http.MethodPatch, "/api/v1/forms/"+form.ID,
		map[string]any{"title": "Draft One", "slug": "yet-another-url"}, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("locked slug: want 422, got %d", code)
	}
	// Unchanged slug on a published form is still fine (title-only edit).
	if code := h.do(http.MethodPatch, "/api/v1/forms/"+form.ID,
		map[string]any{"title": "Draft One Renamed", "slug": "sprint-review-prospera"}, nil); code != http.StatusOK {
		t.Fatalf("published title edit: want 200, got %d", code)
	}
}

func (h *harness) mustCreateQuestion(formID string, body map[string]any, out any) {
	h.t.Helper()
	if code := h.do(http.MethodPost, "/api/v1/forms/"+formID+"/questions", body, out); code != http.StatusCreated {
		h.t.Fatalf("create question %v: want 201, got %d", body["type"], code)
	}
}

// doAuthRaw issues an authenticated request and returns status + raw body
// (used for non-JSON responses like CSV export).
func (h *harness) doAuthRaw(method, path string, body any) (int, string) {
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
	return rec.Code, rec.Body.String()
}

// doAnon issues an UNauthenticated request (no session cookie) — for the public
// runner endpoints — and returns the status plus the raw body.
func (h *harness) doAnon(method, path string, body any, out any) (int, string) {
	h.t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
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
	return rec.Code, rec.Body.String()
}

func TestPublicRunnerFlow(t *testing.T) {
	h := newHarness(t)

	var form struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	if code := h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "Feedback"}, &form); code != http.StatusCreated {
		t.Fatalf("create form: %d", code)
	}

	var q1, q2 struct {
		ID string `json:"id"`
	}
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Name?", "required": true}, &q1)
	h.mustCreateQuestion(form.ID, map[string]any{"type": "rating", "title": "Rate us", "metadata": map[string]any{"scale": 5}}, &q2)

	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: %d", code)
	}

	// Anonymous fetch by slug.
	var public struct {
		Questions []struct {
			ID string `json:"id"`
		} `json:"questions"`
	}
	code, bodyStr := h.doAnon(http.MethodGet, "/api/v1/public/forms/"+form.Slug, nil, &public)
	if code != http.StatusOK {
		t.Fatalf("public get: want 200, got %d", code)
	}
	if len(public.Questions) != 2 {
		t.Fatalf("want 2 questions, got %d", len(public.Questions))
	}
	if strings.Contains(bodyStr, "owner_id") {
		t.Fatalf("public form leaked owner_id: %s", bodyStr)
	}

	// Valid submission.
	valid := map[string]any{"answers": []map[string]any{
		{"question_id": q1.ID, "value": "Alice"},
		{"question_id": q2.ID, "value": 4},
	}}
	if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses", valid, nil); code != http.StatusCreated {
		t.Fatalf("submit valid: want 201, got %d", code)
	}

	// Missing the required name -> 422.
	missing := map[string]any{"answers": []map[string]any{{"question_id": q2.ID, "value": 3}}}
	if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses", missing, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("submit missing required: want 422, got %d", code)
	}

	// Answer to an unknown question -> 422.
	unknown := map[string]any{"answers": []map[string]any{
		{"question_id": q1.ID, "value": "Bob"},
		{"question_id": "00000000-0000-0000-0000-000000000000", "value": "x"},
	}}
	if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses", unknown, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("submit unknown question: want 422, got %d", code)
	}

	// Submitting to an unpublished/unknown slug -> 404.
	if code, _ := h.doAnon(http.MethodGet, "/api/v1/public/forms/does-not-exist", nil, nil); code != http.StatusNotFound {
		t.Fatalf("unknown slug: want 404, got %d", code)
	}
}

// TestParadataInvisibleToOwner is the load-bearing check for FR-RUN-003: an
// in-progress (started but not submitted) response must be INERT — captured as
// paradata but invisible to every owner-facing surface (count, list, analytics)
// until a real submission lands. This guards the exact cross-cutting hazard the
// design had to solve: a new row type silently inflating an existing metric.
func TestParadataInvisibleToOwner(t *testing.T) {
	h := newHarness(t)

	var form struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	h.do(http.MethodPost, "/api/v1/forms", map[string]string{"title": "Funnel"}, &form)
	var q struct {
		ID string `json:"id"`
	}
	h.mustCreateQuestion(form.ID, map[string]any{"type": "short_text", "title": "Name", "required": true}, &q)
	if code := h.do(http.MethodPost, "/api/v1/forms/"+form.ID+"/publish", nil, nil); code != http.StatusOK {
		t.Fatalf("publish: %d", code)
	}

	// Start a session (in_progress row) + advance progress.
	var started struct {
		ResponseID string `json:"response_id"`
	}
	if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/start", nil, &started); code != http.StatusCreated {
		t.Fatalf("start: want 201, got %d", code)
	}
	if started.ResponseID == "" {
		t.Fatal("start returned no response_id")
	}
	if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/progress",
		map[string]any{"response_id": started.ResponseID, "position": 1}, nil); code != http.StatusOK {
		t.Fatalf("progress: want 200, got %d", code)
	}

	// Owner surfaces must NOT see the in-progress row.
	var analytics struct {
		ResponseCount int `json:"response_count"`
	}
	h.do(http.MethodGet, "/api/v1/forms/"+form.ID+"/analytics", nil, &analytics)
	if analytics.ResponseCount != 0 {
		t.Fatalf("in-progress leaked into analytics: response_count = %d, want 0", analytics.ResponseCount)
	}
	var responses []struct{ ID string }
	h.do(http.MethodGet, "/api/v1/forms/"+form.ID+"/responses", nil, &responses)
	if len(responses) != 0 {
		t.Fatalf("in-progress leaked into results list: %d rows, want 0", len(responses))
	}

	// A real submission DOES count.
	h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/responses",
		map[string]any{"answers": []map[string]any{{"question_id": q.ID, "value": "Alice"}}}, nil)
	h.do(http.MethodGet, "/api/v1/forms/"+form.ID+"/analytics", nil, &analytics)
	if analytics.ResponseCount != 1 {
		t.Fatalf("after submit: response_count = %d, want 1", analytics.ResponseCount)
	}

	// Progress on a bogus id -> 404; progress after completion is a no-op (200/404 tolerated).
	if code, _ := h.doAnon(http.MethodPost, "/api/v1/public/forms/"+form.Slug+"/progress",
		map[string]any{"response_id": "00000000-0000-0000-0000-000000000000", "position": 1}, nil); code != http.StatusNotFound {
		t.Fatalf("progress bogus id: want 404, got %d", code)
	}
}
