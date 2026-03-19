package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLoadsNavContentFile(t *testing.T) {
	prev := App
	t.Cleanup(func() { App = prev })

	dir := t.TempDir()

	files := map[string]string{
		"config.yaml": `app:
  env: development
  name: starter
database:
  url: postgres://postgres:postgres@app_data:5432/starter?sslmode=disable
log:
  level: info
server:
  host: 0.0.0.0
  port: 8080
  graceful_timeout: 10
  csp: "default-src 'self'; script-src 'self'; style-src 'self'"
  csrf_cookie_name: _csrf
csp_hashes:
  always: []
  dev_only: []
auth:
  jwt:
    secret: "12345678901234567890123456789012"
    issuer: starter
    token_duration: 86400
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
    - items:
        - label: Login
          href: /login
          visibility: guest
        - type: theme_toggle
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	if err := Init(dir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if got := App.Nav.Brand.Label; got != "Starter" {
		t.Fatalf("Init() brand label = %q, want %q", got, "Starter")
	}
	if len(App.Nav.Sections) != 2 {
		t.Fatalf("Init() sections = %d, want 2", len(App.Nav.Sections))
	}
	if got := App.Nav.Sections[1].Items[0].Visibility; got != "guest" {
		t.Fatalf("Init() visibility = %q, want %q", got, "guest")
	}
	if got := App.Nav.Sections[1].Items[1].Type; got != "theme_toggle" {
		t.Fatalf("Init() special type = %q, want %q", got, "theme_toggle")
	}
}
