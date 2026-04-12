package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sum/assets/config"
)

func TestBuildJSBundlesSingleAppEntrypoint(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "static", "js")
	publicDir := filepath.Join(tmpDir, "public", "js")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		t.Fatalf("mkdir public: %v", err)
	}

	appSource := "import { init } from './components.js'\ninit()\n"
	componentSource := "" +
		"export function init() { console.log('live-marker') }\n" +
		"export function unused() { console.log('unused-marker') }\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "app.js"), []byte(appSource), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "components.js"), []byte(componentSource), 0o644); err != nil {
		t.Fatalf("write components.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(publicDir, "components.js"), []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale components.js: %v", err)
	}

	cfg := config.JSConfig{
		Bundles: []config.JSBundle{{
			Entry:  filepath.Join(sourceDir, "app.js"),
			Target: filepath.Join(publicDir, "app.js"),
		}},
	}

	if err := buildJS(cfg, false); err != nil {
		t.Fatalf("buildJS() error = %v", err)
	}

	out, err := os.ReadFile(filepath.Join(publicDir, "app.js"))
	if err != nil {
		t.Fatalf("read bundled app.js: %v", err)
	}
	got := string(out)

	if strings.Contains(got, "import ") {
		t.Fatalf("bundle still contains import statement: %s", got)
	}
	if !strings.Contains(got, "live-marker") {
		t.Fatalf("bundle missing live marker: %s", got)
	}
	if strings.Contains(got, "unused-marker") {
		t.Fatalf("bundle retained unused export: %s", got)
	}
	if _, err := os.Stat(filepath.Join(publicDir, "components.js")); !os.IsNotExist(err) {
		t.Fatalf("stale public/js/components.js should have been removed, err=%v", err)
	}
}

func TestBuildJSBundlesWithMinification(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "static", "js")
	publicDir := filepath.Join(tmpDir, "public", "js")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		t.Fatalf("mkdir public: %v", err)
	}

	// Source with whitespace and long identifiers that esbuild will minify.
	appSource := "export function greetUser(userName) {\n  console.log('hello ' + userName)\n}\ngreetUser('world')\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "app.js"), []byte(appSource), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	cfg := config.JSConfig{
		Bundles: []config.JSBundle{{
			Entry:  filepath.Join(sourceDir, "app.js"),
			Target: filepath.Join(publicDir, "app.min.js"),
		}},
	}

	if err := buildJS(cfg, true); err != nil {
		t.Fatalf("buildJS(minify=true) error = %v", err)
	}

	minified, err := os.ReadFile(filepath.Join(publicDir, "app.min.js"))
	if err != nil {
		t.Fatalf("read app.min.js: %v", err)
	}

	// Build without minification for size comparison.
	cfg.Bundles[0].Target = filepath.Join(publicDir, "app.js")
	if err := buildJS(cfg, false); err != nil {
		t.Fatalf("buildJS(minify=false) error = %v", err)
	}
	unminified, err := os.ReadFile(filepath.Join(publicDir, "app.js"))
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}

	if len(minified) >= len(unminified) {
		t.Fatalf("minified output (%d bytes) should be smaller than unminified (%d bytes)", len(minified), len(unminified))
	}
	if !strings.Contains(string(minified), "hello") {
		t.Fatalf("minified output missing string literal 'hello': %s", minified)
	}
}
