package swagger

import (
	"fmt"
	"rerng_addicted_api/docs" 

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func Setup(app *fiber.App, host string, port int) {
	// Set Swagger metadata
	docs.SwaggerInfo.Title = "Mini Shop API"
	docs.SwaggerInfo.Description = "Professional API documentation for the Mini Shop backend."
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", host, port)
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Serve Swagger UI
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL: fmt.Sprintf("http://%s:%d/swagger/doc.json", host, port),
	}))
}
