package server

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/myOmikron/echotools/color"
	"github.com/myOmikron/echotools/execution"
	mw "github.com/myOmikron/echotools/middleware"
	"github.com/myOmikron/echotools/worker"
	"github.com/pelletier/go-toml"
	"github.com/pnp-zone/pkg-manager/conf"
	"github.com/pnp-zone/pkg-manager/models"
	"github.com/pnp-zone/pkg-manager/task"
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

	if err := config.Fix(); err != nil {
		color.Println(color.RED, "Config error:")
		color.Println(color.RED, err.Error())
		os.Exit(1)
	}

	// Directory structure generation
	fmt.Print("Creating directory structure ... ")
	if err := os.MkdirAll(config.Server.PGPDir, 0700); err != nil {
		color.Println(color.RED, "error")
		color.Println(color.RED, "Could not create PGPDir")
		color.Println(color.RED, err.Error())
		os.Exit(1)
	}
	if err := os.MkdirAll(config.Server.PGPDir+"keys/", 0700); err != nil {
		color.Println(color.RED, "error")
		color.Println(color.RED, "Could not create keys/ dir in PGPDir")
		color.Println(color.RED, err.Error())
		os.Exit(1)
	}
	if err := os.MkdirAll(config.Server.PkgDir, 0700); err != nil {
		color.Println(color.RED, "error")
		color.Println(color.RED, "Could not create PkgDir")
		color.Println(color.RED, err.Error())
		os.Exit(1)
	}
	if err := os.MkdirAll(config.Server.PkgDir+"packages/", 0700); err != nil {
		color.Println(color.RED, "error")
		color.Println(color.RED, "Could not create packages/ directory in PkgDir")
		color.Println(color.RED, err.Error())
		os.Exit(1)
	}
	color.Println(color.GREEN, "done")

	// Database
	fmt.Print("Initializing database ... ")
	db := initializeDatabase(config)
	color.Println(color.GREEN, "done")

	// Worker pool
	pool := worker.NewPool(&worker.PoolConfig{
		NumWorker: 20,
		QueueSize: 200,
	})
	pool.Start()

	// OpenPGP key generation
	if _, err := os.Stat(config.Server.PGPDir + "pkg.asc"); os.IsNotExist(err) {
		color.Println(color.PURPLE, "Missing key for package manager: "+config.Server.PGPDir+"pkg.asc")
		fmt.Print("Generating new key ... ")

		if key, err := crypto.GenerateKey("pnp-zone-pkg-manager", "pkg@pnp.zone", "rsa", 4096); err != nil {
			color.Println(color.RED, "error")
			color.Println(color.RED, "Could not create key")
			os.Exit(1)
		} else {
			armor, err := key.Armor()
			if err != nil {
				color.Println(color.RED, "error")
				color.Println(color.RED, "Could not get armored version of key")
				os.Exit(1)
			}
			if err := ioutil.WriteFile(config.Server.PGPDir+"pkg.asc", []byte(armor), 0600); err != nil {
				color.Println(color.RED, "error")
				color.Println(color.RED, "Could not write pkg manager key")
				os.Exit(1)
			}
		}

		color.Println(color.GREEN, "done")
	}

	var ring *crypto.KeyRing
	fmt.Print("Importing key for package manager ... ")
	if file, err := ioutil.ReadFile(config.Server.PGPDir + "pkg.asc"); err != nil {
		color.Println(color.RED, "error")
		color.Println(color.RED, err.Error())
		os.Exit(1)
	} else {
		if armored, err := crypto.NewKeyFromArmored(string(file)); err != nil {
			color.Println(color.RED, "error")
			color.Println(color.RED, "Could not retrieve key from file:")
			color.Println(color.RED, err.Error())
			os.Exit(1)
		} else {
			color.Println(color.GREEN, "done")

			fmt.Print("Creating keyring ... ")
			if keyRing, err := crypto.NewKeyRing(armored); err != nil {
				color.Println(color.RED, "error")
				color.Println(color.RED, "Could not create keyring from retrieved key:")
				color.Println(color.RED, err.Error())
				os.Exit(1)
			} else {
				color.Println(color.GREEN, "done")
				ring = keyRing
			}
		}
	}

	maintainer := make([]models.Maintainer, 0)
	db.Find(&maintainer)
	removeAll := false
	for _, mt := range maintainer {
		filename := config.Server.PGPDir + "keys/" + mt.Fingerprint
		if content, err := ioutil.ReadFile(filename); err != nil {
			if removeAll {
				color.Printf(color.CYAN, "Skipping %s\n", mt.Fingerprint)
				continue
			}
			color.Printf(color.RED, "Error retrieving key: %s\n", filename)
			color.Println(color.RED, err.Error())
			reader := bufio.NewReader(os.Stdin)
			for {
				color.Print(color.PURPLE, "Skip key? [y/n/a] ")
				line, _, err := reader.ReadLine()
				if err != nil {
					color.Println(color.PURPLE, "Invalid choice.")
					continue
				}
				if string(line) == "n" {
					os.Exit(1)
				} else if string(line) == "y" {
					color.Printf(color.CYAN, "Skipping %s\n", mt.Fingerprint)
					break
				} else if string(line) == "a" {
					color.Printf(color.CYAN, "Skipping %s\n", mt.Fingerprint)
					removeAll = true
					break
				} else {
					color.Println(color.PURPLE, "Invalid choice.")
					continue
				}
			}
		} else {
			if armored, err := crypto.NewKeyFromArmored(string(content)); err != nil {
				color.Println(color.RED, "error")
				color.Printf(color.RED, "Error retrieving key from file: %s\n", filename)
				color.Println(color.RED, err.Error())
				os.Exit(1)
			} else {
				if err := ring.AddKey(armored); err != nil {
					color.Println(color.RED, "error")
					color.Printf(color.RED, "Error adding key to keyring: %s\n", filename)
					color.Println(color.RED, err.Error())
					os.Exit(1)
				}
			}
		}
	}

	// Webserver
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

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
	defineRoutes(e, db, pool, ring, config)

	// Start generation of index
	go task.BuildIndex(db, ring, config)

	// Start server
	fmt.Println("Server is listening on:", color.Colorize(color.CYAN, config.Server.ListenAddress))
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
