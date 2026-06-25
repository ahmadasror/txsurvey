package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/dto"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// LogicHandler exposes the nested logic-rule endpoints under a form.
type LogicHandler struct {
	svc *service.LogicService
}

func NewLogicHandler(svc *service.LogicService) *LogicHandler {
	return &LogicHandler{svc: svc}
}

func (h *LogicHandler) List(c *gin.Context) {
	rules, err := h.svc.List(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "list logic rules")
		return
	}
	response.OK(c, rules, "ok")
}

func (h *LogicHandler) Create(c *gin.Context) {
	req, ok := bindJSON[dto.LogicRuleInput](c)
	if !ok {
		return
	}
	rule, err := h.svc.Create(c.Request.Context(), userID(c), c.Param("id"), req)
	if err != nil {
		handleServiceError(c, err, "create logic rule")
		return
	}
	response.Created(c, rule, "logic rule created")
}

func (h *LogicHandler) Update(c *gin.Context) {
	req, ok := bindJSON[dto.LogicRuleInput](c)
	if !ok {
		return
	}
	rule, err := h.svc.Update(c.Request.Context(), userID(c), c.Param("id"), c.Param("rid"), req)
	if err != nil {
		handleServiceError(c, err, "update logic rule")
		return
	}
	response.OK(c, rule, "logic rule updated")
}

func (h *LogicHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), userID(c), c.Param("id"), c.Param("rid")); err != nil {
		handleServiceError(c, err, "delete logic rule")
		return
	}
	response.OK(c, nil, "logic rule deleted")
}
