// Package router wires the Gin engine: middleware stack, CORS, health, and the
// /api/v1 route groups (creator-authenticated vs public).
package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/middleware"
)

// Handlers holds every HTTP handler. Fields are populated in main as phases add
// them; registerRoutes only mounts the ones that are non-nil.
type Handlers struct {
	Auth *struct{} // placeholder until Phase 1 wires the real handlers
}

// Setup builds the configured Gin engine.
func Setup(cfg *config.Config, h *Handlers) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerRoutes(r, cfg, h)
	return r
}
