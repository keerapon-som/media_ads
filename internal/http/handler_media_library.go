package http

import (
	"media_ads/internal/domain"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MediaLibraryHTTPHandler struct {
	mediaProvider   *domain.MediaLibrary
	commandBus      *cqrs.CommandBus
	eventBus        *cqrs.EventBus
	watermilllogger *log.WatermillLogrusAdapter
}

type MediaLibraryProviderHTTPInterface interface {
	UploadMedia(c *fiber.Ctx) error
	GetObject(c *fiber.Ctx) error
}

func NewMediaLibraryProviderHTTPHandler(mediaProvider *domain.MediaLibrary, commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, watermillLogger *log.WatermillLogrusAdapter) MediaLibraryProviderHTTPInterface {
	return &MediaLibraryHTTPHandler{
		mediaProvider:   mediaProvider,
		commandBus:      commandBus,
		eventBus:        eventBus,
		watermilllogger: watermillLogger,
	}
}

func (h *MediaLibraryHTTPHandler) UploadMedia(c *fiber.Ctx) error {

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

// func (h *Handler) DownloadMedia(c *fiber.Ctx) error {

// 	objectID := c.Params("object_id")
// 	if objectID == "" {
// 		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
// 			"error": "object_id is required in path",
// 		})
// 	}

// 	res, err := h.mediaProvider.DownloadMedia(objectID)
// 	if err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error":  "failed to download media",
// 			"detail": err.Error(),
// 		})
// 	}

// 	c.Download(res.File.Name(), objectID+"."+res.Extension)

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status": "success",
// 	})
// }

func (h *MediaLibraryHTTPHandler) GetObject(c *fiber.Ctx) error {

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

	c.Set("Content-Type", res.ContentType)
	c.Set("Content-Disposition", "inline; filename=\""+objectID+"."+res.Extension+"\"")

	if res.SizeBytes > 0 {
		return c.Status(http.StatusOK).SendStream(res.File, int(res.SizeBytes))
	}

	return c.Status(http.StatusOK).SendStream(res.File)

}
