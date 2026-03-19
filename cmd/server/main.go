package main

import (
	"log/slog"
	"os"

	"starter/internal/handlers"
	internalserver "starter/internal/server"
	"starter/internal/services"
	pkgserver "starter/pkg/server"
)

func main() {
	c := services.NewContainer()
	defer c.Shutdown()

	h := handlers.New(c.DB, c.Config)
	internalserver.RegisterRoutes(c.Web, h, c.ServerConfig.PublicPrefix, c.PublicDir)

	if err := pkgserver.Start(c.Web, c.ServerConfig); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
