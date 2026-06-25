package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
)

// registerRoutes mounts the /api/v1 surface. Route groups are added per phase:
//   - Phase 1: /auth/* (Google sign-in) + SessionAuth-protected groups
//   - Phase 2: /forms, /forms/:id/questions
//   - Phase 3: /public/forms/:slug (+ responses submit)
//   - Phase 4: /forms/:id/{responses,analytics,export.csv}
//   - Phase 5: /forms/:id/logic
func registerRoutes(r *gin.Engine, cfg *config.Config, h *Handlers) {
	_ = r.Group("/api/v1")
	_ = cfg
	_ = h
}
