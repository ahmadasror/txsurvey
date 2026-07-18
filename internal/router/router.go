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
	Asset    *handler.AssetHandler
}

// Setup builds the configured Gin engine. jwtMgr backs the SessionAuth
// middleware that gates creator-only routes.
func Setup(cfg *config.Config, h *Handlers, jwtMgr *auth.JWTManager) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// Behind nginx on the same host: trust the loopback proxy so c.ClientIP()
	// (used by the rate limiter) reads the real client from X-Forwarded-For.
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	// Defensive headers on every response (CSP, nosniff, frame-ancestors, …).
	// HSTS only in production, where traffic is HTTPS (behind nginx/Cloudflare).
	r.Use(middleware.SecurityHeaders(cfg.Env == "production"))
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

	// Uploaded form assets (banner/logo) — public, served from disk.
	r.Static("/uploads", cfg.UploadDir)

	registerRoutes(r, cfg, h, jwtMgr)

	// Serve the embedded SPA when present (production `-tags embedspa` build).
	// In dev the SPA runs on Vite and this is a no-op.
	if distFS, ok := web.DistFS(); ok {
		serveSPA(r, distFS, cfg.AppBaseURL)
	}
	return r
}

// serveSPA serves static assets from the embedded dist, falling back to
// index.html for any non-API, non-asset path so client-side routing works.
func serveSPA(r *gin.Engine, dist fs.FS, appBaseURL string) {
	fileServer := http.FileServer(http.FS(dist))
	indexHTML, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return
	}
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
			return
		}
		if p != "/" && strings.HasSuffix(p, "/") {
			target := strings.TrimRight(p, "/")
			if c.Request.URL.RawQuery != "" {
				target += "?" + c.Request.URL.RawQuery
			}
			c.Redirect(http.StatusMovedPermanently, target)
			return
		}
		if p == "/sitemap.xml" {
			c.Header("Cache-Control", "public, max-age=3600")
			c.Data(http.StatusOK, "application/xml; charset=utf-8", sitemapXML(appBaseURL))
			return
		}
		if p == "/robots.txt" {
			c.Header("Cache-Control", "public, max-age=3600")
			c.Data(http.StatusOK, "text/plain; charset=utf-8", robotsText(appBaseURL))
			return
		}
		if trimmed := strings.TrimPrefix(path.Clean(p), "/"); trimmed != "" {
			if f, err := dist.Open(trimmed); err == nil {
				_ = f.Close()
				if strings.HasPrefix(p, "/assets/") {
					c.Header("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					c.Header("Cache-Control", "public, max-age=604800")
				}
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		page, known := seoPageForPath(p)
		status := http.StatusOK
		if !known {
			status = http.StatusNotFound
			page = seoPage{
				Title:       "Halaman Tidak Ditemukan · txsurvey",
				Description: "Halaman yang kamu cari tidak tersedia.",
				Robots:      "noindex, nofollow",
				Heading:     "Halaman tidak ditemukan.",
				Summary:     "Tautannya mungkin keliru atau halaman sudah dipindahkan.",
			}
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(status, "text/html; charset=utf-8", renderSEOHTML(indexHTML, page, appBaseURL))
	})
}
