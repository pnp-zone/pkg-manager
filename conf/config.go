package conf

import (
	"errors"
	"strings"
)

type Database struct {
	Driver   string
	Port     uint16
	Host     string
	Name     string
	User     string
	Password string
}

type AllowedHost struct {
	Host  string
	Https bool
}

type Server struct {
	ListenAddress           string
	AllowedHosts            []AllowedHost
	UseForwardedProtoHeader bool
	TemplateDir             string
	StaticDir               string
	PGPDir                  string
	PkgDir                  string
}

type Config struct {
	Server   Server
	Database Database
}

func (c *Config) Fix() error {
	if c.Server.StaticDir == "" {
		return errors.New("StaticDir must not be empty")
	} else if !strings.HasSuffix(c.Server.StaticDir, "/") {
		c.Server.StaticDir += "/"
	}

	if c.Server.TemplateDir == "" {
		return errors.New("TemplateDir must not be empty")
	} else if !strings.HasSuffix(c.Server.TemplateDir, "/") {
		c.Server.TemplateDir += "/"
	}

	if c.Server.PGPDir == "" {
		return errors.New("PGPDir must not be empty")
	} else if !strings.HasSuffix(c.Server.PGPDir, "/") {
		c.Server.PGPDir += "/"
	}

	if c.Server.PkgDir == "" {
		return errors.New("PkgDir must not be empty")
	} else if !strings.HasSuffix(c.Server.PkgDir, "/") {
		c.Server.PkgDir += "/"
	}

	if len(c.Server.AllowedHosts) == 0 {
		return errors.New("AllowedHosts must not be empty")
	}

	return nil
}
