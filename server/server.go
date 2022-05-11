package server

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/myOmikron/echotools/color"
	"github.com/myOmikron/echotools/execution"
	mw "github.com/myOmikron/echotools/middleware"
	"github.com/pelletier/go-toml"
	"github.com/pnp-zone/pkg-manager/conf"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func StartServer(configPath string) {
	config := &conf.Config{}

	if configBytes, err := ioutil.ReadFile(configPath); errors.Is(err, fs.ErrNotExist) {
		color.Printf(color.RED, "Config was not found at %s\n", configPath)
		b, _ := toml.Marshal(config)
		fmt.Print(string(b))
		os.Exit(1)
	} else {
		if err := toml.Unmarshal(configBytes, config); err != nil {
			panic(err)
		}
	}

	// Database
	db := initializeDatabase(config)

	e := echo.New()

	// Template rendering
	if !strings.HasSuffix(config.Server.TemplateDir, "/") {
		config.Server.TemplateDir += "/"
	}
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob(config.Server.TemplateDir + "*.gohtml")),
	}
	e.Renderer = renderer

	// Middleware definition
	e.Use(emw.Logger())
	e.Use(emw.Recover())

	allowedHosts := make([]mw.AllowedHost, 0)
	for _, host := range config.Server.AllowedHosts {
		allowedHosts = append(allowedHosts, mw.AllowedHost{
			Host:  host.Host,
			Https: host.Https,
		})
	}
	secConfig := &mw.SecurityConfig{
		AllowedHosts:            allowedHosts,
		UseForwardedProtoHeader: config.Server.UseForwardedProtoHeader,
	}
	e.Use(mw.Security(secConfig))

	cookieAge := time.Hour * 24
	e.Use(mw.Session(db, &mw.SessionConfig{
		CookieName:     "sessionid",
		CookieAge:      &cookieAge,
		CookiePath:     "/",
		DisableLogging: false,
	}))

	// Define routes
	defineRoutes(e, config)

	// Start server
	execution.SignalStart(e, config.Server.ListenAddress, &execution.Config{
		ReloadFunc: func() {
			StartServer(configPath)
		},
		StopFunc: func() {

		},
		TerminateFunc: func() {

		},
	})
}
