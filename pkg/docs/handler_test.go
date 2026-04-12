package docs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestResolvePath(t *testing.T) {
	root := filepath.Join("public", "doc")

	tests := []struct {
		name      string
		request   string
		wantPath  string
		wantAsset bool
		wantErr   bool
	}{
		{name: "home", request: "/docs", wantPath: filepath.Join(root, "index.html")},
		{name: "nested page", request: "/docs/architecture/design_guide", wantPath: filepath.Join(root, "architecture", "design_guide", "index.html")},
		{name: "asset", request: "/docs/css/main.css", wantPath: filepath.Join(root, "css", "main.css"), wantAsset: true},
		{name: "traversal", request: "/docs/../../etc/passwd", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotAsset, err := resolvePath(root, tc.request)
			if tc.wantErr {
				if err == nil {
					t.Fatal("resolvePath() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolvePath() error = %v", err)
			}
			if gotPath != tc.wantPath {
				t.Fatalf("path = %q, want %q", gotPath, tc.wantPath)
			}
			if gotAsset != tc.wantAsset {
				t.Fatalf("asset = %v, want %v", gotAsset, tc.wantAsset)
			}
		})
	}
}

func TestHandlerServesPagesAssetsAndDocs404(t *testing.T) {
	tmpDir := t.TempDir()
	docsRoot := filepath.Join(tmpDir, "doc")
	if err := os.MkdirAll(filepath.Join(docsRoot, "architecture", "api-rules"), 0o755); err != nil {
		t.Fatalf("mkdir docs page: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(docsRoot, "css"), 0o755); err != nil {
		t.Fatalf("mkdir docs asset: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(docsRoot, "js"), 0o755); err != nil {
		t.Fatalf("mkdir docs js asset: %v", err)
	}

	files := map[string]string{
		filepath.Join(docsRoot, "index.html"):                              "<h1>Docs</h1>",
		filepath.Join(docsRoot, "architecture", "api-rules", "index.html"): "<h1>API Rules</h1>",
		filepath.Join(docsRoot, "css", "main.css"):                         "body{color:#000;}",
		filepath.Join(docsRoot, "js", "darkmode.js"):                       "console.log('theme');",
		filepath.Join(docsRoot, "404.html"):                                "<h1>Document not found</h1>",
	}
	for name, contents := range files {
		if err := os.WriteFile(name, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	e := echo.New()
	h := NewHandler(tmpDir)
	mustAdd(t, e, http.MethodGet, "/docs", "docs.index", h.Handle)
	mustAdd(t, e, http.MethodGet, "/docs/*", "docs.show", h.Handle)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
		wantType   string
	}{
		{path: "/docs", wantStatus: http.StatusOK, wantBody: "<h1>Docs</h1>", wantType: "text/html; charset=utf-8"},
		{path: "/docs/architecture/api-rules", wantStatus: http.StatusOK, wantBody: "<h1>API Rules</h1>", wantType: "text/html; charset=utf-8"},
		{path: "/docs/css/main.css", wantStatus: http.StatusOK, wantBody: "body{color:#000;}", wantType: "text/css; charset=utf-8"},
		{path: "/docs/js/darkmode.js", wantStatus: http.StatusOK, wantBody: "console.log('theme');", wantType: "text/javascript; charset=utf-8"},
		{path: "/docs/missing", wantStatus: http.StatusNotFound, wantBody: "<h1>Document not found</h1>", wantType: "text/html; charset=utf-8"},
		{path: "/docs/css/missing.css", wantStatus: http.StatusNotFound, wantBody: `{"message":"Not Found"}`, wantType: "application/json"},
		{path: "/docs/../../etc/passwd", wantStatus: http.StatusNotFound, wantBody: `{"message":"Not Found"}`, wantType: "application/json"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tc.wantBody {
				t.Fatalf("body = %q, want %q", got, tc.wantBody)
			}
			if tc.wantType != "" {
				if got := rec.Header().Get(echo.HeaderContentType); got != tc.wantType {
					t.Fatalf("content-type = %q, want %q", got, tc.wantType)
				}
			}
		})
	}
}

func TestHandlerCacheControlHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	docsRoot := filepath.Join(tmpDir, "doc")
	if err := os.MkdirAll(filepath.Join(docsRoot, "css"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	files := map[string]string{
		filepath.Join(docsRoot, "index.html"):      "<h1>Docs</h1>",
		filepath.Join(docsRoot, "css", "main.css"): "body{}",
	}
	for name, contents := range files {
		if err := os.WriteFile(name, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	e := echo.New()
	h := NewHandler(tmpDir)
	mustAdd(t, e, http.MethodGet, "/docs", "docs.index", h.Handle)
	mustAdd(t, e, http.MethodGet, "/docs/*", "docs.show", h.Handle)

	tests := []struct {
		path          string
		wantCacheCtrl string
	}{
		{path: "/docs", wantCacheCtrl: "no-cache"},
		{path: "/docs/css/main.css", wantCacheCtrl: "public, max-age=3600"},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if got := rec.Header().Get("Cache-Control"); got != tc.wantCacheCtrl {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.wantCacheCtrl)
			}
		})
	}
}

func mustAdd(t *testing.T, e *echo.Echo, method, path, name string, h echo.HandlerFunc) {
	t.Helper()
	if _, err := e.AddRoute(echo.Route{Method: method, Path: path, Name: name, Handler: h}); err != nil {
		t.Fatalf("AddRoute %s %s: %v", method, path, err)
	}
}
