package http

import (
	"crypto/subtle"
	"media_ads/internal/config"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

var (
	buildtime, buildcommit, version string
)

func NewHTTPRouter(mediaArchiveHandler ObjectLibraryProviderHTTPInterface, internalSecureToken string) *fiber.App {
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
	objectLib.Post("/researve_upload_slot", requireInternalSecureKey(internalSecureToken), mediaArchiveHandler.ReserveUploadSlot)
	objectLib.Post("/publish/:object_id", requireInternalSecureKey(internalSecureToken), mediaArchiveHandler.PublishObject)
	objectLib.Post("/unpublish/:object_id", requireInternalSecureKey(internalSecureToken), mediaArchiveHandler.UnpublishObject)
	objectLib.Delete("/object/:object_id", requireInternalSecureKey(internalSecureToken), mediaArchiveHandler.DeleteObject)
	objectLib.Put("/upload/:upload_id", mediaArchiveHandler.UploadObject)
	objectLib.Get("/object/:object_id", mediaArchiveHandler.GetObject)
	objectLib.Get("/object_info/:object_id", mediaArchiveHandler.GetObjectInfo)

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

func requireInternalSecureKey(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		expected := strings.TrimSpace(token)
		if expected == "" {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal token is not configured",
			})
		}

		token := strings.TrimSpace(c.Get("X-Internal-Token"))
		if token == "" {
			authHeader := strings.TrimSpace(c.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				token = strings.TrimSpace(authHeader[7:])
			}
		}

		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing internal token",
			})
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid internal token",
			})
		}

		return c.Next()
	}
}
