package router

import (
	"rerng_addicted_api/pkg/utils"
	"time"

	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/text/language"
)

func New() *fiber.App {
	f := fiber.New(fiber.Config{
		// EnablePrintRoutes: true,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute, 
		IdleTimeout:  5 * time.Minute,
		ErrorHandler: utils.GlobalErrorHandler,
	})

	f.Use(logger.New())
	f.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	f.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	})).Use(
		fiberi18n.New(&fiberi18n.Config{
			RootPath:         "./pkg/i18n/localize",
			AcceptLanguages:  []language.Tag{language.Chinese, language.Khmer, language.English},
			DefaultLanguage:  language.Khmer,
			FormatBundleFile: "json",
		}),
	)

	return f
}
