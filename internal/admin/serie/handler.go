package serie

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	share "rerng_addicted_api/pkg/model"
)

type SerieHandler struct {
	DBPool          *sqlx.DB
	SerieService func(c *fiber.Ctx) *SerieService
}

func NewSerieHandler(db_pool *sqlx.DB) *SerieHandler {
	return &SerieHandler{
		DBPool: db_pool,
		SerieService: func(c *fiber.Ctx) *SerieService {
			var uCtx share.UserContext
			uCtx, ok := c.Locals("UserContext").(share.UserContext)
			if !ok {
				uCtx = share.UserContext{}
			}
			return NewSerieService(db_pool, &uCtx)
		},
	}
}
