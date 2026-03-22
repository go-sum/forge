package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetsNewAndPath(t *testing.T) {
	dir := t.TempDir()
	cssDir := filepath.Join(dir, "css")
	if err := os.MkdirAll(cssDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(cssDir, "app.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	a, err := New(dir, "/public")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	path := a.Path("css/app.css")
	if !strings.HasPrefix(path, "/public/css/app.css?v=") {
		t.Fatalf("Path() = %q", path)
	}
	if got := a.Path("js/app.js"); got != "/public/js/app.js" {
		t.Fatalf("fallback Path() = %q", got)
	}
}

func TestAssetsMissingDirAndPackageHelpers(t *testing.T) {
	a, err := New(filepath.Join(t.TempDir(), "missing"), "/public")
	if err != nil {
		t.Fatalf("New(missing) error = %v", err)
	}
	if got := a.Path("css/app.css"); got != "/public/css/app.css" {
		t.Fatalf("Path() = %q", got)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('x')"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := Init(dir, "/assets"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if got := Path("app.js"); !strings.HasPrefix(got, "/assets/app.js?v=") {
		t.Fatalf("package Path() = %q", got)
	}
	if Must(a, nil) != a {
		t.Fatal("Must() did not return assets instance")
	}
	MustInit(dir, "/assets")
}
