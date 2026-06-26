package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/pkg/response"
)

const maxUploadBytes = 2 << 20 // 2 MiB

// allowedImageTypes maps a sniffed MIME type (http.DetectContentType) to a file
// extension. GIF is intentionally excluded — it is a classic polyglot/XSS
// vector and unnecessary for a banner/logo.
var allowedImageTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

// AssetHandler stores and serves uploaded form assets (banner/logo).
type AssetHandler struct {
	forms    *repository.FormRepo
	dir      string
	limitDir int64 // total bytes allowed in dir (<=0 = unlimited)
}

func NewAssetHandler(forms *repository.FormRepo, dir string, limitDir int64) *AssetHandler {
	return &AssetHandler{forms: forms, dir: dir, limitDir: limitDir}
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
	_ = hdr // the multipart Content-Type header is attacker-controlled; ignored.

	// Determine the real type from the file's magic bytes, not the client's
	// declared Content-Type (which can lie). DetectContentType reads <=512 bytes.
	sniff := make([]byte, 512)
	n, err := io.ReadFull(file, sniff)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		handleServiceError(c, err, "read upload")
		return
	}
	ext, ok := allowedImageTypes[http.DetectContentType(sniff[:n])]
	if !ok {
		response.Error(c, http.StatusUnprocessableEntity, "BAD_IMAGE", "only PNG, JPEG or WEBP images are allowed")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		handleServiceError(c, err, "rewind upload")
		return
	}

	// Enforce the total-storage budget for this stage.
	if h.limitDir > 0 {
		used, err := dirSize(h.dir)
		if err != nil {
			handleServiceError(c, err, "measure upload dir")
			return
		}
		if used+hdr.Size > h.limitDir {
			response.Error(c, http.StatusInsufficientStorage, "STORAGE_FULL",
				"storage is full — delete some images or contact the owner")
			return
		}
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

// dirSize sums the bytes of every regular file under dir (0 if dir is absent).
func dirSize(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if os.IsNotExist(err) {
		return 0, nil
	}
	return total, err
}
