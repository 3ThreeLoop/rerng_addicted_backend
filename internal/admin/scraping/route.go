package scraping

import (
	"rerng_addicted_api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ScrapingRoute struct {
	App             *fiber.App
	DBPool          *sqlx.DB
	ScrapingHandler *ScrapingHandler
}

func NewRoute(app *fiber.App, db_pool *sqlx.DB) *ScrapingRoute {
	return &ScrapingRoute{
		App:             app,
		DBPool:          db_pool,
		ScrapingHandler: NewScrapingHandler(db_pool),
	}
}

func (sc *ScrapingRoute) RegisterScrapingRoute() *ScrapingRoute {
	scraping := sc.App.Group("/api/v1/admin/scraping")

	scraping.Get("/search", middlewares.NewJwtMiddleware(sc.DBPool), sc.ScrapingHandler.Search)
	scraping.Get("/detail/:key", middlewares.NewJwtMiddleware(sc.DBPool), sc.ScrapingHandler.GetDetail)
	scraping.Get("/deep/detail/:key", middlewares.NewJwtMiddleware(sc.DBPool), sc.ScrapingHandler.GetDeepDetail)

	return sc
}
