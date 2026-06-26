package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// maxSubmitBytes caps an anonymous submission body — generous for any real
// survey, but bounds memory against a hostile oversized payload.
const maxSubmitBytes = 256 << 10 // 256 KiB

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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSubmitBytes)
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
