package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/middleware"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/auth"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// AuthHandler runs the Google sign-in endpoints and session lifecycle.
type AuthHandler struct {
	svc *service.AuthService
	jwt *auth.JWTManager
	cfg *config.Config
}

func NewAuthHandler(svc *service.AuthService, jwt *auth.JWTManager, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, jwt: jwt, cfg: cfg}
}

// GoogleLogin redirects the browser to Google's consent screen.
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if !h.cfg.GoogleConfigured() {
		response.Error(c, http.StatusServiceUnavailable, "GOOGLE_NOT_CONFIGURED", "Google sign-in is not configured")
		return
	}
	c.Redirect(http.StatusFound, h.svc.AuthURL())
}

// GoogleCallback completes the OAuth handshake, mints a session cookie, and
// redirects back to the SPA.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if !h.cfg.GoogleConfigured() {
		response.Error(c, http.StatusServiceUnavailable, "GOOGLE_NOT_CONFIGURED", "Google sign-in is not configured")
		return
	}
	user, err := h.svc.HandleCallback(c.Request.Context(), c.Query("state"), c.Query("code"))
	if err != nil {
		handleServiceError(c, err, "google callback")
		return
	}
	token, err := h.jwt.GenerateSessionToken(user.ID, user.Email, user.Name)
	if err != nil {
		handleServiceError(c, err, "mint session")
		return
	}
	h.setSessionCookie(c, token, int(h.jwt.TTL().Seconds()))
	c.Redirect(http.StatusFound, strings.TrimRight(h.cfg.AppBaseURL, "/")+"/app")
}

// Logout revokes the current session token (so it can't be replayed even if it
// was captured) and clears the cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	if raw, err := c.Cookie(middleware.SessionCookieName); err == nil && raw != "" {
		h.jwt.RevokeToken(raw)
	}
	h.setSessionCookie(c, "", -1)
	response.OK(c, nil, "logged out")
}

// DevLogin mints a session for a synthetic test creator without Google. It is
// for local development / Playwright E2E only and is refused in production
// (defence-in-depth: the route is also not mounted when APP_ENV=production).
func (h *AuthHandler) DevLogin(c *gin.Context) {
	if h.cfg.Env == "production" {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
		return
	}
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "email is required")
		return
	}
	user, err := h.svc.DevLogin(c.Request.Context(), req.Email, req.Name)
	if err != nil {
		handleServiceError(c, err, "dev login")
		return
	}
	token, err := h.jwt.GenerateSessionToken(user.ID, user.Email, user.Name)
	if err != nil {
		handleServiceError(c, err, "mint session")
		return
	}
	h.setSessionCookie(c, token, int(h.jwt.TTL().Seconds()))
	response.OK(c, user, "dev login ok")
}

// Me returns the currently signed-in creator.
func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.svc.CurrentUser(c.Request.Context(), userID(c))
	if err != nil {
		handleServiceError(c, err, "current user")
		return
	}
	if user == nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	response.OK(c, user, "ok")
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, value string, maxAge int) {
	path := h.cfg.CookiePath
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
