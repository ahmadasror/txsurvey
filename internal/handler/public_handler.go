package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// PublicHandler exposes the anonymous runner endpoints (no auth).
type PublicHandler struct {
	svc *service.ResponseService
}

func NewPublicHandler(svc *service.ResponseService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

// GetForm returns the runner contract for a published form by slug.
func (h *PublicHandler) GetForm(c *gin.Context) {
	form, err := h.svc.GetPublicForm(c.Request.Context(), c.Param("slug"))
	if err != nil {
		handleServiceError(c, err, "get public form")
		return
	}
	response.OK(c, form, "ok")
}

// Submit validates and stores a completed submission.
func (h *PublicHandler) Submit(c *gin.Context) {
	req, ok := bindJSON[dto.SubmitResponseRequest](c)
	if !ok {
		return
	}
	meta := model.ResponseMeta{
		UserAgent: c.Request.UserAgent(),
		Referrer:  c.Request.Referer(),
	}
	id, err := h.svc.Submit(c.Request.Context(), c.Param("slug"), req, meta)
	if err != nil {
		handleServiceError(c, err, "submit response")
		return
	}
	response.Created(c, gin.H{"response_id": id}, "response submitted")
}
