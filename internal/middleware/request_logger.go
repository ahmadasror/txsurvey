package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/logging"
)

// RequestLogger emits one structured slog line per request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logging.With(c.Request.Context()).Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
