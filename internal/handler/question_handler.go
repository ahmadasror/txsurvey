package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// QuestionHandler exposes the nested question endpoints under a form.
type QuestionHandler struct {
	svc *service.QuestionService
}

func NewQuestionHandler(svc *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: svc}
}

func (h *QuestionHandler) Create(c *gin.Context) {
	req, ok := bindJSON[dto.QuestionInput](c)
	if !ok {
		return
	}
	q, err := h.svc.Add(c.Request.Context(), userID(c), c.Param("id"), req)
	if err != nil {
		handleServiceError(c, err, "add question")
		return
	}
	response.Created(c, q, "question added")
}

func (h *QuestionHandler) Update(c *gin.Context) {
	req, ok := bindJSON[dto.QuestionInput](c)
	if !ok {
		return
	}
	q, err := h.svc.Update(c.Request.Context(), userID(c), c.Param("id"), c.Param("qid"), req)
	if err != nil {
		handleServiceError(c, err, "update question")
		return
	}
	response.OK(c, q, "question updated")
}

func (h *QuestionHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), userID(c), c.Param("id"), c.Param("qid")); err != nil {
		handleServiceError(c, err, "delete question")
		return
	}
	response.OK(c, nil, "question deleted")
}

func (h *QuestionHandler) Reorder(c *gin.Context) {
	req, ok := bindJSON[dto.ReorderRequest](c)
	if !ok {
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), userID(c), c.Param("id"), req.OrderedIDs); err != nil {
		handleServiceError(c, err, "reorder questions")
		return
	}
	response.OK(c, nil, "questions reordered")
}
