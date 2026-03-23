package config

import (
	"os"
	"path/filepath"
	"strings"
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
    - align: end
      items:
        - label: Login
          href: /login
          visibility: guest
        - slot: theme_toggle
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	if err := InitConfig(dir); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

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

func TestInitRejectsInvalidNavContentFile(t *testing.T) {
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
				"nav.yaml": tt.nav,
			}
			for name, contents := range files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error = %v", name, err)
				}
			}

			if err := InitConfig(dir); err == nil {
				t.Fatal("InitConfig() error = nil, want validation error")
			} else if !strings.Contains(err.Error(), "nav.yaml") {
				t.Fatalf("InitConfig() error = %v, want nav.yaml attribution", err)
			}
		})
	}
}
