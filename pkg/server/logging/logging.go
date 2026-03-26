package logging

import (
	"io"
	"log/slog"
	"os"
)

// Config controls the default logger that New and Init construct.
type Config struct {
	Development bool
	Level       string
	TextOutput  io.Writer
	JSONOutput  io.Writer
}

// New constructs a slog.Logger using text output in development and JSON output
// otherwise. Nil outputs default to stderr for text and stdout for JSON.
func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{Level: level}

	if cfg.Development {
		out := cfg.TextOutput
		if out == nil {
			out = os.Stderr
		}
		return slog.New(slog.NewTextHandler(out, opts))
	}

	out := cfg.JSONOutput
	if out == nil {
		out = os.Stdout
	}
	return slog.New(slog.NewJSONHandler(out, opts))
}

// Init constructs a logger from cfg, installs it as the slog default, and
// returns it for optional direct use by the caller.
func Init(cfg Config) *slog.Logger {
	logger := New(cfg)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
