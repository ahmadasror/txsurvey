package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// FormHandler exposes the creator-authenticated form endpoints.
type FormHandler struct {
	svc *service.FormService
}

func NewFormHandler(svc *service.FormService) *FormHandler {
	return &FormHandler{svc: svc}
}

func (h *FormHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))
	items, total, err := h.svc.List(c.Request.Context(), userID(c), page, perPage)
	if err != nil {
		handleServiceError(c, err, "list forms")
		return
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	response.Paginated(c, items, response.Meta{Page: page, PerPage: perPage, Total: total}, "ok")
}

func (h *FormHandler) Create(c *gin.Context) {
	req, ok := bindJSON[dto.CreateFormRequest](c)
	if !ok {
		return
	}
	form, err := h.svc.Create(c.Request.Context(), userID(c), req)
	if err != nil {
		handleServiceError(c, err, "create form")
		return
	}
	response.Created(c, form, "form created")
}

func (h *FormHandler) Get(c *gin.Context) {
	form, err := h.svc.Get(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "get form")
		return
	}
	response.OK(c, form, "ok")
}

func (h *FormHandler) Update(c *gin.Context) {
	req, ok := bindJSON[dto.UpdateFormRequest](c)
	if !ok {
		return
	}
	form, err := h.svc.Update(c.Request.Context(), userID(c), c.Param("id"), req)
	if err != nil {
		handleServiceError(c, err, "update form")
		return
	}
	response.OK(c, form, "form updated")
}

func (h *FormHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), userID(c), c.Param("id")); err != nil {
		handleServiceError(c, err, "delete form")
		return
	}
	response.OK(c, nil, "form deleted")
}

func (h *FormHandler) Publish(c *gin.Context) {
	form, err := h.svc.Publish(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "publish form")
		return
	}
	response.OK(c, form, "form published")
}

func (h *FormHandler) Unpublish(c *gin.Context) {
	form, err := h.svc.Unpublish(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "unpublish form")
		return
	}
	response.OK(c, form, "form unpublished")
}
