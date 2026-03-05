package http

import (
	"context"
	"media_ads/internal/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/gofiber/fiber/v2"
	"github.com/lithammer/shortuuid/v3"
)

func (h *Handler) HelloCQRSCommand(c *fiber.Ctx) error {
	h.watermilllogger.Info("HelloCQRSCommand", nil)

	command := &entities.HelloCQRSCommand{
		Hello: "Hello CQRS",
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	if log.CorrelationIDFromContext(ctx) == "" {
		ctx = log.ContextWithCorrelationID(ctx, shortuuid.New())
	}

	err := h.commandBus.Send(ctx, command)
	if err != nil {
		h.watermilllogger.Error("Error sending command", err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	h.watermilllogger.Info("Command sent", nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   command,
	})
}
