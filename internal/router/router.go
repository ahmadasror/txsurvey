// Package router wires the Gin engine: middleware stack, CORS, health, and the
// /api/v1 route groups (creator-authenticated vs public).
package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/handler"
	"github.com/ahmadasror/txsurvey/internal/middleware"
	"github.com/ahmadasror/txsurvey/pkg/auth"
)

// Handlers holds every HTTP handler, wired in main as phases add them.
type Handlers struct {
	Auth     *handler.AuthHandler
	Form     *handler.FormHandler
	Question *handler.QuestionHandler
	Public   *handler.PublicHandler
	Results  *handler.ResultsHandler
	Logic    *handler.LogicHandler
}

// Setup builds the configured Gin engine. jwtMgr backs the SessionAuth
// middleware that gates creator-only routes.
func Setup(cfg *config.Config, h *Handlers, jwtMgr *auth.JWTManager) *gin.Engine {
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

	registerRoutes(r, cfg, h, jwtMgr)
	return r
}
