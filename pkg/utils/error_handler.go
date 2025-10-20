package utils

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	switch code {
	case http.StatusNotFound:
		return c.Status(code).JSON(fiber.Map{
			"error": Translate("not_found", nil, c),
		})
	case http.StatusMethodNotAllowed:
		return c.Status(code).JSON(fiber.Map{
			"error": Translate("method_not_allow", nil, c),
		})
	default:
		return c.Status(code).JSON(fiber.Map{
			"error": Translate("internal_server_error", nil, c),
		})
	}
}
