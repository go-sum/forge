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
func (r *Runtime) initConfig() {
	cfg, err := config.Load(os.Getenv("APP_ENV"))
	if err != nil {
		panic(fmt.Sprintf("config: %v", err))
	}
	config.App = cfg
	r.Config = cfg
}

// Logging bootstrap.
func (r *Runtime) initLogger() {
	logging.Init(logging.Config{
		Development: r.Config.IsDevelopment(),
		Level:       r.Config.App.Log.Level,
	})
	slog.Info("logger initialized", "level", r.Config.App.Log.Level, "env", r.Config.App.Env)
}

// Sender bootstrap.
func (r *Runtime) initSender() {
	sender, err := send.New(r.Config.Service.Send.Delivery)
	if err != nil {
		panic(fmt.Sprintf("app: sender: %v", err))
	}
	r.Sender = sender
}

// Assets bootstrap.
func (r *Runtime) initAssets() {
	publicDir := r.Config.App.Assets.PublicDir
	if publicDir == "" {
		publicDir = defaultPublicDir
	}
	publicPrefix := r.Config.App.Assets.PublicPrefix
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

	r.Assets = assets.Default
	r.PublicDir = publicDir
	r.PublicPrefix = publicPrefix
}

// checkHealth returns a closure that reports database reachability.
func (r *Runtime) checkHealth() func(context.Context) error {
	return func(ctx context.Context) error {
		if r.StartupError != nil {
			return r.StartupError
		}
		if r.DB == nil {
			return errors.New("database pool is not initialized")
		}
		return r.DB.Ping(ctx)
	}
}

// Web bootstrap.
func (r *Runtime) initWeb() {
	cfg := r.Config
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

	r.ServerConfig = server.Config{
		Host:            cfg.App.Server.Host,
		Port:            strconv.Itoa(cfg.App.Server.Port),
		GracefulTimeout: time.Duration(cfg.App.Server.GracefulTimeout) * time.Second,
	}
	r.RateLimiters = appserver.NewRateLimiters(cfg)
	r.Web = server.NewWithConfig(echo.Config{
		HTTPErrorHandler: appserver.NewErrorHandler(appserver.ErrorHandlerConfig{
			Debug:  cfg.IsDevelopment(),
			Logger: slog.Default(),
			Config: cfg,
		}),
	})
	appserver.RegisterMiddleware(r.Web, cfg, processedCSP)
}
