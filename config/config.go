package config

import (
	uilayout "starter/pkg/components/ui/layout"
	pkgconfig "starter/pkg/config"
)

// envPrefix is stripped from environment variable names before key mapping.
const envPrefix = "CTX_"

// App is the global configuration singleton, populated by Init().
var App *Config

type CSPHashesConfig struct {
	Always  []string `koanf:"always"`
	DevOnly []string `koanf:"dev_only"`
}

type Config struct {
	App       AppConfig          `koanf:"app"`
	Server    ServerConfig       `koanf:"server"`
	Database  DatabaseConfig     `koanf:"database"`
	Auth      AuthConfig         `koanf:"auth"`
	Log       LogConfig          `koanf:"log"`
	Site      SiteConfig         `koanf:"site"`
	Nav       uilayout.NavConfig `koanf:"nav"`
	CSPHashes CSPHashesConfig    `koanf:"csp_hashes"`
}

type AppConfig struct {
	Env  string `koanf:"env"  validate:"required,oneof=development production test"`
	Name string `koanf:"name" validate:"required"`
}

type ServerConfig struct {
	Host            string `koanf:"host"             validate:"required"`
	Port            int    `koanf:"port"             validate:"required,min=1,max=65535"`
	GracefulTimeout int    `koanf:"graceful_timeout"`
	CSP             string `koanf:"csp"              validate:"required"`
	CSRFCookieName  string `koanf:"csrf_cookie_name" validate:"required"`
}

type DatabaseConfig struct {
	URL string `koanf:"url" validate:"required"`
}

type AuthConfig struct {
	JWT     JWTConfig     `koanf:"jwt"`
	Session SessionConfig `koanf:"session"`
}

type JWTConfig struct {
	Secret        string `koanf:"secret"         validate:"required,min=32"`
	Issuer        string `koanf:"issuer"`
	TokenDuration int    `koanf:"token_duration" validate:"required,min=1"`
}

type SessionConfig struct {
	Name       string `koanf:"name"`
	AuthKey    string `koanf:"auth_key"    validate:"required,min=32"`
	EncryptKey string `koanf:"encrypt_key" validate:"required,min=32,max=32"`
	MaxAge     int    `koanf:"max_age"`
	Secure     bool   `koanf:"secure"`
}

type LogConfig struct {
	Level string `koanf:"level" validate:"required,oneof=debug info warn error"`
}

type SiteConfig struct {
	Title        string   `koanf:"title"         validate:"required"`
	Description  string   `koanf:"description"`
	LogoPath     string   `koanf:"logo_path"`
	FaviconPath  string   `koanf:"favicon_path"`
	MetaKeywords []string `koanf:"meta_keywords"`
	OGImage      string   `koanf:"og_image"`
}

// Init populates the App singleton from YAML files and CTX_-prefixed env vars in baseDir.
func Init(baseDir string) error {
	cfg := &Config{}
	if err := pkgconfig.Load(cfg, pkgconfig.Options{
		EnvPrefix:      envPrefix,
		BaseDir:        baseDir,
		EnvKey:         "app.env",
		ValidatorSetup: uilayout.RegisterNavValidations,
		ContentFiles: []pkgconfig.ContentFile{
			{Filename: "site.yaml", Target: &cfg.Site},
			{Filename: "nav.yaml", Target: &cfg.Nav},
		},
	}); err != nil {
		return err
	}
	App = cfg
	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// DSN is an alias for Database.URL.
func (c *Config) DSN() string {
	return c.Database.URL
}
