package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_DevelopmentUsesTextHandler(t *testing.T) {
	var out bytes.Buffer

	logger := New(Config{
		Development: true,
		Level:       "debug",
		TextOutput:  &out,
	})
	logger.Debug("debug message")

	got := out.String()
	if !strings.Contains(got, "level=DEBUG") {
		t.Fatalf("expected text handler output, got %q", got)
	}
	if !strings.Contains(got, "msg=\"debug message\"") {
		t.Fatalf("expected debug message in output, got %q", got)
	}
}

func TestNew_ProductionUsesJSONHandler(t *testing.T) {
	var out bytes.Buffer

	logger := New(Config{
		Level:      "info",
		JSONOutput: &out,
	})
	logger.Info("structured")

	got := out.String()
	if !strings.Contains(got, `"level":"INFO"`) {
		t.Fatalf("expected JSON handler output, got %q", got)
	}
	if !strings.Contains(got, `"msg":"structured"`) {
		t.Fatalf("expected structured message in output, got %q", got)
	}
}

func TestInit_InstallsDefaultLogger(t *testing.T) {
	original := slog.Default()
	t.Cleanup(func() { slog.SetDefault(original) })

	var out bytes.Buffer
	logger := Init(Config{
		Development: true,
		TextOutput:  &out,
	})

	if slog.Default() != logger {
		t.Fatalf("default logger not updated")
	}
}
