package assetconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNormalizesPathsAndURLs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".assets.yaml")
	content := `
paths:
  source_dir: web/static
  public_dir: web/public
  public_prefix: assets/
js:
  downloads:
    - name: htmx
      version: "2.0.0"
      url: https://example.com/{version}/htmx.js
      target: js/htmx.js
  bundles:
    - entry: js/app.js
      target: js/app.js
css:
  - tool: tailwind
    input: css/app.css
    output: css/app.css
sprites:
  icons:
    enabled: true
    target: img/icons.svg
    sources:
      - path: svg/icons
        files: [one.svg]
      - path: https://example.com/icons
        files: [two.svg]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Paths.SourceRoot() != filepath.Clean("web/static") {
		t.Fatalf("SourceRoot() = %q", cfg.Paths.SourceRoot())
	}
	if cfg.Paths.PublicRoot() != filepath.Clean("web/public") {
		t.Fatalf("PublicRoot() = %q", cfg.Paths.PublicRoot())
	}
	if cfg.Paths.URLPrefix() != "/assets" {
		t.Fatalf("URLPrefix() = %q", cfg.Paths.URLPrefix())
	}
	if cfg.Paths.PublicURL("css/app.css") != "/assets/css/app.css" {
		t.Fatalf("PublicURL() = %q", cfg.Paths.PublicURL("css/app.css"))
	}
	if cfg.JS.Downloads[0].Target != filepath.Clean("web/public/js/htmx.js") {
		t.Fatalf("download target = %q", cfg.JS.Downloads[0].Target)
	}
	if cfg.JS.Bundles[0].Entry != filepath.Clean("web/static/js/app.js") {
		t.Fatalf("bundle entry = %q", cfg.JS.Bundles[0].Entry)
	}
	if cfg.CSS[0].Output != filepath.Clean("web/public/css/app.css") {
		t.Fatalf("css output = %q", cfg.CSS[0].Output)
	}
	if cfg.Sprites["icons"].Sources[0].Path != filepath.Clean("web/static/svg/icons") {
		t.Fatalf("sprite source = %q", cfg.Sprites["icons"].Sources[0].Path)
	}
	if cfg.Sprites["icons"].Sources[1].Path != "https://example.com/icons" {
		t.Fatalf("remote sprite source = %q", cfg.Sprites["icons"].Sources[1].Path)
	}
}

func TestPathsDefaults(t *testing.T) {
	var paths Paths
	if paths.SourceRoot() != defaultSourceDir {
		t.Fatalf("SourceRoot() = %q", paths.SourceRoot())
	}
	if paths.PublicRoot() != defaultPublicDir {
		t.Fatalf("PublicRoot() = %q", paths.PublicRoot())
	}
	if paths.URLPrefix() != defaultPublicPrefix {
		t.Fatalf("URLPrefix() = %q", paths.URLPrefix())
	}
}
