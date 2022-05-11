package main

import (
	"fmt"
	"github.com/hellflame/argparse"
	"github.com/pnp-zone/pkg-manager/server"
	"os"
)

func main() {
	parser := argparse.NewParser("pnp-zone-pkg-manager", "", &argparse.ParserConfig{DisableDefaultShowHelp: true})
	configPath := parser.String("", "config-path", &argparse.Option{
		Help:    "Specify an alternative path to the configuration file. Defaults to /etc/pnp-zone-pkg-manager/config.toml",
		Default: "/etc/pnp-zone-pkg-manager/config.toml",
	})

	if err := parser.Parse(nil); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	server.StartServer(*configPath)
}
