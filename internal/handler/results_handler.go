package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

// ResultsHandler exposes a form owner's results: responses, analytics, CSV.
type ResultsHandler struct {
	svc *service.ResultsService
}

func NewResultsHandler(svc *service.ResultsService) *ResultsHandler {
	return &ResultsHandler{svc: svc}
}

func (h *ResultsHandler) ListResponses(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))
	items, total, err := h.svc.ListResponses(c.Request.Context(), userID(c), c.Param("id"), page, perPage)
	if err != nil {
		handleServiceError(c, err, "list responses")
		return
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	response.Paginated(c, items, response.Meta{Page: page, PerPage: perPage, Total: total}, "ok")
}

func (h *ResultsHandler) GetResponse(c *gin.Context) {
	resp, err := h.svc.GetResponse(c.Request.Context(), userID(c), c.Param("id"), c.Param("rid"))
	if err != nil {
		handleServiceError(c, err, "get response")
		return
	}
	response.OK(c, resp, "ok")
}

// DeleteResponses clears all collected responses for a form (keeps the form and
// its questions). Destructive — the SPA guards it behind a confirm dialog.
func (h *ResultsHandler) DeleteResponses(c *gin.Context) {
	deleted, err := h.svc.DeleteResponses(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "delete responses")
		return
	}
	response.OK(c, gin.H{"deleted": deleted}, "responses deleted")
}

func (h *ResultsHandler) Analytics(c *gin.Context) {
	a, err := h.svc.Analytics(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "analytics")
		return
	}
	response.OK(c, a, "ok")
}

// Funnel returns the response drop-off funnel (per-question retention).
func (h *ResultsHandler) Funnel(c *gin.Context) {
	f, err := h.svc.Funnel(c.Request.Context(), userID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "funnel")
		return
	}
	response.OK(c, f, "ok")
}

// ExportCSV buffers the CSV first so any error still yields a clean JSON error
// (headers aren't committed until the whole file is built).
func (h *ResultsHandler) ExportCSV(c *gin.Context) {
	var buf bytes.Buffer
	filename, err := h.svc.ExportCSV(c.Request.Context(), userID(c), c.Param("id"), &buf)
	if err != nil {
		handleServiceError(c, err, "export csv")
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}
