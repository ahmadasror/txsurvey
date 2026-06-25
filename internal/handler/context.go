// context.go — Gin context helpers shared across handlers.
//
// The auth middleware (SessionAuth) writes the signed-in creator id under the
// key "user_id". Always read it via userID(c) so the key lives in one place.
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/pkg/response"
)

// userID returns the authenticated creator id from the Gin context.
func userID(c *gin.Context) string {
	return c.GetString("user_id")
}

// bindJSON unmarshals the request body into a new zero-value T. On failure it
// writes a 422 ValidationError and returns ok=false; the caller MUST return.
func bindJSON[T any](c *gin.Context) (T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return req, false
	}
	return req, true
}
