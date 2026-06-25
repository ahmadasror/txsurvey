// Package router wires the Gin engine: middleware stack, CORS, health, and the
// /api/v1 route groups (creator-authenticated vs public).
package router

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/handler"
	"github.com/ahmadasror/txsurvey/internal/middleware"
	"github.com/ahmadasror/txsurvey/internal/web"
	"github.com/ahmadasror/txsurvey/pkg/auth"
	"github.com/ahmadasror/txsurvey/pkg/response"
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

	// Serve the embedded SPA when present (production `-tags embedspa` build).
	// In dev the SPA runs on Vite and this is a no-op.
	if distFS, ok := web.DistFS(); ok {
		serveSPA(r, distFS)
	}
	return r
}

// serveSPA serves static assets from the embedded dist, falling back to
// index.html for any non-API, non-asset path so client-side routing works.
func serveSPA(r *gin.Engine, dist fs.FS) {
	fileServer := http.FileServer(http.FS(dist))
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
			return
		}
		if trimmed := strings.TrimPrefix(path.Clean(p), "/"); trimmed != "" {
			if f, err := dist.Open(trimmed); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		c.Request.URL.Path = "/" // SPA fallback -> index.html
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
