package proxy

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ProxyRoute struct {
	App          *fiber.App
	DBPool       *sqlx.DB
	ProxyHandler *ProxyHandler
}

func NewRoute(app *fiber.App, db_pool *sqlx.DB) *ProxyRoute {
	return &ProxyRoute{
		App:          app,
		DBPool:       db_pool,
		ProxyHandler: NewProxyHandler(db_pool),
	}
}

func (pr *ProxyRoute) RegisterProxyRoute() *ProxyRoute {
	proxy := pr.App.Group("/api/v1/admin/proxy")

	proxy.Get("/m3u8/*", pr.ProxyHandler.M3u8)

	proxy.Get("/mp4", pr.ProxyHandler.Mp4)

	proxy.Get("/subtitle/*", pr.ProxyHandler.Subtitle)

	proxy.Get("/download", pr.ProxyHandler.Download)

	return pr
}
