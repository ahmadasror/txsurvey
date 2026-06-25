// Package config loads + validates txsurvey configuration from the environment
// (struct-based, validated at startup — the house pattern from trail-service).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the full txsurvey configuration.
type Config struct {
	Env        string
	ServerPort string
	// AppBaseURL is the public origin of this app (SPA + OAuth return target).
	AppBaseURL string

	DatabaseURL string

	// Session (app-minted JWT, httpOnly cookie).
	JWTSecret    string
	SessionTTL   time.Duration
	CookieSecure bool
	// CookiePath scopes the session cookie. Set to the deploy subpath (e.g.
	// "/txsurvey") so the cookie isn't shared with other apps on the same host.
	CookiePath string

	// Google OAuth (sign-in only).
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// CORS allow-list (split-host dev only; same-origin in prod via embed).
	CORSAllowedOrigins []string

	// UploadDir is where uploaded form assets (banner/logo) are stored and
	// served from under /uploads.
	UploadDir string
}

// Load reads config from env and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{
		Env:                getEnv("APP_ENV", "development"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		AppBaseURL:         getEnv("APP_BASE_URL", "http://localhost:8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		SessionTTL:         getDuration("SESSION_TTL", 24*time.Hour),
		CookieSecure:       getBool("COOKIE_SECURE", false),
		CookiePath:         getEnv("COOKIE_PATH", "/"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		CORSAllowedOrigins: getCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		UploadDir:          getEnv("UPLOAD_DIR", "uploads"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET is required and must be >= 32 chars")
	}
	return cfg, nil
}

// GoogleConfigured reports whether Google OAuth credentials are present. When
// false, the auth endpoints return a clear 503 instead of a cryptic failure —
// useful for the Phase 0 skeleton before credentials are set.
func (c *Config) GoogleConfigured() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return def
}

func getCSV(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}
