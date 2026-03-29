package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type loaderTestConfig struct {
	App struct {
		Env string `koanf:"env" validate:"required,oneof=development production"`
	} `koanf:"app"`
	Auth struct {
		Session struct {
			AuthKey string `koanf:"auth_key" validate:"required"`
		} `koanf:"session"`
	} `koanf:"auth"`
	Site struct {
		Title string `koanf:"title" validate:"required"`
	} `koanf:"site"`
}

func TestLoadMergesOverlayEnvAndContentFiles(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: ${APP_ENV:-production}
auth:
  session:
    auth_key: ${TEST_AUTH_KEY}
site:
  title: base-title
`)
	// Overlay sets app.env so cfg.App.Env reflects the active environment.
	writeLoaderFile(t, dir, "app.development.yaml", `
app:
  env: development
`)
	// Content file wins over base and overlay values.
	writeLoaderFile(t, dir, "site.yaml", `
site:
  title: content-title
`)

	t.Setenv("APP_ENV", "development")
	t.Setenv("TEST_AUTH_KEY", "env-key")

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{
		EnvKey: os.Getenv("APP_ENV"),
		Files: []ConfigFile{
			{Filepath: filepath.Join(dir, "app.yaml")},
			{Filepath: filepath.Join(dir, "site.yaml")},
		},
	})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("App.Env = %q, want development", cfg.App.Env)
	}
	if cfg.Auth.Session.AuthKey != "env-key" {
		t.Fatalf("AuthKey = %q, want env-key", cfg.Auth.Session.AuthKey)
	}
	if cfg.Site.Title != "content-title" {
		t.Fatalf("Site.Title = %q, want content-title", cfg.Site.Title)
	}
}

func TestLoadReportsInvalidMergedConfig(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: production
auth:
  session:
    auth_key: base-key
site:
  title: base-title
`)
	writeLoaderFile(t, dir, "site.yaml", `
site:
  title: ""
`)

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{
		Files: []ConfigFile{
			{Filepath: filepath.Join(dir, "app.yaml")},
			{Filepath: filepath.Join(dir, "site.yaml")},
		},
	})
	if err == nil {
		t.Fatal("loadConfig() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "config: validation:") {
		t.Fatalf("err = %v, want root validation error", err)
	}
}

func TestLoadAllowsMissingOptionalFileOnly(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: production
auth:
  session:
    auth_key: base-key
site:
  title: base-title
`)

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{
		Files: []ConfigFile{
			{Filepath: filepath.Join(dir, "app.yaml"), Required: true},
			{Filepath: filepath.Join(dir, "nav.yaml")},
		},
	})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
}

func TestLoadFailsWhenRequiredFileIsMissing(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: production
auth:
  session:
    auth_key: base-key
site:
  title: base-title
`)

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{
		Files: []ConfigFile{
			{Filepath: filepath.Join(dir, "app.yaml"), Required: true},
			{Filepath: filepath.Join(dir, "site.yaml"), Required: true},
		},
	})
	if err == nil {
		t.Fatal("loadConfig() error = nil, want missing required file error")
	}
	if !strings.Contains(err.Error(), "site.yaml") {
		t.Fatalf("err = %v, want missing required path", err)
	}
}

func TestLoadFailsOnInvalidOptionalFileContent(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: production
auth:
  session:
    auth_key: base-key
site:
  title: base-title
`)
	writeLoaderFile(t, dir, "site.yaml", `site: [`)

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{
		Files: []ConfigFile{
			{Filepath: filepath.Join(dir, "app.yaml"), Required: true},
			{Filepath: filepath.Join(dir, "site.yaml")},
		},
	})
	if err == nil {
		t.Fatal("loadConfig() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "site.yaml") {
		t.Fatalf("err = %v, want invalid file path", err)
	}
}

func TestEnvVarExpansionInYAML(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: production
auth:
  session:
    auth_key: ${TEST_EXPAND_KEY}
site:
  title: fixed-title
`)

	t.Setenv("TEST_EXPAND_KEY", "injected-value")

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{Files: []ConfigFile{{Filepath: filepath.Join(dir, "app.yaml")}}})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.Auth.Session.AuthKey != "injected-value" {
		t.Fatalf("AuthKey = %q, want injected-value", cfg.Auth.Session.AuthKey)
	}
	// Unset variable expands to empty string, not the literal placeholder.
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: production
auth:
  session:
    auth_key: ${TEST_UNSET_VAR}
site:
  title: fixed-title
`)
	var cfg2 loaderTestConfig
	_ = loadConfig(&cfg2, Options{Files: []ConfigFile{{Filepath: filepath.Join(dir, "app.yaml")}}}) // will fail validation (required), that's fine
	if cfg2.Auth.Session.AuthKey != "" {
		t.Fatalf("unset AuthKey = %q, want empty string", cfg2.Auth.Session.AuthKey)
	}
}

func TestEnvVarDefaultExpansionInYAML(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "app.yaml", `
app:
  env: ${TEST_ENV_VAR:-production}
auth:
  session:
    auth_key: ${TEST_KEY:-fallback-key}
site:
  title: fixed-title
`)

	// With env var set, it wins over the default.
	t.Setenv("TEST_ENV_VAR", "development")
	t.Setenv("TEST_KEY", "real-key")

	var cfg loaderTestConfig
	if err := loadConfig(&cfg, Options{Files: []ConfigFile{{Filepath: filepath.Join(dir, "app.yaml")}}}); err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("App.Env = %q, want development", cfg.App.Env)
	}
	if cfg.Auth.Session.AuthKey != "real-key" {
		t.Fatalf("AuthKey = %q, want real-key", cfg.Auth.Session.AuthKey)
	}

	// Without env vars, defaults are used.
	t.Setenv("TEST_ENV_VAR", "")
	t.Setenv("TEST_KEY", "")

	var cfg2 loaderTestConfig
	if err := loadConfig(&cfg2, Options{Files: []ConfigFile{{Filepath: filepath.Join(dir, "app.yaml")}}}); err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg2.App.Env != "production" {
		t.Fatalf("App.Env = %q, want production", cfg2.App.Env)
	}
	if cfg2.Auth.Session.AuthKey != "fallback-key" {
		t.Fatalf("AuthKey = %q, want fallback-key", cfg2.Auth.Session.AuthKey)
	}
}

func writeLoaderFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}
