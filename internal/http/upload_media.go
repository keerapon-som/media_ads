package http

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) UploadMedia(c *fiber.Ctx) error {

	objectID := c.FormValue("object_id")
	if objectID == "" {
		objectID = uuid.NewString()
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "file is required in form field 'file'",
		})
	}

	err = h.mediaProvider.UploadMedia(objectID, fileHeader)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to upload media",
			"detail": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"objectID":         objectID,
		"progress_percent": 100,
		"completed":        true,
	})
}

func (h *Handler) DownloadMedia(c *fiber.Ctx) error {

	objectID := c.Params("object_id")
	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	res, err := h.mediaProvider.DownloadMedia(objectID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to download media",
			"detail": err.Error(),
		})
	}

	c.Download(res.File.Name(), objectID+"."+res.Extension)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
	})
}

// func (h *Handler) UploadMediaStatus(c *fiber.Ctx) error {
// 	uploadID := strings.TrimSpace(c.Params("upload_id"))
// 	if uploadID == "" {
// 		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
// 			"error": "upload_id is required",
// 		})
// 	}

// 	finalPath, found := findCompletedFile(uploadID)
// 	if !found {
// 		return c.Status(http.StatusNotFound).JSON(fiber.Map{
// 			"error": "upload not found",
// 		})
// 	}

// 	checksum, err := fileSHA256(finalPath)
// 	if err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "failed to calculate checksum",
// 		})
// 	}

// 	return c.Status(http.StatusOK).JSON(fiber.Map{
// 		"upload_id":        uploadID,
// 		"completed":        true,
// 		"progress_percent": 100,
// 		"stored_path":      finalPath,
// 		"checksum_sha256":  checksum,
// 	})
// }

// func findCompletedFile(uploadID string) (string, bool) {
// 	completedDir := filepath.Join(os.TempDir(), "media_ads_uploads", "completed")
// 	entries, err := os.ReadDir(completedDir)
// 	if err != nil {
// 		return "", false
// 	}

// 	prefix := uploadID + "_"
// 	for _, entry := range entries {
// 		if entry.IsDir() {
// 			continue
// 		}
// 		name := entry.Name()
// 		if strings.HasPrefix(name, prefix) {
// 			return filepath.Join(completedDir, name), true
// 		}
// 	}

// 	return "", false
// }

// func fileSHA256(path string) (string, error) {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer f.Close()

// 	h := sha256.New()
// 	if _, err := io.Copy(h, f); err != nil {
// 		return "", err
// 	}

// 	return hex.EncodeToString(h.Sum(nil)), nil
// }
