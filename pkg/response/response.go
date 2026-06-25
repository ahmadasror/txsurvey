// Package response is the single JSON envelope used by every handler:
// {success, data, message, meta?} on success and {success:false, error:{...}}
// on failure.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func OK(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, SuccessResponse{Success: true, Data: data, Message: message})
}

func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, SuccessResponse{Success: true, Data: data, Message: message})
}

func Paginated(c *gin.Context, data interface{}, meta Meta, message string) {
	c.JSON(http.StatusOK, SuccessResponse{Success: true, Data: data, Message: message, Meta: &meta})
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Success: false, Error: ErrorDetail{Code: code, Message: message}})
}

func ValidationError(c *gin.Context, details interface{}) {
	c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
		Success: false,
		Error:   ErrorDetail{Code: "VALIDATION_ERROR", Message: "Validation failed", Details: details},
	})
}
