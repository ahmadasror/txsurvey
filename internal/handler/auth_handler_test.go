package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/middleware"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/auth"
)

// fakeUserRepo satisfies service.UserRepository for handler tests.
type fakeUserRepo struct {
	user *model.User
}

func (f *fakeUserRepo) UpsertByGoogleSub(_ context.Context, _ model.GoogleProfile) (*model.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) UpsertByGoogleSubCapped(_ context.Context, _ model.GoogleProfile, _ int) (*model.User, bool, error) {
	return f.user, false, nil
}
func (f *fakeUserRepo) GetByID(_ context.Context, id string) (*model.User, error) {
	if f.user != nil && f.user.ID == id {
		return f.user, nil
	}
	return nil, nil
}

func newTestAuthRouter(t *testing.T, user *model.User) (*gin.Engine, *auth.JWTManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppBaseURL: "http://localhost:8080"}
	jwtMgr := auth.NewJWTManager("test-secret-at-least-32-characters-long!!", time.Hour)
	authSvc := service.NewAuthService(cfg, &fakeUserRepo{user: user})
	h := NewAuthHandler(authSvc, jwtMgr, cfg)

	r := gin.New()
	api := r.Group("/api/v1")
	authed := api.Group("")
	authed.Use(middleware.SessionAuth(jwtMgr))
	authed.GET("/auth/me", h.Me)
	return r, jwtMgr
}

func TestMe_NoCookie_Returns401(t *testing.T) {
	r, _ := newTestAuthRouter(t, &model.User{ID: "u1", Email: "a@b.com"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestMe_ValidCookie_ReturnsUser(t *testing.T) {
	user := &model.User{ID: "u1", Email: "a@b.com", Name: "Alice"}
	r, jwtMgr := newTestAuthRouter(t, user)

	tok, err := jwtMgr.GenerateSessionToken(user.ID, user.Email, user.Name)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: tok})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var body struct {
		Success bool       `json:"success"`
		Data    model.User `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.Success || body.Data.Email != "a@b.com" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
