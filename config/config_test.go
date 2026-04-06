package config

import (
	"os"
	"path/filepath"
	"testing"
)

func loadTestConfig(dir string) (*Config, error) {
	return LoadFrom(dir, "")
}

func TestInitLoadsNavContentFile(t *testing.T) {
	prev := App
	t.Cleanup(func() { App = prev })

	dir := t.TempDir()

	files := map[string]string{
		"app.yaml": `app:
  env: development
  name: starter
  database:
    url: postgres://postgres:postgres@db:5432/starter?sslmode=disable
  log:
    level: info
  server:
    host: 0.0.0.0
    port: 8080
    graceful_timeout: 10
  security:
    external_origin: http://localhost:3000
    origin:
      enabled: true
      require_header: true
      allowed_origins: []
    fetch_metadata:
      enabled: true
      allowed_sites: [same-origin, same-site]
      allowed_modes: [cors, navigate, same-origin]
      allowed_destinations: []
      fallback_when_missing: true
      reject_cross_site_navigate: true
    headers:
      xss_protection: "0"
      content_type_nosniff: true
      frame_options: DENY
      content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self'"
      hsts:
        enabled: false
        max_age: 31536000
        include_subdomains: true
        preload: false
    csrf:
      key: "12345678901234567890123456789012"
      context_key: csrf
      form_field: _csrf
      header_name: X-CSRF-Token
  csp_hashes:
    always: []
    dev_only: []
  session:
    name: _session
    auth_key: "12345678901234567890123456789012"
    encrypt_key: "12345678901234567890123456789012"
    max_age: 86400
    secure: false
`,
		"site.yaml": `site:
  title: starter
`,
		"service.yaml": `service:
  send:
    send_to: admin@example.com
    delivery:
      selected: log
      providers:
        log: {}
        noop: {}
        memory: {}
        resend:
          api_key: resend-key
          send_from: no-reply@example.com
        mailchannels:
          api_key: mc-key
          send_from: fallback@example.com
  auth:
    selected: email_totp
    methods:
      email_totp:
        enabled: true
        issuer: Forge
        period_seconds: 300
`,
		"nav.yaml": `nav:
  brand:
    label: Starter
    href: /
  sections:
    - items:
        - label: Home
          href: /
        - label: Explore
          items:
            - label: Components
              href: /_components
    - align: end
      items:
        - label: Signin
          href: /signin
          visibility: guest
        - slot: theme_toggle
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	cfg, err := loadTestConfig(dir)
	if err != nil {
		t.Fatalf("loadTestConfig() error = %v", err)
	}
	App = cfg

	if got := App.Nav.Brand.Label; got != "Starter" {
		t.Fatalf("InitConfig() brand label = %q, want %q", got, "Starter")
	}
	if len(App.Nav.Sections) != 2 {
		t.Fatalf("InitConfig() sections = %d, want 2", len(App.Nav.Sections))
	}
	if got := App.Nav.Sections[1].Align; got != "end" {
		t.Fatalf("InitConfig() section align = %q, want %q", got, "end")
	}
	if got := App.Nav.Sections[1].Items[1].Slot; got != "theme_toggle" {
		t.Fatalf("InitConfig() slot = %q, want %q", got, "theme_toggle")
	}
}

func TestInitAllowsNavShapesWithoutCustomCrossFieldValidation(t *testing.T) {
	prev := App
	t.Cleanup(func() { App = prev })

	tests := []struct {
		name string
		nav  string
	}{
		{
			name: "href and action conflict",
			nav: `nav:
  sections:
    - items:
        - label: Broken
          href: /broken
          action: /submit
`,
		},
		{
			name: "slot conflicts with href",
			nav: `nav:
  sections:
    - items:
        - slot: theme_toggle
          href: /broken
`,
		},
		{
			name: "method requires action",
			nav: `nav:
  sections:
    - items:
        - label: Broken
          method: post
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{
				"app.yaml": `app:
  env: development
  name: starter
  database:
    url: postgres://postgres:postgres@db:5432/starter?sslmode=disable
  log:
    level: info
  server:
    host: 0.0.0.0
    port: 8080
    graceful_timeout: 10
  security:
    external_origin: http://localhost:3000
    origin:
      enabled: true
      require_header: true
      allowed_origins: []
    fetch_metadata:
      enabled: true
      allowed_sites: [same-origin, same-site]
      allowed_modes: [cors, navigate, same-origin]
      allowed_destinations: []
      fallback_when_missing: true
      reject_cross_site_navigate: true
    headers:
      xss_protection: "0"
      content_type_nosniff: true
      frame_options: DENY
      content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self'"
      hsts:
        enabled: false
        max_age: 31536000
        include_subdomains: true
        preload: false
    csrf:
      key: "12345678901234567890123456789012"
      context_key: csrf
      form_field: _csrf
      header_name: X-CSRF-Token
  csp_hashes:
    always: []
    dev_only: []
  session:
    name: _session
    auth_key: "12345678901234567890123456789012"
    encrypt_key: "12345678901234567890123456789012"
    max_age: 86400
    secure: false
`,
				"site.yaml": `site:
  title: starter
`,
				"service.yaml": `service:
  send:
    send_to: admin@example.com
    delivery:
      selected: log
      providers:
        log: {}
        noop: {}
        memory: {}
        resend:
          api_key: resend-key
          send_from: no-reply@example.com
        mailchannels:
          api_key: mc-key
          send_from: fallback@example.com
  auth:
    selected: email_totp
    methods:
      email_totp:
        enabled: true
        issuer: Forge
        period_seconds: 300
`,
				"nav.yaml": tt.nav,
			}
			for name, contents := range files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error = %v", name, err)
				}
			}

			cfg, err := loadTestConfig(dir)
			if err != nil {
				t.Fatalf("loadTestConfig() error = %v, want config to load without custom nav validation", err)
			}
			if cfg == nil {
				t.Fatal("loadTestConfig() cfg = nil, want non-nil config")
			}
		})
	}
}

func TestLoadFromLoadsNestedServiceProviderConfig(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"app.yaml": `app:
  env: development
  name: starter
  database:
    url: postgres://postgres:postgres@db:5432/starter?sslmode=disable
  log:
    level: info
  server:
    host: 0.0.0.0
    port: 8080
    graceful_timeout: 10
  security:
    external_origin: http://localhost:3000
    origin:
      enabled: true
      require_header: true
      allowed_origins: []
    fetch_metadata:
      enabled: true
      allowed_sites: [same-origin, same-site]
      allowed_modes: [cors, navigate, same-origin]
      allowed_destinations: []
      fallback_when_missing: true
      reject_cross_site_navigate: true
    headers:
      xss_protection: "0"
      content_type_nosniff: true
      frame_options: DENY
      content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self'"
      hsts:
        enabled: false
        max_age: 31536000
        include_subdomains: true
        preload: false
    csrf:
      key: "12345678901234567890123456789012"
      context_key: csrf
      form_field: _csrf
      header_name: X-CSRF-Token
  csp_hashes:
    always: []
    dev_only: []
  session:
    name: _session
    auth_key: "12345678901234567890123456789012"
    encrypt_key: "12345678901234567890123456789012"
    max_age: 86400
    secure: false
`,
		"site.yaml": `site:
  title: starter
`,
		"service.yaml": `service:
  send:
    send_to: admin@example.com
    delivery:
      selected: resend
      providers:
        log: {}
        noop: {}
        memory: {}
        resend:
          api_key: resend-key
          send_from: no-reply@example.com
        mailchannels:
          api_key: mc-key
          send_from: fallback@example.com
  auth:
    selected: email_totp
    methods:
      email_totp:
        enabled: true
        issuer: Forge
        period_seconds: 300
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	cfg, err := LoadFrom(dir, "")
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if got := cfg.Service.Send.SendTo; got != "admin@example.com" {
		t.Fatalf("Service.Send.SendTo = %q, want %q", got, "admin@example.com")
	}
	if got := cfg.Service.Send.Delivery.SelectedProvider(); got != "resend" {
		t.Fatalf("Service.Send.Delivery.SelectedProvider() = %q, want %q", got, "resend")
	}
	if got := cfg.Service.Send.Delivery.Providers.Resend.SendFrom; got != "no-reply@example.com" {
		t.Fatalf("Service.Send.Delivery.Providers.Resend.SendFrom = %q, want %q", got, "no-reply@example.com")
	}
	if got := cfg.Service.Auth.SelectedMethod(); got != "email_totp" {
		t.Fatalf("Service.Auth.SelectedMethod() = %q, want %q", got, "email_totp")
	}
	if !cfg.Service.Auth.Methods.EmailTOTP.Enabled {
		t.Fatal("Service.Auth.Methods.EmailTOTP.Enabled = false, want true")
	}
	if got := cfg.Service.Auth.Methods.EmailTOTP.Issuer; got != "Forge" {
		t.Fatalf("Service.Auth.Methods.EmailTOTP.Issuer = %q, want %q", got, "Forge")
	}
	if got := cfg.Service.Auth.Methods.EmailTOTP.PeriodSeconds; got != 300 {
		t.Fatalf("Service.Auth.Methods.EmailTOTP.PeriodSeconds = %d, want %d", got, 300)
	}
}

func TestLoadFromLoadsCSRFSecurityTokenTTLSeconds(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"app.yaml": `app:
  env: development
  name: starter
  database:
    url: postgres://postgres:postgres@db:5432/starter?sslmode=disable
  log:
    level: info
  server:
    host: 0.0.0.0
    port: 8080
    graceful_timeout: 10
  security:
    external_origin: http://localhost:3000
    origin:
      enabled: true
      require_header: true
      allowed_origins: []
    fetch_metadata:
      enabled: true
      allowed_sites: [same-origin, same-site]
      allowed_modes: [cors, navigate, same-origin]
      allowed_destinations: []
      fallback_when_missing: true
      reject_cross_site_navigate: true
    headers:
      xss_protection: "0"
      content_type_nosniff: true
      frame_options: DENY
      content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self'"
      hsts:
        enabled: false
        max_age: 31536000
        include_subdomains: true
        preload: false
    csrf:
      key: "12345678901234567890123456789012"
      context_key: csrf
      form_field: _csrf
      header_name: X-CSRF-Token
      token_ttl: 3600
  csp_hashes:
    always: []
    dev_only: []
  session:
    name: _session
    auth_key: "12345678901234567890123456789012"
    encrypt_key: "12345678901234567890123456789012"
    max_age: 86400
    secure: false
`,
		"site.yaml": `site:
  title: starter
`,
		"service.yaml": `service:
  send:
    send_to: admin@example.com
    delivery:
      selected: log
      providers:
        log: {}
        noop: {}
        memory: {}
        resend:
          api_key: resend-key
          send_from: no-reply@example.com
        mailchannels:
          api_key: mc-key
          send_from: fallback@example.com
  auth:
    selected: email_totp
    methods:
      email_totp:
        enabled: true
        issuer: Forge
        period_seconds: 300
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	cfg, err := LoadFrom(dir, "")
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if got := cfg.App.Security.CSRF.TokenTTL; got != 3600 {
		t.Fatalf("App.Security.CSRF.TokenTTL = %d, want %d", got, 3600)
	}
}

func TestLoadFromRequiresServiceFile(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"app.yaml": `app:
  env: development
  name: starter
  database:
    url: postgres://postgres:postgres@db:5432/starter?sslmode=disable
  log:
    level: info
  server:
    host: 0.0.0.0
    port: 8080
    graceful_timeout: 10
  security:
    external_origin: http://localhost:3000
    origin:
      enabled: true
      require_header: true
      allowed_origins: []
    fetch_metadata:
      enabled: true
      allowed_sites: [same-origin, same-site]
      allowed_modes: [cors, navigate, same-origin]
      allowed_destinations: []
      fallback_when_missing: true
      reject_cross_site_navigate: true
    headers:
      xss_protection: "0"
      content_type_nosniff: true
      frame_options: DENY
      content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self'"
      hsts:
        enabled: false
        max_age: 31536000
        include_subdomains: true
        preload: false
    csrf:
      key: "12345678901234567890123456789012"
      context_key: csrf
      form_field: _csrf
      header_name: X-CSRF-Token
  csp_hashes:
    always: []
    dev_only: []
  session:
    name: _session
    auth_key: "12345678901234567890123456789012"
    encrypt_key: "12345678901234567890123456789012"
    max_age: 86400
    secure: false
`,
		"site.yaml": `site:
  title: starter
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	if _, err := LoadFrom(dir, ""); err == nil {
		t.Fatal("LoadFrom() error = nil, want missing required service.yaml error")
	}
}
