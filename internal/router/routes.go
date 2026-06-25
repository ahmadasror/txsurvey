package router

import (
	"time"

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

	// Public runner endpoints (anonymous, rate-limited per IP). Submissions get
	// a stricter cap than reads.
	public := api.Group("/public")
	public.Use(middleware.RateLimit(120, time.Minute))
	public.GET("/forms/:slug", h.Public.GetForm)
	public.POST("/forms/:slug/responses", middleware.RateLimit(20, time.Minute), h.Public.Submit)

	// Creator-authenticated group (session cookie required).
	authed := api.Group("")
	authed.Use(middleware.SessionAuth(jwtMgr))
	authed.POST("/auth/logout", h.Auth.Logout)
	authed.GET("/auth/me", h.Auth.Me)

	// Forms.
	authed.GET("/forms", h.Form.List)
	authed.POST("/forms", h.Form.Create)
	authed.GET("/forms/:id", h.Form.Get)
	authed.PATCH("/forms/:id", h.Form.Update)
	authed.DELETE("/forms/:id", h.Form.Delete)
	authed.POST("/forms/:id/publish", h.Form.Publish)
	authed.POST("/forms/:id/unpublish", h.Form.Unpublish)

	// Asset upload (banner/logo) for an owned form.
	authed.POST("/forms/:id/assets", h.Asset.Upload)

	// Questions (nested under a form).
	authed.POST("/forms/:id/questions", h.Question.Create)
	authed.PUT("/forms/:id/questions/reorder", h.Question.Reorder)
	authed.PATCH("/forms/:id/questions/:qid", h.Question.Update)
	authed.DELETE("/forms/:id/questions/:qid", h.Question.Delete)

	// Logic rules (nested under a form).
	authed.GET("/forms/:id/logic", h.Logic.List)
	authed.POST("/forms/:id/logic", h.Logic.Create)
	authed.PATCH("/forms/:id/logic/:rid", h.Logic.Update)
	authed.DELETE("/forms/:id/logic/:rid", h.Logic.Delete)

	// Results: responses, analytics, CSV export.
	authed.GET("/forms/:id/responses", h.Results.ListResponses)
	authed.GET("/forms/:id/responses/:rid", h.Results.GetResponse)
	authed.GET("/forms/:id/analytics", h.Results.Analytics)
	authed.GET("/forms/:id/export.csv", h.Results.ExportCSV)

	_ = cfg
}
