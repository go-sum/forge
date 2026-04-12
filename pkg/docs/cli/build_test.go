package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildInvokesHugoAndRebuildsOutput(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "public", "doc")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	stalePath := filepath.Join(outputDir, "stale.txt")
	if err := os.WriteFile(stalePath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	capturePath := filepath.Join(tmpDir, "hugo.args")
	fakeHugoDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(fakeHugoDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	fakeHugoPath := filepath.Join(fakeHugoDir, "hugo")
	script := "#!/bin/sh\n" +
		"set -eu\n" +
		"src=''\n" +
		"dest=''\n" +
		"while [ \"$#\" -gt 0 ]; do\n" +
		"  case \"$1\" in\n" +
		"    --source) src=\"$2\"; shift 2 ;;\n" +
		"    --destination) dest=\"$2\"; shift 2 ;;\n" +
		"    *) shift ;;\n" +
		"  esac\n" +
		"done\n" +
		"printf '%s\\n%s\\n' \"$src\" \"$dest\" > \"$TEST_CAPTURE\"\n" +
		"mkdir -p \"$dest/guide\"\n" +
		"printf '<h1>Docs</h1>' > \"$dest/index.html\"\n" +
		"printf '<h1>Guide</h1>' > \"$dest/guide/index.html\"\n" +
		"printf '<h1>Missing</h1>' > \"$dest/404.html\"\n"
	if err := os.WriteFile(fakeHugoPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake hugo: %v", err)
	}

	t.Setenv("PATH", fakeHugoDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TEST_CAPTURE", capturePath)

	if err := build(".docs", outputDir); err != nil {
		t.Fatalf("build() error = %v", err)
	}

	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("stale docs output should have been removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "index.html")); err != nil {
		t.Fatalf("generated index missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "guide", "index.html")); err != nil {
		t.Fatalf("generated guide missing: %v", err)
	}

	argsRaw, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	args := strings.Split(strings.TrimSpace(string(argsRaw)), "\n")
	if len(args) != 2 {
		t.Fatalf("captured args = %q, want source and destination", string(argsRaw))
	}
	if got := args[0]; got != ".docs" {
		t.Fatalf("source = %q, want %q", got, ".docs")
	}
	if got := filepath.Clean(args[1]); got != filepath.Clean(outputDir) {
		t.Fatalf("destination = %q, want %q", got, outputDir)
	}
}

func TestBuildForwardsCustomSource(t *testing.T) {
	tmpDir := t.TempDir()
	capturePath := filepath.Join(tmpDir, "hugo.args")
	fakeHugoDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(fakeHugoDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	fakeHugoPath := filepath.Join(fakeHugoDir, "hugo")
	script := "#!/bin/sh\n" +
		"set -eu\n" +
		"src=''\n" +
		"while [ \"$#\" -gt 0 ]; do\n" +
		"  case \"$1\" in\n" +
		"    --source) src=\"$2\"; shift 2 ;;\n" +
		"    *) shift ;;\n" +
		"  esac\n" +
		"done\n" +
		"printf '%s\\n' \"$src\" > \"$TEST_CAPTURE\"\n"
	if err := os.WriteFile(fakeHugoPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake hugo: %v", err)
	}

	t.Setenv("PATH", fakeHugoDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TEST_CAPTURE", capturePath)

	outputDir := filepath.Join(tmpDir, "public", "doc")
	if err := build("my-docs", outputDir); err != nil {
		t.Fatalf("build() error = %v", err)
	}

	gotRaw, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	if got := strings.TrimSpace(string(gotRaw)); got != "my-docs" {
		t.Fatalf("source = %q, want %q", got, "my-docs")
	}
}

func TestBuildReturnsErrorWhenHugoFails(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHugoDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(fakeHugoDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	fakeHugoPath := filepath.Join(fakeHugoDir, "hugo")
	if err := os.WriteFile(fakeHugoPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write fake hugo: %v", err)
	}

	t.Setenv("PATH", fakeHugoDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := build(".docs", filepath.Join(tmpDir, "public", "doc")); err == nil {
		t.Fatal("build() error = nil, want non-nil when hugo exits 1")
	}
}

func TestBuildResolvesRelativeDestination(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHugoDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(fakeHugoDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}

	capturePath := filepath.Join(tmpDir, "hugo.args")
	fakeHugoPath := filepath.Join(fakeHugoDir, "hugo")
	script := "#!/bin/sh\n" +
		"set -eu\n" +
		"dest=''\n" +
		"while [ \"$#\" -gt 0 ]; do\n" +
		"  case \"$1\" in\n" +
		"    --destination) dest=\"$2\"; shift 2 ;;\n" +
		"    *) shift ;;\n" +
		"  esac\n" +
		"done\n" +
		"printf '%s\\n' \"$dest\" > \"$TEST_CAPTURE\"\n"
	if err := os.WriteFile(fakeHugoPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake hugo: %v", err)
	}

	t.Setenv("PATH", fakeHugoDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TEST_CAPTURE", capturePath)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	if err := build(".docs", "public/doc"); err != nil {
		t.Fatalf("build() error = %v", err)
	}

	gotRaw, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	got := filepath.Clean(strings.TrimSpace(string(gotRaw)))
	want := filepath.Join(cwd, "public", "doc")
	if got != want {
		t.Fatalf("destination = %q, want %q", got, want)
	}
}
