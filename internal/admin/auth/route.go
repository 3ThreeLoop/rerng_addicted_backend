package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type AuthRoute struct {
	App         *fiber.App
	DBPool      *sqlx.DB
	AuthHandler *AuthHandler
}

func NewRoute(app *fiber.App, db_pool *sqlx.DB) *AuthRoute {
	return &AuthRoute{
		App:         app,
		DBPool:      db_pool,
		AuthHandler: NewAuthHandler(db_pool),
	}
}

func (au *AuthRoute) RegisterAuthRoute() *AuthRoute {
	auth := au.App.Group("/api/v1/admin/auth")

	auth.Post("/login", au.AuthHandler.Login)

	return au
}
