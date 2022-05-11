package server

import (
	"github.com/labstack/echo/v4"
	"github.com/pnp-zone/pkg-manager/conf"
)

func defineRoutes(e *echo.Echo, config *conf.Config) {

	e.Static("/static/", config.Server.StaticDir)
}
