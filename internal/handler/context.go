// context.go — Gin context helpers shared across handlers.
//
// The auth middleware (SessionAuth) writes the signed-in creator id under the
// key "user_id". Always read it via userID(c) so the key lives in one place.
package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/pkg/response"
)

// userID returns the authenticated creator id from the Gin context.
func userID(c *gin.Context) string {
	return c.GetString("user_id")
}

// bindJSON unmarshals the request body into a new zero-value T. On failure it
// writes a generic 422 (the raw binding error names internal Go types/fields, so
// it is logged server-side rather than returned) and returns ok=false; the
// caller MUST return.
func bindJSON[T any](c *gin.Context) (T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Debug("request body binding failed", "path", c.FullPath(), "error", err)
		response.ValidationError(c, "request body is malformed or missing required fields")
		return req, false
	}
	return req, true
}
