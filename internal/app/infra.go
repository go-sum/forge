package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	authsvc "github.com/go-sum/auth/service"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/componentry/assetconfig"
	"github.com/go-sum/componentry/assets"
	icons "github.com/go-sum/componentry/icons"
	install "github.com/go-sum/componentry/install"
	"github.com/go-sum/componentry/interactive"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/adapters"
	"github.com/go-sum/forge/internal/health"
	"github.com/go-sum/forge/internal/repository"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/server"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/validate"
)

const assetConfigPath = assetconfig.DefaultConfigPath

var componentIconOverrides = map[icons.Key]icons.Ref{}

func (c *Container) initConfig() {
	if err := config.InitConfig("config"); err != nil {
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

	install.ApplyDefault(install.Config{
		PathFunc:      assets.Path,
		IconOverrides: componentIconOverrides,
	})

	c.Assets = assets.Default
	c.AssetPaths = assetCfg.Paths
	c.PublicDir = publicDir
}

func (c *Container) initDatabase() {
	pool, err := database.Connect(context.Background(), c.Config.DSN())
	if err != nil {
		panic(fmt.Sprintf("database: %v", err))
	}
	if err := health.VerifyRequiredRelations(context.Background(), pool); err != nil {
		panic(fmt.Sprintf("database: %v", err))
	}
	slog.Info("database connected")
	c.DB = pool
}

// injectCSPHashes appends all hash tokens to the script-src directive of csp.
func injectCSPHashes(csp string, hashes []string) string {
	if len(hashes) == 0 {
		return csp
	}
	return strings.Replace(csp, "script-src", "script-src "+strings.Join(hashes, " "), 1)
}

func (c *Container) initWeb() {
	cfg := c.Config
	hashes := []string{interactive.ScriptCSPHash}
	hashes = append(hashes, cfg.CSPHashes.Always...)
	if cfg.IsDevelopment() {
		hashes = append(hashes, cfg.CSPHashes.DevOnly...)
	}
	c.ServerConfig = server.Config{
		Host:            cfg.Server.Host,
		Port:            strconv.Itoa(cfg.Server.Port),
		Debug:           cfg.IsDevelopment(),
		GracefulTimeout: time.Duration(cfg.Server.GracefulTimeout) * time.Second,
		CookieSecure:    cfg.Auth.Session.Secure,
		CSP:             injectCSPHashes(cfg.Server.CSP, hashes),
		CSRFCookieName:  cfg.Server.CSRFCookieName,
		PublicPrefix:    c.AssetPaths.URLPrefix(),
	}
	c.Web = server.New()
	appserver.RegisterMiddleware(c.Web, c.ServerConfig, cfg)
}

func (c *Container) initAuth() {
	sm, err := session.NewSessionStore(session.SessionConfig{
		Name:       c.Config.Auth.Session.Name,
		AuthKey:    c.Config.Auth.Session.AuthKey,
		EncryptKey: c.Config.Auth.Session.EncryptKey,
		MaxAge:     c.Config.Auth.Session.MaxAge,
		Secure:     c.Config.Auth.Session.Secure,
	})
	if err != nil {
		panic(fmt.Sprintf("session: %v", err))
	}
	c.Sessions = sm
}

func (c *Container) initValidator() {
	c.Validator = validate.New()
}

func (c *Container) initRepos() {
	c.Repos = repository.NewRepositories(c.DB)
}

func (c *Container) initServices() {
	c.Services = service.NewServices(c.Repos)

	// Instantiate the auth service from the auth module, wired with adapted repositories.
	authFactory := adapters.NewAuthTxFactory(c.DB)
	c.AuthService = authsvc.NewAuthService(
		adapters.NewAuthUserReader(c.Repos.User),
		adapters.NewAuthPasswordStore(c.Repos.Password),
		authFactory,
		c.DB,
	)
}
