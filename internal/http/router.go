package http

import (
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

var (
	buildtime, buildcommit, version string
)

type Handler struct {
	commandBus      *cqrs.CommandBus
	eventBus        *cqrs.EventBus
	watermilllogger *log.WatermillLogrusAdapter
}

func NewHTTPRouter(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, watermillLogger *log.WatermillLogrusAdapter) *fiber.App {
	app := fiber.New(fiber.Config{
		Immutable: true,
	})

	app.Use(pprof.New())

	app.Get("/version", getVersion)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
		})
	})

	h := &Handler{
		commandBus:      commandBus,
		eventBus:        eventBus,
		watermilllogger: watermillLogger,
	}

	app.Get("/hello_command", h.HelloCQRSCommand)
	app.Get("/hello_event", h.HelloCQRSEvent)

	return app

}

func getVersion(c *fiber.Ctx) error {

	versionInfo := struct {
		BuildCommit string
		BuildTime   string
		Version     string
	}{
		BuildCommit: buildcommit,
		BuildTime:   buildtime,
		Version:     version,
	}

	return c.Status(http.StatusOK).JSON(versionInfo)
}
