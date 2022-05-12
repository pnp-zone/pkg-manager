package server

import (
	"github.com/labstack/echo/v4"
	"github.com/pnp-zone/pkg-manager/conf"
	"github.com/pnp-zone/pkg-manager/handler/frontend"
	"github.com/pnp-zone/pkg-manager/handler/frontendapi"
	"gorm.io/gorm"
)

func defineRoutes(e *echo.Echo, db *gorm.DB, config *conf.Config) {
	f := frontend.Wrapper{
		DB:     db,
		Config: config,
	}

	e.GET("/", f.Index)
	e.GET("/login", f.Login)
	e.GET("/register", f.Register)

	fa := frontendapi.Wrapper{
		DB:     db,
		Config: config,
	}
	e.POST("/frontend/register", fa.Register)
	e.POST("/frontend/login", fa.Login)

	e.Static("/static/", config.Server.StaticDir)
}
