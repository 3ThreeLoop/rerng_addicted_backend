package main

import (
	"fmt"
	"rerng_addicted_api/configs"
	"rerng_addicted_api/db/postgresql"
	"rerng_addicted_api/handler"
	"rerng_addicted_api/pkg/logs"
	"rerng_addicted_api/pkg/redis"
	"rerng_addicted_api/pkg/swagger"
	"rerng_addicted_api/router"
)

// @title       Rerng Addicted API
// @version     1.0.0
// @description Professional API documentation for the Rerng Addicted backend
// @BasePath    /api/v1

// @schemes     http
func main() {
	// load environment variable from .env file
	app_configs := configs.NewAppConfig()

	// log
	log_level := "info"
	logs.NewLog(log_level)

	// init postgresql database and connection pool
	pool, err := postgresql.ConnectDB()
	if err != nil {
		fmt.Println("Error connect database : ", err)
	}

	// init redis
	_ = redis.NewRedis()

	// init go fiber framework, cors and handler configuration
	apps := router.New()

	// swagger
	swagger.Setup(apps, app_configs.AppHost, app_configs.AppPort)

	// init router
	handler.NewServiceHandlers(apps, pool)

	// http server
	err = apps.Listen(fmt.Sprintf("%s:%d", app_configs.AppHost, app_configs.AppPort))
	if err != nil {
		fmt.Printf("%v", err)
	}
}
