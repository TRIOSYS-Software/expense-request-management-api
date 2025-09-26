package main

import (
	"os"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/routes"

	_ "shwetaik-expense-management-api/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Expense Request System API
// @version 1.0
// @description This is a Expense Request System API Documentation.
// @termsOfService http://swagger.io/terms/

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey JWT Token
// @in header
// @name Authorization
// @description Type JWT token without Bearer prefix.

// @host localhost:1234
// @BasePath /api/v1
func main() {
	e := echo.New()

	e.Use(middleware.CORS())

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} ${method} ${uri} ${status} ${latency_human}` + "\n",
		Output: os.Stdout, // or a log file
	}))

	cfg := configs.Envs

	if err := cfg.ConnectDB(); err != nil {
		e.Logger.Fatal(err)
	}

	if err := cfg.SetupFirebase(); err != nil {
		e.Logger.Fatal(err)
	}
	cfg.InitializedDB()

	if cfg.Environment == "dev" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
	routes.InitialRoute(e, cfg.DB, cfg.FirebaseApp)

	e.Logger.Fatal(e.Start(cfg.ServerIP + ":" + cfg.ServerPort))
}
