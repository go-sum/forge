package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestParseWorkFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantErr bool
	}{
		{
			name: "block syntax",
			content: `go 1.26.0

use (
	.
	./pkg/foo
	./pkg/bar
)
`,
			want: []string{".", "./pkg/foo", "./pkg/bar"},
		},
		{
			name:    "single-line use",
			content: "go 1.26.0\n\nuse ./pkg/solo\n",
			want:    []string{"./pkg/solo"},
		},
		{
			name: "mixed block and single-line",
			content: `go 1.26.0

use (
	./pkg/a
	./pkg/b
)

use ./pkg/c
`,
			want: []string{"./pkg/a", "./pkg/b", "./pkg/c"},
		},
		{
			name:    "empty use block",
			content: "go 1.26.0\n\nuse (\n)\n",
			want:    nil,
		},
		{
			name: "comments and blank lines ignored",
			content: `go 1.26.0

// workspace modules
use (
	// core
	.

	./pkg/foo
)
`,
			want: []string{".", "./pkg/foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "go.work")
			if err := os.WriteFile(tmp, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := parseWorkFile(tmp)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d modules %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("module[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseWorkFile_NotFound(t *testing.T) {
	_, err := parseWorkFile("/nonexistent/go.work")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseWorkFile_RealFile(t *testing.T) {
	// Parse the actual go.work in the repo root.
	root, err := gitRepoRoot()
	if err != nil {
		t.Skip("not in a git repo")
	}
	modules, err := parseWorkFile(filepath.Join(root, "go.work"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modules) < 2 {
		t.Fatalf("expected at least 2 modules, got %d: %v", len(modules), modules)
	}
	// First module should be the root.
	if modules[0] != "." {
		t.Errorf("first module = %q, want %q", modules[0], ".")
	}
}

func TestPrefixWriter(t *testing.T) {
	var mu sync.Mutex
	var buf bytes.Buffer

	pw := newPrefixWriter(&mu, &buf, "./pkg/auth", 16)

	// Write a complete line.
	pw.Write([]byte("ok  auth/service 0.01s\n"))
	got := buf.String()
	want := "[./pkg/auth]     ok  auth/service 0.01s\n"
	if got != want {
		t.Fatalf("complete line:\ngot:  %q\nwant: %q", got, want)
	}

	// Write a partial line, then the rest.
	buf.Reset()
	pw.Write([]byte("ok  auth"))
	if buf.String() != "" {
		t.Fatalf("partial line should not output yet, got: %q", buf.String())
	}
	pw.Write([]byte("/model 0.02s\n"))
	got = buf.String()
	want = "[./pkg/auth]     ok  auth/model 0.02s\n"
	if got != want {
		t.Fatalf("joined partial line:\ngot:  %q\nwant: %q", got, want)
	}

	// Flush a trailing partial line.
	buf.Reset()
	pw.Write([]byte("no newline"))
	pw.Flush()
	got = buf.String()
	want = "[./pkg/auth]     no newline\n"
	if got != want {
		t.Fatalf("flushed partial:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestPrefixWriterMultipleLines(t *testing.T) {
	var mu sync.Mutex
	var buf bytes.Buffer

	pw := newPrefixWriter(&mu, &buf, ".", 6)

	// Write multiple lines in one call.
	pw.Write([]byte("line1\nline2\nline3\n"))
	got := buf.String()
	want := "[.]    line1\n[.]    line2\n[.]    line3\n"
	if got != want {
		t.Fatalf("multi-line:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFilterModules(t *testing.T) {
	all := []string{".", "./pkg/auth", "./pkg/kv", "./pkg/componentry", "./pkg/server"}

	tests := []struct {
		name     string
		includes []string
		excludes []string
		want     []string
	}{
		{
			name: "no filters",
			want: all,
		},
		{
			name:     "include auth",
			includes: []string{"auth"},
			want:     []string{"./pkg/auth"},
		},
		{
			name:     "exclude auth",
			excludes: []string{"auth"},
			want:     []string{".", "./pkg/kv", "./pkg/componentry", "./pkg/server"},
		},
		{
			name:     "include pkg then exclude auth",
			includes: []string{"pkg"},
			excludes: []string{"auth"},
			want:     []string{"./pkg/kv", "./pkg/componentry", "./pkg/server"},
		},
		{
			name:     "include matches nothing",
			includes: []string{"nonexistent"},
			want:     nil,
		},
		{
			name:     "multiple includes",
			includes: []string{"auth", "kv"},
			want:     []string{"./pkg/auth", "./pkg/kv"},
		},
		{
			name:     "multiple excludes",
			excludes: []string{"auth", "kv"},
			want:     []string{".", "./pkg/componentry", "./pkg/server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterModules(all, tt.includes, tt.excludes)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d modules %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("module[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
