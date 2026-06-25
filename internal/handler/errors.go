package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/pkg/apperror"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// handleServiceError maps a service error to an HTTP response: a nil error means
// "not found", a *ClientError is surfaced verbatim, anything else is logged and
// returned as a generic 500.
func handleServiceError(c *gin.Context, err error, action string) {
	if err == nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Resource not found")
		return
	}
	if ce, ok := apperror.IsClient(err); ok {
		response.Error(c, ce.Status, ce.Code, ce.Message)
		return
	}
	slog.Error("unexpected error", "action", action, "error", err)
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}
