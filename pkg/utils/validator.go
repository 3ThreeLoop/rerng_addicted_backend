package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (v *Validator) Validate(i interface{}, c *fiber.Ctx) error {
	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}

	var error_messages []string

	if validation_errors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validation_errors {
			message := formatErrorMessage(e, c)
			error_messages = append(error_messages, message)
		}
	}

	error_string := strings.Join(error_messages, ", ")

	return fmt.Errorf(strings.ToLower(error_string))
}

func formatErrorMessage(e validator.FieldError, c *fiber.Ctx) string {
	switch e.Tag() {
	case "required":
		return Translate("required", map[string]interface{}{
			"field": e.Field(),
		}, c)
	case "email":
		return Translate("email", map[string]interface{}{
			"field": e.Field(),
		}, c)
	case "min":
		return Translate("min", map[string]interface{}{
			"field":  e.Field(),
			"number": e.Param(),
		}, c)
	case "max":
		return Translate("max", map[string]interface{}{
			"field":  e.Field(),
			"number": e.Param(),
		}, c)
	default:
		return Translate("invalid", map[string]interface{}{
			"field": e.Field(),
		}, c)
	}
}
