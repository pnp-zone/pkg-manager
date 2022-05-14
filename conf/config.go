package conf

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
