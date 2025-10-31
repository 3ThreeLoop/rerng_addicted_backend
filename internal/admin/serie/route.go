package serie

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type SerieRoute struct {
	App             *fiber.App
	DBPool          *sqlx.DB
	SerieHandler *SerieHandler
}

func NewRoute(app *fiber.App, db_pool *sqlx.DB) *SerieRoute {
	return &SerieRoute{
		App:             app,
		DBPool:          db_pool,
		SerieHandler: NewSerieHandler(db_pool),
	}
}

func (r *SerieRoute) RegisterSerieRoute() *SerieRoute {
	//group := r.App.Group("/api/v1/admin/serie")

	// TODO: add your routes here
	// group.Get("/search", middlewares.NewJwtMiddleware(r.DBPool), r.SerieHandler.Search)

	return r
}
