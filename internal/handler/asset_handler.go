package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

const maxUploadBytes = 2 << 20 // 2 MiB

var allowedImageTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// AssetHandler stores and serves uploaded form assets (banner/logo).
type AssetHandler struct {
	forms *repository.FormRepo
	dir   string
}

func NewAssetHandler(forms *repository.FormRepo, dir string) *AssetHandler {
	return &AssetHandler{forms: forms, dir: dir}
}

// Upload accepts a multipart image ("file") for an owned form and returns a
// relative URL ("uploads/<name>") to store in the form's settings.
func (h *AssetHandler) Upload(c *gin.Context) {
	form, err := h.forms.GetByIDOwned(c.Request.Context(), c.Param("id"), userID(c))
	if err != nil {
		handleServiceError(c, err, "asset upload form lookup")
		return
	}
	if form == nil {
		response.Error(c, http.StatusNotFound, "FORM_NOT_FOUND", "form not found")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)
	file, hdr, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "NO_FILE", "an image file is required (max 2MB)")
		return
	}
	defer file.Close()

	ext, ok := allowedImageTypes[hdr.Header.Get("Content-Type")]
	if !ok {
		response.Error(c, http.StatusUnprocessableEntity, "BAD_IMAGE", "only PNG, JPEG, WEBP or GIF images are allowed")
		return
	}

	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		handleServiceError(c, err, "ensure upload dir")
		return
	}
	name := randomName() + ext
	dst, err := os.Create(filepath.Join(h.dir, name))
	if err != nil {
		handleServiceError(c, err, "create asset file")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		handleServiceError(c, err, "write asset file")
		return
	}

	response.Created(c, gin.H{"url": "uploads/" + name}, "uploaded")
}

func randomName() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
