package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"starter/config"
	"starter/internal/repository"
	internalserver "starter/internal/server"
	"starter/internal/service"
	"starter/pkg/assetconfig"
	"starter/pkg/assets"
	"starter/pkg/auth"
	componenticons "starter/pkg/components/icons"
	componentinstall "starter/pkg/components/install"
	"starter/pkg/components/interactive"
	"starter/pkg/database"
	pkgserver "starter/pkg/server"
	"starter/pkg/validate"
)

const assetConfigPath = assetconfig.DefaultConfigPath

var componentIconOverrides = map[componenticons.Key]componenticons.Ref{}

func (c *Container) initConfig() {
	if err := config.Init("config"); err != nil {
		panic(fmt.Sprintf("config: %v", err))
	}
	c.Config = config.App
}

// initLogger configures the global slog logger. Handler type is environment-driven
// (text/stderr in development, JSON/stdout in production); level comes from
// config.Log.Level ("debug", "info", "warn", "error").
func (c *Container) initLogger() {
	dev := c.Config.IsDevelopment()
	level := parseLogLevel(c.Config.Log.Level)
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if dev {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(h))
	slog.Info("logger initialized", "level", level, "env", c.Config.App.Env)
}

// parseLogLevel maps a config log level string to slog.Level.
// The LogConfig.Level validate tag ("oneof=debug info warn error") guarantees
// the value is one of the four known strings; the default branch is unreachable
// in normal operation and falls back to Info.
func parseLogLevel(s string) slog.Level {
	switch s {
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

func (c *Container) initAssets() {
	assetCfg, err := assetconfig.Load(assetConfigPath)
	if err != nil {
		panic(fmt.Sprintf("asset config: %v", err))
	}

	publicDir := assetCfg.Paths.PublicRoot()
	publicPrefix := assetCfg.Paths.URLPrefix()
	if err := assets.Init(publicDir, publicPrefix); err != nil {
		panic(fmt.Sprintf("assets: %v", err))
	}

	componentinstall.ApplyDefault(componentinstall.Config{
		PathFunc:      assets.Path,
		IconOverrides: componentIconOverrides,
	})

	// assets.Default is set by Init; store a reference for test injection.
	c.Assets = assets.Default
	c.AssetPaths = assetCfg.Paths
	c.PublicDir = publicDir
}

func (c *Container) initDatabase() {
	pool, err := database.Connect(context.Background(), c.Config.DSN())
	if err != nil {
		panic(fmt.Sprintf("database: %v", err))
	}
	slog.Info("database connected")
	c.DB = pool
}

// injectCSPHashes appends all hash tokens to the script-src directive of csp
// in a single replacement, preserving their order. Called with the assembled
// token list from initWeb.
func injectCSPHashes(csp string, hashes []string) string {
	if len(hashes) == 0 {
		return csp
	}
	return strings.Replace(csp, "script-src", "script-src "+strings.Join(hashes, " "), 1)
}

// initWeb builds ServerConfig from the app config and creates the Echo instance
// with the full middleware stack applied.
// Port int -> string and GracefulTimeout int (seconds) -> time.Duration conversions
// happen here (see config.ServerConfig field comments).
func (c *Container) initWeb() {
	cfg := c.Config
	hashes := []string{interactive.ScriptCSPHash}
	hashes = append(hashes, cfg.CSPHashes.Always...)
	if cfg.IsDevelopment() {
		hashes = append(hashes, cfg.CSPHashes.DevOnly...)
	}
	c.ServerConfig = pkgserver.Config{
		Host:            cfg.Server.Host,
		Port:            strconv.Itoa(cfg.Server.Port),
		Debug:           cfg.IsDevelopment(),
		GracefulTimeout: time.Duration(cfg.Server.GracefulTimeout) * time.Second,
		CookieSecure:    cfg.Auth.Session.Secure,
		CSP:             injectCSPHashes(cfg.Server.CSP, hashes),
		CSRFCookieName:  cfg.Server.CSRFCookieName,
		PublicPrefix:    c.AssetPaths.URLPrefix(),
	}
	c.Web = pkgserver.New(c.ServerConfig)
	internalserver.Setup(c.Web, c.ServerConfig, cfg.Nav)
}

func (c *Container) initAuth() {
	c.Sessions = auth.NewSessionStore(auth.SessionConfig{
		Name:       c.Config.Auth.Session.Name,
		AuthKey:    c.Config.Auth.Session.AuthKey,
		EncryptKey: c.Config.Auth.Session.EncryptKey,
		MaxAge:     c.Config.Auth.Session.MaxAge,
		Secure:     c.Config.Auth.Session.Secure,
	})
}

func (c *Container) initValidator() {
	c.Validator = validate.New()
}

func (c *Container) initRepos() {
	c.Repos = repository.NewRepositories(c.DB)
}

func (c *Container) initServices() {
	c.Services = service.NewServices(c.Repos, c.DB)
}
