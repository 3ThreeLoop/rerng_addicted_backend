package handler

import (
	"rerng_addicted_api/internal/admin/auth"
	scraping "rerng_addicted_api/internal/admin/scraping"
	auth_front "rerng_addicted_api/internal/front/auth"
	"rerng_addicted_api/internal/front/user"
	"rerng_addicted_api/internal/shared/proxy"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// group all the module factories
type ServiceHandler struct {
	Front  *FrontService
	Admin  *AdminService
	Shared *SharedService
}

// register modules route to front service
type FrontService struct {
	AuthRoute *auth_front.AuthRoute
	UserRoute *user.UserRoute
}

// register modules route to admin service
type AdminService struct {
	AuthRoute     *auth.AuthRoute
	ScrapingRoute *scraping.ScrapingRoute
}

type SharedService struct {
	ProxyRoute *proxy.ProxyRoute
}

func NewFrontService(app *fiber.App, db_pool *sqlx.DB) *FrontService {
	au := auth_front.NewRoute(app, db_pool).RegisterAuthRoute()
	user := user.NewUserRoute(app, db_pool).RegisterUserRoute()

	return &FrontService{
		AuthRoute: au,
		UserRoute: user,
	}
}

func NewAdminService(app *fiber.App, db_pool *sqlx.DB) *AdminService {
	au := auth.NewRoute(app, db_pool).RegisterAuthRoute()
	sc := scraping.NewRoute(app, db_pool).RegisterScrapingRoute()

	return &AdminService{
		AuthRoute:     au,
		ScrapingRoute: sc,
	}
}

func NewSharedService(app *fiber.App, db_pool *sqlx.DB) *SharedService {
	pr := proxy.NewRoute(app, db_pool).RegisterProxyRoute()

	return &SharedService{
		ProxyRoute: pr,
	}
}

func NewServiceHandlers(app *fiber.App, db_pool *sqlx.DB) *ServiceHandler {
	front := NewFrontService(app, db_pool)
	admin := NewAdminService(app, db_pool)
	shared := NewSharedService(app, db_pool)

	return &ServiceHandler{
		Front:  front,
		Admin:  admin,
		Shared: shared,
	}
}
