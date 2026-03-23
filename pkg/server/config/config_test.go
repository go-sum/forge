package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knadh/koanf/v2"
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
	writeLoaderFile(t, dir, "config.yaml", `
app:
  env: production
auth:
  session:
    auth_key: base-key
site:
  title: base-title
`)
	writeLoaderFile(t, dir, "config.development.yaml", `
auth:
  session:
    auth_key: overlay-key
`)
	writeLoaderFile(t, dir, "site.yaml", `
site:
  title: content-title
`)

	t.Setenv("CTX_APP_ENV", "development")
	t.Setenv("CTX_AUTH_SESSION_AUTH_KEY", "env-key")
	t.Setenv("CTX_SITE_TITLE", "env-title")

	var cfg loaderTestConfig
	err := loadConfig(&cfg, Options{
		EnvPrefix: "CTX_",
		BaseDir:   dir,
		EnvKey:    "app.env",
		ContentFiles: []ContentFile{{
			Filename: "site.yaml",
			Target:   &cfg.Site,
		}},
	})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("App.Env = %q", cfg.App.Env)
	}
	if cfg.Auth.Session.AuthKey != "env-key" {
		t.Fatalf("AuthKey = %q", cfg.Auth.Session.AuthKey)
	}
	if cfg.Site.Title != "content-title" {
		t.Fatalf("Site.Title = %q", cfg.Site.Title)
	}
}

func TestLoadReportsInvalidContentFile(t *testing.T) {
	dir := t.TempDir()
	writeLoaderFile(t, dir, "config.yaml", `
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
		EnvPrefix: "CTX_",
		BaseDir:   dir,
		ContentFiles: []ContentFile{{
			Filename: "site.yaml",
			Target:   &cfg.Site,
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "site.yaml") {
		t.Fatalf("err = %v", err)
	}
}

func TestTransformKeyFallsBackToDottedPath(t *testing.T) {
	if got := transformKey(koanf.New("."), "auth_session_auth_key"); got != "auth.session.auth.key" {
		t.Fatalf("transformKey() = %q", got)
	}
}

func writeLoaderFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}
