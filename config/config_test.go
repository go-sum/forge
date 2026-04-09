package config

import (
	"strings"
	"testing"

	cfgs "github.com/go-sum/server/config"
)

// validSecrets sets the env vars required by the production default config.
// Without these, unrelated required/min validation failures would obscure the
// cross-field rule under test.
func validSecrets(t *testing.T) {
	t.Helper()
	t.Setenv("AUTH_SESSION_AUTH_KEY", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")   // 32 chars
	t.Setenv("AUTH_SESSION_ENCRYPT_KEY", "bbbbbbbbbbbbbbbb")                 // 16 chars
	t.Setenv("SECURITY_CSRF_KEY", "cccccccccccccccccccccccccccccccc")        // 32 chars
	t.Setenv("EXTERNAL_ORIGIN", "https://example.com")
	t.Setenv("EMAIL_SEND_TO", "test@example.com")
	t.Setenv("EMAIL_API_KEY", "test-api-key")
	t.Setenv("EMAIL_SEND_FROM", "noreply@example.com")
}

func TestLoadDevelopmentValid(t *testing.T) {
	validSecrets(t)
	if _, err := Load("development"); err != nil {
		t.Fatalf("Load(development) unexpected error: %v", err)
	}
}

func TestLoadServerStoreWithoutKVFails(t *testing.T) {
	validSecrets(t)
	_, err := cfgs.Load(productionDefault, func(c *Config) {
		c.Session.Auth.Store = "server"
		c.Store.KV.Enabled = false
	})
	if err == nil {
		t.Fatal("Load() error = nil, want requires_kv validation error")
	}
	if !strings.Contains(err.Error(), "requires_kv") {
		t.Fatalf("Load() error = %q, want it to contain %q", err.Error(), "requires_kv")
	}
}

func TestLoadCookieStoreWithoutKVPasses(t *testing.T) {
	validSecrets(t)
	if _, err := Load(""); err != nil {
		t.Fatalf("Load() unexpected error for cookie store without KV: %v", err)
	}
}

func TestLoadServerStoreWithKVPasses(t *testing.T) {
	validSecrets(t)
	_, err := cfgs.Load(productionDefault, func(c *Config) {
		c.Session.Auth.Store = "server"
		c.Store.KV.Enabled = true
	})
	if err != nil {
		t.Fatalf("Load() unexpected error for server store with KV enabled: %v", err)
	}
}

func TestLoadNavValidationFails(t *testing.T) {
	validSecrets(t)
	_, err := cfgs.Load(productionDefault, func(c *Config) {
		// MatchPrefix without Href violates nav cross-field rule.
		c.Nav.Sections = []NavSection{
			{Items: []NavItem{{Label: "Home", MatchPrefix: true, Href: ""}}},
		}
	})
	if err == nil {
		t.Fatal("Load() error = nil, want nav validation error for MatchPrefix without Href")
	}
}
