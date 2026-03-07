package http

import (
	"media_ads/internal/config"
	"media_ads/internal/domain"
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
	mediaProvider   *domain.MediaProvider
}

func NewHTTPRouter(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus, mediaProvider *domain.MediaProvider, watermillLogger *log.WatermillLogrusAdapter) *fiber.App {
	app := fiber.New(fiber.Config{
		Immutable: true,
		BodyLimit: config.GetConfig().ServerConfig.HTTP.BodyLimitBytes,
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
		mediaProvider:   mediaProvider,
	}

	app.Get("/hello_command", h.HelloCQRSCommand)
	app.Get("/hello_event", h.HelloCQRSEvent)
	app.Post("/upload_media", h.UploadMedia)
	// app.Get("/upload_media/:upload_id/status", h.UploadMediaStatus)

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
