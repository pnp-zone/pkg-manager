package server

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/labstack/echo/v4"
	"github.com/myOmikron/echotools/worker"
	"github.com/pnp-zone/pkg-manager/conf"
	"github.com/pnp-zone/pkg-manager/handler/frontend"
	"github.com/pnp-zone/pkg-manager/handler/frontendapi"
	"gorm.io/gorm"
)

func defineRoutes(e *echo.Echo, db *gorm.DB, pool worker.Pool, keyring *crypto.KeyRing, config *conf.Config) {
	f := frontend.Wrapper{
		DB:     db,
		Config: config,
	}

	e.GET("/", f.Index)
	e.GET("/login", f.Login)
	e.GET("/register", f.Register)

	fa := frontendapi.Wrapper{
		DB:      db,
		Config:  config,
		Keyring: keyring,
	}
	e.POST("/frontend/register", fa.Register)
	e.POST("/frontend/login", fa.Login)

	e.Static("/static/", config.Server.StaticDir)
	e.Static("/packages/", config.Server.PkgDir)
	e.Static("/keys/", config.Server.PGPDir+"keys/")
}
