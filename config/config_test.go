package config

import (
	"strings"
	"testing"

	auth "github.com/go-sum/auth"
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

func TestAuthDefaultPeriodSecondsApplied(t *testing.T) {
	validSecrets(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	// PeriodSeconds is no longer set in config/service.go; it comes from
	// pkg/auth defaultConfig via ApplyDefaults called in initAuth.
	// At load time we just confirm the value stored in the config literal is 0
	// (default not applied until runtime boundary), which is the expected state.
	// The real coverage: auth.ApplyDefaults fills it to 300.
	normalized := auth.ApplyDefaults(cfg.Service.Auth)
	if normalized.Methods.EmailTOTP.PeriodSeconds != 300 {
		t.Errorf("ApplyDefaults PeriodSeconds = %d, want 300", normalized.Methods.EmailTOTP.PeriodSeconds)
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
