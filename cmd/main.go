package main

import (
	"os"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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

	cfg.InitializedDB()

	routes.InitialRoute(e, cfg.DB)

	e.Logger.Fatal(e.Start(cfg.ServerIP + ":" + cfg.ServerPort))
}
