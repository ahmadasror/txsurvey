package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/middleware"
	"github.com/ahmadasror/txsurvey/pkg/auth"
)

// registerRoutes mounts the /api/v1 surface. Groups are added per phase:
//   - Phase 1: /auth/* (Google sign-in) + SessionAuth-protected groups
//   - Phase 2: /forms, /forms/:id/questions
//   - Phase 3: /public/forms/:slug (+ responses submit)
//   - Phase 4: /forms/:id/{responses,analytics,export.csv}
//   - Phase 5: /forms/:id/logic
func registerRoutes(r *gin.Engine, cfg *config.Config, h *Handlers, jwtMgr *auth.JWTManager) {
	api := r.Group("/api/v1")

	// Public auth endpoints (browser-driven OAuth handshake).
	api.GET("/auth/google/login", h.Auth.GoogleLogin)
	api.GET("/auth/google/callback", h.Auth.GoogleCallback)

	// Creator-authenticated group (session cookie required).
	authed := api.Group("")
	authed.Use(middleware.SessionAuth(jwtMgr))
	authed.POST("/auth/logout", h.Auth.Logout)
	authed.GET("/auth/me", h.Auth.Me)

	_ = cfg
}
