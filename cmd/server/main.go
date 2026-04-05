package main

import (
	"log/slog"
	"os"

	"github.com/go-sum/forge/internal/app"
)

// version is set at build time via -ldflags="-X main.version=...".
var version string

func main() {
	a := app.New(version)
	defer a.Shutdown()

	if err := a.Start(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
