// Package config defines the application's configuration schema.
// Types are split by their YAML source: app.go, site.go, nav.go, service.go.
// Configuration is loaded at startup by internal/app.
package config

import (
	"fmt"
	"os"
	"strings"

	cfgs "github.com/go-sum/server/config"
)

// App is the global configuration singleton, initialised at startup.
var App *Config

// Environment returns c.App.Env lowercased, defaulting to "production" when empty.
func (c *Config) Environment() string {
	if c.App.Env != "" {
		return strings.ToLower(c.App.Env)
	}
	return "production"
}

// IsDevelopment reports whether the application is running in development mode.
func (c *Config) IsDevelopment() bool { return c.Environment() == "development" }

// IsProduction reports whether the application is running in production mode.
func (c *Config) IsProduction() bool { return c.Environment() == "production" }

// DSN returns the PostgreSQL connection string. If App.Database.URL is set
// (e.g. via YAML), it is returned as-is. Otherwise a DSN is built from the
// standard PG* environment variables (PGHOST, PGPORT, PGDATABASE, PGUSER,
// PGPASSWORD).
func (c *Config) DSN() string {
	if c.App.Database.URL != "" {
		return c.App.Database.URL
	}
	return buildDSN()
}

// buildDSN constructs a PostgreSQL DSN from standard PG* environment variables.
func buildDSN() string {
	host := envOr("PGHOST", "localhost")
	port := envOr("PGPORT", "5432")
	name := envOr("PGDATABASE", "postgres")
	user := envOr("PGUSER", "postgres")
	password := os.Getenv("PGPASSWORD")
	sslmode := envOr("PGSSLMODE", "disable")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, name, sslmode)
	return dsn
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load loads the application configuration from the default config/ directory.
// appEnv is typically os.Getenv("APP_ENV").
func Load(appEnv string) (*Config, error) {
	return LoadFrom("config", appEnv)
}

// LoadFrom loads configuration from the given directory.
// It is the primary entry point for both production and test use.
func LoadFrom(dir, appEnv string) (*Config, error) {
	return cfgs.Load(func(cfg *Config) cfgs.Options {
		return cfgs.Options{
			EnvKey: appEnv,
			Files: []cfgs.ConfigFile{
				{Filepath: dir + "/app.yaml", Required: true},
				{Filepath: dir + "/site.yaml", Required: true},
				{Filepath: dir + "/nav.yaml"},
				{Filepath: dir + "/service.yaml", Required: true},
			},
		}
	})
}
