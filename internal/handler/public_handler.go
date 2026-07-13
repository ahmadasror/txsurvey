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

// Start opens an in-progress response (paradata capture) and returns its id for
// the runner to echo to Progress as the respondent navigates.
func (h *PublicHandler) Start(c *gin.Context) {
	meta := model.ResponseMeta{
		UserAgent: c.Request.UserAgent(),
		Referrer:  c.Request.Referer(),
	}
	id, err := h.svc.StartSession(c.Request.Context(), c.Param("slug"), meta)
	if err != nil {
		handleServiceError(c, err, "start response")
		return
	}
	response.Created(c, gin.H{"response_id": id}, "session started")
}

// Progress advances an in-progress response's furthest-reached position.
func (h *PublicHandler) Progress(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSubmitBytes)
	req, ok := bindJSON[dto.ProgressRequest](c)
	if !ok {
		return
	}
	if err := h.svc.UpdateProgress(c.Request.Context(), req.ResponseID, req.Position); err != nil {
		handleServiceError(c, err, "update progress")
		return
	}
	response.OK(c, gin.H{"ok": true}, "progress recorded")
}
