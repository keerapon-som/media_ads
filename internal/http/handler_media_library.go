package http

import (
	"media_ads/internal/domain"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ObjectLibraryHTTPHandler struct {
	mediaProvider   *domain.ObjectLibrary
	commandBus      *cqrs.CommandBus
	eventBus        *cqrs.EventBus
	watermilllogger *log.WatermillLogrusAdapter
}

type ObjectLibraryProviderHTTPInterface interface {
	ReserveUploadSlot(c *fiber.Ctx) error
	UploadObject(c *fiber.Ctx) error
	GetObject(c *fiber.Ctx) error
	GetObjectInfo(c *fiber.Ctx) error
	DeleteObject(c *fiber.Ctx) error
	PublishObject(c *fiber.Ctx) error
	UnpublishObject(c *fiber.Ctx) error
}

func NewObjectLibraryProviderHTTPHandler(mediaProvider *domain.ObjectLibrary, commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, watermillLogger *log.WatermillLogrusAdapter) ObjectLibraryProviderHTTPInterface {
	return &ObjectLibraryHTTPHandler{
		mediaProvider:   mediaProvider,
		commandBus:      commandBus,
		eventBus:        eventBus,
		watermilllogger: watermillLogger,
	}
}

func (h *ObjectLibraryHTTPHandler) ReserveUploadSlot(c *fiber.Ctx) error {
	uploadID, err := h.mediaProvider.ReserveUploadSlot()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to reserve upload slot",
			"detail": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"upload_id": uploadID,
	})
}

func (h *ObjectLibraryHTTPHandler) UploadObject(c *fiber.Ctx) error {

	upload_id := c.Params("upload_id")
	if upload_id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "upload_id is required in path",
		})
	}

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

	err = h.mediaProvider.UploadObject(upload_id, objectID, fileHeader)
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

func (h *ObjectLibraryHTTPHandler) GetObject(c *fiber.Ctx) error {

	objectID := c.Params("object_id")
	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	res, err := h.mediaProvider.GetObject(objectID)
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

func (h *ObjectLibraryHTTPHandler) GetObjectInfo(c *fiber.Ctx) error {
	objectID := c.Params("object_id")
	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	res, err := h.mediaProvider.GetObjectInfo(objectID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to get media info",
			"detail": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   res,
	})
}

func (h *ObjectLibraryHTTPHandler) DeleteObject(c *fiber.Ctx) error {

	objectID := c.Params("object_id")
	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	err := h.mediaProvider.DeleteObject(objectID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to delete media",
			"detail": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"id":     objectID,
	})
}

func (h *ObjectLibraryHTTPHandler) PublishObject(c *fiber.Ctx) error {
	objectID := c.Params("object_id")
	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	err := h.mediaProvider.PublishObject(objectID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to publish media",
			"detail": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"id":     objectID,
	})
}

func (h *ObjectLibraryHTTPHandler) UnpublishObject(c *fiber.Ctx) error {
	objectID := c.Params("object_id")

	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	err := h.mediaProvider.UnpublishObject(objectID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to unpublish media",
			"detail": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"id":     objectID,
	})
}
