package config

// Config is the root application configuration struct.
type Config struct {
	App       AppConfig          `koanf:"app"`
	Server    ServerConfig       `koanf:"server"`
	Database  DatabaseConfig     `koanf:"database"`
	Auth      AuthConfig         `koanf:"auth"`
	Log       LogConfig          `koanf:"log"`
	Site      SiteConfig         `koanf:"site"`
	Nav       NavConfig          `koanf:"nav"`
	CSPHashes CSPHashesConfig    `koanf:"csp_hashes"`
	Keys      ContextKeysConfig  `koanf:"keys"`
}

// IsDevelopment reports whether the application is running in development mode.
func (c *Config) IsDevelopment() bool { return c.App.Env == "development" }

// IsProduction reports whether the application is running in production mode.
func (c *Config) IsProduction() bool { return c.App.Env == "production" }

// DSN is an alias for Database.URL.
func (c *Config) DSN() string { return c.Database.URL }

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
	Session SessionConfig `koanf:"session"`
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
	Title         string   `koanf:"title"          validate:"required"`
	Description   string   `koanf:"description"`
	LogoPath      string   `koanf:"logo_path"`
	FaviconPath   string   `koanf:"favicon_path"`
	MetaKeywords  []string `koanf:"meta_keywords"`
	OGImage       string   `koanf:"og_image"`
	CopyrightYear int      `koanf:"copyright_year"`
}

type CSPHashesConfig struct {
	Always  []string `koanf:"always"`
	DevOnly []string `koanf:"dev_only"`
}

// ContextKeysConfig defines the Echo context key names written by auth middleware
// and read by the view layer. Override in config.yaml under the keys: node.
type ContextKeysConfig struct {
	UserID      string `koanf:"user_id"`
	UserRole    string `koanf:"user_role"`
	DisplayName string `koanf:"display_name"`
	CSRF        string `koanf:"csrf"`
}
