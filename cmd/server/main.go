package main

import (
	"log/slog"
	"os"

	"starter/internal/handler"
	"starter/internal/services"
	pkgserver "starter/pkg/server"
)

func main() {
	c := services.NewContainer()
	defer c.Shutdown()

	h := handler.New(c.Services, c.Sessions, c.Validator, c.DB, c.ServerConfig.CSRFCookieName)
	h.RegisterRoutes(c.Web, c.Sessions, c.ServerConfig.PublicPrefix, c.PublicDir)

	if err := pkgserver.Start(c.Web, c.ServerConfig); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
