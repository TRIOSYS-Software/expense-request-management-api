package main

import (
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/routes"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	configs := configs.Config{}
	cfg := configs.LoadEnv(".env")
	err := cfg.ConnectDB()
	if err != nil {
		e.Logger.Fatal(err)
	}

	cfg.InitializedDB()

	routes.InitialRoute(e, cfg.DB)

	e.Logger.Fatal(e.Start(cfg.ServerIP + ":" + cfg.ServerPort))
}
