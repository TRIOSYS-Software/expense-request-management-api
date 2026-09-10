package controllers

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func serveUploadedFile(c echo.Context, uploadDir, name string) error {
	if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "attachment not found"})
	}

	fullPath := filepath.Join(uploadDir, name)

	// Confirm the result is still inside uploadDir after resolution.
	base, err := filepath.Abs(uploadDir)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "attachment not found"})
	}
	resolved, err := filepath.Abs(fullPath)
	if err != nil || !strings.HasPrefix(resolved, base+string(os.PathSeparator)) {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "attachment not found"})
	}

	return c.File(resolved)
}

type SavedAttachment struct {
	StoredName   string
	OriginalName string
	ContentType  string
}

func storeUpload(fh *multipart.FileHeader, uploadDir, storedName string) error {
	src, err := fh.Open()
	if err != nil {
		return errors.New("Failed to open file")
	}
	defer src.Close()

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return errors.New("Failed to create upload directory")
	}

	dst, err := os.Create(filepath.Join(uploadDir, storedName))
	if err != nil {
		return errors.New("Failed to create file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return errors.New("Failed to save file")
	}
	return nil
}

func saveLegacyAttachment(c echo.Context, uploadDir string) (*string, error) {
	fh, err := c.FormFile("attachment")
	if err != nil {
		return nil, nil
	}
	storedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fh.Filename))
	if err := storeUpload(fh, uploadDir, storedName); err != nil {
		return nil, err
	}
	return &storedName, nil
}

func saveMultiAttachments(c echo.Context, uploadDir string) []SavedAttachment {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil
	}

	files := form.File["attachments"]
	saved := make([]SavedAttachment, 0, len(files))
	for _, fh := range files {
		storedName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "multi", filepath.Ext(fh.Filename))
		if err := storeUpload(fh, uploadDir, storedName); err != nil {
			continue
		}
		saved = append(saved, SavedAttachment{
			StoredName:   storedName,
			OriginalName: fh.Filename,
			ContentType:  fh.Header.Get("Content-Type"),
		})
	}
	return saved
}

func attachmentRetention(c echo.Context) (keptIDs []uint, keepLegacy bool) {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil, false
	}
	for _, raw := range form.Value["kept_attachment_ids"] {
		if id, err := strconv.Atoi(raw); err == nil {
			keptIDs = append(keptIDs, uint(id))
		}
	}
	if v, ok := form.Value["keep_legacy_attachment"]; ok && len(v) > 0 {
		keepLegacy, _ = strconv.ParseBool(v[0])
	}
	return keptIDs, keepLegacy
}

func discardStoredFiles(uploadDir string, names ...string) {
	for _, name := range names {
		if name == "" {
			continue
		}
		os.Remove(filepath.Join(uploadDir, name))
	}
}
