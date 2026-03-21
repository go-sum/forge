package main

import (
	"log/slog"
	"os"

	"starter/internal/app"
)

func main() {
	app := app.New()
	defer app.Shutdown()

	if err := app.Run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
