package http

import (
	"fmt"
	"media_ads/internal/config"
	"media_ads/internal/domain"
	"media_ads/internal/entities"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ObjectLibraryHTTPHandler struct {
	objectLibrary   domain.ObjectLibraryInterface
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

func NewObjectLibraryProviderHTTPHandler(objectLibrary domain.ObjectLibraryInterface, commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, watermillLogger *log.WatermillLogrusAdapter) ObjectLibraryProviderHTTPInterface {
	return &ObjectLibraryHTTPHandler{
		objectLibrary:   objectLibrary,
		commandBus:      commandBus,
		eventBus:        eventBus,
		watermilllogger: watermillLogger,
	}
}

func (h *ObjectLibraryHTTPHandler) ReserveUploadSlot(c *fiber.Ctx) error {
	uploadID, err := h.objectLibrary.ReserveUploadSlot()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to reserve upload slot",
			"detail": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"upload_id": fmt.Sprintf("%s/object_library/upload/%s", config.GetConfig().ServiceConfig.ObjectLibraryDomain.BaseURL, uploadID),
	})
}

func (h *ObjectLibraryHTTPHandler) UploadObject(c *fiber.Ctx) error {

	upload_id := c.Params("upload_id")
	if upload_id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "upload_id is required in path",
		})
	}

	mediaID := c.FormValue("media_id")
	if mediaID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "media_id is required in form data",
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

	mediaInfo, err := h.objectLibrary.UploadObject(upload_id, objectID, fileHeader)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to upload media",
			"detail": err.Error(),
		})
	}

	// do message callback upload success
	err = h.eventBus.Publish(c.Context(), &entities.RequestUploadCallbackEvent{
		Header:      entities.NewEventHeader(),
		MediaID:     mediaID,
		ObjectID:    objectID,
		ContentType: mediaInfo.ContentType,
		Success:     true,
	})
	if err != nil {
		h.watermilllogger.Error("failed to publish RequestUploadCallbackEvent", err, nil)
		// we don't return error to client even if the callback event publish fails, because the upload itself is successful. The callback will be retried by the event bus until it succeeds.
	}

	return c.Status(http.StatusOK).JSON(mediaInfo)
}

func (h *ObjectLibraryHTTPHandler) GetObject(c *fiber.Ctx) error {

	objectID := c.Params("object_id")
	if objectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required in path",
		})
	}

	res, err := h.objectLibrary.GetObject(objectID)
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

	res, err := h.objectLibrary.GetObjectInfo(objectID)
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

	err := h.objectLibrary.DeleteObject(objectID)
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

	err := h.objectLibrary.PublishObject(objectID)
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

	err := h.objectLibrary.UnpublishObject(objectID)
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
