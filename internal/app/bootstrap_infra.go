package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/go-sum/componentry/assets"
	icons "github.com/go-sum/componentry/icons"
	install "github.com/go-sum/componentry/install"
	"github.com/go-sum/componentry/interactive"
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/forge/config"
	appserver "github.com/go-sum/forge/internal/server"
	secheaders "github.com/go-sum/security/headers"
	"github.com/go-sum/send"
	"github.com/go-sum/server"
	"github.com/go-sum/server/logging"

	"github.com/labstack/echo/v5"
)

const (
	defaultPublicDir    = "public"
	defaultPublicPrefix = "/public"
)

var componentIconOverrides = map[icons.Key]icons.Ref{}

// Config bootstrap.
func (c *Container) initConfig() {
	cfg, err := config.Load(os.Getenv("APP_ENV"))
	if err != nil {
		panic(fmt.Sprintf("config: %v", err))
	}
	config.App = cfg
	c.Config = cfg
}

// Logging bootstrap.
func (c *Container) initLogger() {
	logging.Init(logging.Config{
		Development: c.Config.IsDevelopment(),
		Level:       c.Config.App.Log.Level,
	})
	slog.Info("logger initialized", "level", c.Config.App.Log.Level, "env", c.Config.App.Env)
}

// Sender bootstrap.
func (c *Container) initSender() {
	sender, err := send.New(c.Config.Service.Send.Delivery)
	if err != nil {
		panic(fmt.Sprintf("app: sender: %v", err))
	}
	c.Sender = sender
}

// Assets bootstrap.
func (c *Container) initAssets() {
	publicDir := c.Config.App.Assets.PublicDir
	if publicDir == "" {
		publicDir = defaultPublicDir
	}
	publicPrefix := c.Config.App.Assets.PublicPrefix
	if publicPrefix == "" {
		publicPrefix = defaultPublicPrefix
	}

	if err := assets.Init(publicDir, publicPrefix); err != nil {
		panic(fmt.Sprintf("assets: %v", err))
	}

	install.ApplyDefault(install.Config{
		PathFunc:      assets.Path,
		IconOverrides: componentIconOverrides,
	})

	c.Assets = assets.Default
	c.PublicDir = publicDir
	c.PublicPrefix = publicPrefix
}

// checkHealth returns a closure that reports database reachability.
func (c *Container) checkHealth() func(context.Context) error {
	return func(ctx context.Context) error {
		if c.StartupError != nil {
			return c.StartupError
		}
		if c.DB == nil {
			return errors.New("database pool is not initialized")
		}
		return c.DB.Ping(ctx)
	}
}

// Web bootstrap.
func (c *Container) initWeb() {
	cfg := c.Config
	scriptHashes := []string{interactive.ScriptCSPHash}
	scriptHashes = append(scriptHashes, cfg.App.CSPHashes.Always...)
	if cfg.IsDevelopment() {
		scriptHashes = append(scriptHashes, cfg.App.CSPHashes.DevOnly...)
	}

	fontCSP := font.CollectCSPSources(font.BuildProviders(cfg.Site.Fonts, assets.Path))
	styleSrcs := append(fontCSP.StyleSources, fontCSP.StyleInlineHashes...)
	processedCSP := secheaders.InjectDirectiveSources(cfg.App.Security.Headers.ContentSecurityPolicy, "script-src", scriptHashes)
	processedCSP = secheaders.InjectDirectiveSources(processedCSP, "style-src", styleSrcs)
	processedCSP = secheaders.InjectDirectiveSources(processedCSP, "font-src", fontCSP.FontSources)

	c.ServerConfig = server.Config{
		Host:            cfg.App.Server.Host,
		Port:            strconv.Itoa(cfg.App.Server.Port),
		GracefulTimeout: time.Duration(cfg.App.Server.GracefulTimeout) * time.Second,
	}
	c.RateLimiters = appserver.NewRateLimiters(cfg)
	c.Web = server.NewWithConfig(echo.Config{
		HTTPErrorHandler: appserver.NewErrorHandler(appserver.ErrorHandlerConfig{
			Debug:  cfg.IsDevelopment(),
			Logger: slog.Default(),
			Config: cfg,
		}),
	})
	appserver.RegisterMiddleware(c.Web, cfg, processedCSP)
}
