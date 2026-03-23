package main

import (
	"log/slog"
	"os"

	"github.com/go-sum/forge/internal/app"
)

func main() {
	app := app.New()
	defer app.Shutdown()

	if err := app.Start(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
