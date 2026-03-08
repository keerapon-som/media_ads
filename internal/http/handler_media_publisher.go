package http

import (
	"media_ads/internal/domain"
	"media_ads/internal/entities"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MediaPublisherHTTPHandler struct {
	mediaPublisher  domain.MediaPublisherInterface
	commandBus      *cqrs.CommandBus
	eventBus        *cqrs.EventBus
	watermilllogger *log.WatermillLogrusAdapter
}

func NewMediaPublisherHTTPHandler(mediaPublisher domain.MediaPublisherInterface, commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, watermillLogger *log.WatermillLogrusAdapter) *MediaPublisherHTTPHandler {
	return &MediaPublisherHTTPHandler{
		mediaPublisher:  mediaPublisher,
		commandBus:      commandBus,
		eventBus:        eventBus,
		watermilllogger: watermillLogger,
	}
}

// userID string, mediaID string, title string, description string
type SaveMediaRequest struct {
	UserID      string `json:"user_id"`
	MediaID     string `json:"media_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *MediaPublisherHTTPHandler) SaveMediaRequest(c *fiber.Ctx) error {

	var req SaveMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.Title == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and title are required",
		})
	}

	if req.MediaID == "" {
		req.MediaID = uuid.NewString()
	}

	resp, err := h.mediaPublisher.SaveMediaRequest(
		req.UserID,
		req.MediaID,
		req.Title,
		req.Description,
	)
	if err != nil {
		h.watermilllogger.Error("Error saving media request", err, nil)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"media_id":   resp.MediaID,
		"upload_url": resp.UploadURL,
	})
}

func (h *MediaPublisherHTTPHandler) ReceiveUploadCallback(c *fiber.Ctx) error {

	var req struct {
		MediaID     string `json:"media_id"`
		ObjectID    string `json:"object_id"`
		ContentType string `json:"content_type"`
		Success     bool   `json:"success"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.MediaID == "" || req.ContentType == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "media_id and content_type are required",
		})
	}

	if req.Success && req.ObjectID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "object_id is required when success is true",
		})
	}

	err := h.eventBus.Publish(c.Context(), &entities.ReceiveUploadCallbackEvent{
		Header:      entities.NewEventHeader(),
		MediaID:     req.MediaID,
		ObjectID:    req.ObjectID,
		ContentType: req.ContentType,
		Success:     req.Success,
	})
	if err != nil {
		h.watermilllogger.Error("Error publishing ReceiveUploadCallbackEvent", err, nil)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
	})
}
