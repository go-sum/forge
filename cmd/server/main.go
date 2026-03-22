package main

import (
	"log/slog"
	"os"

	"github.com/y-goweb/foundry/internal/app"
)

func main() {
	app := app.New()
	defer app.Shutdown()

	if err := app.Run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
