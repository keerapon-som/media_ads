package http

import (
	"media_ads/internal/config"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

var (
	buildtime, buildcommit, version string
)

func NewHTTPRouter(mediaArchiveHandler ObjectLibraryProviderHTTPInterface) *fiber.App {
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

	// app.Get("/hello_command", h.HelloCQRSCommand)
	// app.Get("/hello_event", h.HelloCQRSEvent)

	objectLib := app.Group("/object_library")
	objectLib.Post("/researve_upload_slot", mediaArchiveHandler.ReserveUploadSlot) // should use as Internal and need secure
	objectLib.Post("/publish/:object_id", mediaArchiveHandler.PublishObject)       // should use as Internal and need secure
	objectLib.Post("/unpublish/:object_id", mediaArchiveHandler.UnpublishObject)   // should use as Internal and need secure
	objectLib.Put("/upload/:upload_id", mediaArchiveHandler.UploadObject)
	objectLib.Get("/object/:object_id", mediaArchiveHandler.GetObject)
	objectLib.Get("/object_info/:object_id", mediaArchiveHandler.GetObjectInfo)
	objectLib.Delete("/object/:object_id", mediaArchiveHandler.DeleteObject)

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
