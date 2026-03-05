package http

import (
	"media_ads/internal/entities"
	"media_ads/internal/util"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HelloCQRSEvent(c *fiber.Ctx) error {
	h.watermilllogger.Log.Info("HELLO")

	event := &entities.HelloCQRSEvent{
		Header: entities.NewEventHeader(),
		Hello:  "Hello CQRS",
	}

	err := h.eventBus.Publish(util.NewContext(c.Context()), event)
	if err != nil {
		h.watermilllogger.Error("Error publishing event", err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	h.watermilllogger.Log.Info("DONEEE")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   event,
	})
}
