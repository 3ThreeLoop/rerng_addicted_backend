package auth

import (
	custom_log "rerng_addicted_api/pkg/logs"
	"rerng_addicted_api/pkg/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LoginRequest struct {
	UserName string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (au *LoginRequest) Bind(c *fiber.Ctx, v *utils.Validator) error {
	if err := c.BodyParser(au); err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		return fmt.Errorf("%s", utils.Translate("invalid_body", nil, c))
	}

	if err := v.Validate(au, c); err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		return err
	}

	return nil
}

type LoginResponse struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

type User struct {
	UserUUID uuid.UUID `json:"user_uuid" db:"user_uuid"`
}
type UserInfo struct {
	ID           int    `json:"id" db:"id"`
	UserUUID     string `json:"user_uuid" db:"user_uuid"`
	UserName     string `json:"user_name" db:"user_name"`
	RoleID       int    `json:"role_id" db:"role_id"`
	LoginSession string `json:"login_session" db:"login_session"`
	StatusID     int    `json:"status_id" db:"status_id"`
}
