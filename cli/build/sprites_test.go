package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-sum/componentry/assetconfig"
)

func TestProcessSVG(t *testing.T) {
	tests := []struct {
		name string
		data string
		id   string
		want string
	}{
		{
			name: "extracts viewBox and wraps inner content in symbol",
			data: `<svg viewBox="0 0 16 16"><path d="M0 0h16v16H0z"/></svg>`,
			id:   "icon-box",
			want: `    <symbol id="icon-box" viewBox="0 0 16 16">` + "\n" +
				`      <path d="M0 0h16v16H0z"/>` + "\n" +
				`    </symbol>`,
		},
		{
			name: "defaults to 0 0 24 24 when viewBox is missing",
			data: `<svg><circle cx="12" cy="12" r="10"/></svg>`,
			id:   "no-viewbox",
			want: `    <symbol id="no-viewbox" viewBox="0 0 24 24">` + "\n" +
				`      <circle cx="12" cy="12" r="10"/>` + "\n" +
				`    </symbol>`,
		},
		{
			name: "strips script tags",
			data: `<svg viewBox="0 0 24 24"><script>alert('xss')</script><path d="M1 1"/></svg>`,
			id:   "no-script",
			want: `    <symbol id="no-script" viewBox="0 0 24 24">` + "\n" +
				`      <path d="M1 1"/>` + "\n" +
				`    </symbol>`,
		},
		{
			name: "strips event handler attributes",
			data: `<svg viewBox="0 0 24 24"><path d="M1 1" onclick="alert('xss')"/></svg>`,
			id:   "no-events",
			want: `    <symbol id="no-events" viewBox="0 0 24 24">` + "\n" +
				`      <path d="M1 1"/>` + "\n" +
				`    </symbol>`,
		},
		{
			name: "transfers presentation attributes from outer svg to symbol",
			data: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 1"/></svg>`,
			id:   "pres-attrs",
			want: `    <symbol id="pres-attrs" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">` + "\n" +
				`      <path d="M1 1"/>` + "\n" +
				`    </symbol>`,
		},
		{
			name: "produces self-closing symbol when SVG has no inner content",
			data: `<svg viewBox="0 0 10 10"></svg>`,
			id:   "empty",
			want: `    <symbol id="empty" viewBox="0 0 10 10"/>`,
		},
		{
			name: "strips multiline script tags",
			data: "<svg viewBox=\"0 0 24 24\"><script type=\"text/javascript\">\nvar x = 1;\n</script><rect width=\"10\" height=\"10\"/></svg>",
			id:   "multiline-script",
			want: `    <symbol id="multiline-script" viewBox="0 0 24 24">` + "\n" +
				`      <rect width="10" height="10"/>` + "\n" +
				`    </symbol>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processSVG([]byte(tt.data), tt.id)
			if err != nil {
				t.Fatalf("processSVG() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("processSVG() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestAllRemoteSources(t *testing.T) {
	tests := []struct {
		name    string
		sources []assetconfig.SourcesConfig
		want    bool
	}{
		{
			name: "all https sources returns true",
			sources: []assetconfig.SourcesConfig{
				{Path: "https://cdn.example.com/icons/"},
				{Path: "https://other.example.com/svg/"},
			},
			want: true,
		},
		{
			name: "mixed local and remote returns false",
			sources: []assetconfig.SourcesConfig{
				{Path: "https://cdn.example.com/icons/"},
				{Path: "./local/icons"},
			},
			want: false,
		},
		{
			name: "all local sources returns false",
			sources: []assetconfig.SourcesConfig{
				{Path: "./icons"},
				{Path: "/absolute/icons"},
			},
			want: false,
		},
		{
			name: "empty slice returns false",
			sources: []assetconfig.SourcesConfig{},
			want:    false,
		},
		{
			name:    "nil slice returns false",
			sources: nil,
			want:    false,
		},
		{
			name: "http source returns true",
			sources: []assetconfig.SourcesConfig{
				{Path: "http://cdn.example.com/icons/"},
			},
			want: true,
		},
		{
			name: "file URI is not remote",
			sources: []assetconfig.SourcesConfig{
				{Path: "file:///local/icons"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := allRemoteSources(tt.sources)
			if got != tt.want {
				t.Fatalf("allRemoteSources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchSVGLocalFile(t *testing.T) {
	dir := t.TempDir()
	content := `<svg viewBox="0 0 24 24"><circle r="10"/></svg>`
	writeTempFile(t, dir, "test.svg", content)

	data, err := fetchSVG(dir, "test.svg")
	if err != nil {
		t.Fatalf("fetchSVG() error = %v", err)
	}
	if string(data) != content {
		t.Fatalf("fetchSVG() = %q, want %q", string(data), content)
	}
}

func TestFetchSVGFileURI(t *testing.T) {
	dir := t.TempDir()
	content := `<svg viewBox="0 0 16 16"><rect/></svg>`
	writeTempFile(t, dir, "icon.svg", content)

	data, err := fetchSVG("file://"+dir, "icon.svg")
	if err != nil {
		t.Fatalf("fetchSVG() error = %v", err)
	}
	if string(data) != content {
		t.Fatalf("fetchSVG() = %q, want %q", string(data), content)
	}
}

func TestFetchSVGLocalFileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := fetchSVG(dir, "nonexistent.svg")
	if err == nil {
		t.Fatal("fetchSVG() error = nil, want error for missing file")
	}
}

func writeTempFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
