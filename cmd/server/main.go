package main

import (
	"context"
	"log/slog"
	"os"

	"starter/internal/handler"
	"starter/internal/server"
	"starter/internal/services"
	"starter/pkg/database"
	pkgserver "starter/pkg/server"
)

func main() {
	c := services.NewContainer()
	defer c.Shutdown()

	h := handler.New(
		c.Services,
		c.Sessions,
		c.Validator,
		func(ctx context.Context) error { return database.CheckHealth(ctx, c.DB) },
		c.ServerConfig.CSRFCookieName,
		c.Config.Nav,
	)
	server.RegisterRoutes(c.Web, h, c.Sessions, c.Services.User, c.ServerConfig.PublicPrefix, c.PublicDir)

	if err := pkgserver.Start(c.Web, c.ServerConfig); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
