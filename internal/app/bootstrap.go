package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	auth "github.com/go-sum/auth"
	authsvc "github.com/go-sum/auth/service"
	"github.com/go-sum/kv/redisstore"
	"github.com/go-sum/componentry/assetconfig"
	"github.com/go-sum/componentry/assets"
	icons "github.com/go-sum/componentry/icons"
	install "github.com/go-sum/componentry/install"
	"github.com/go-sum/componentry/interactive"
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/adapters/authmail"
	"github.com/go-sum/forge/internal/adapters/authsession"
	"github.com/go-sum/forge/internal/health"
	"github.com/go-sum/session"
	"github.com/go-sum/forge/internal/repository"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/forge/internal/service"
	secheaders "github.com/go-sum/security/headers"
	"github.com/go-sum/send"
	"github.com/go-sum/server"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/logging"
	"github.com/go-sum/server/validate"

	"github.com/labstack/echo/v5"
)

const assetConfigPath = assetconfig.DefaultConfigPath

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

// Database bootstrap.
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

// KV store bootstrap.
func (c *Container) initKV() {
	if !c.Config.App.KV.Enabled {
		slog.Info("kv store disabled")
		return
	}

	cfg := c.Config.App.KV.Redis
	store, err := redisstore.New(redisstore.Config{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})
	if err != nil {
		panic(fmt.Sprintf("kv: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		panic(fmt.Sprintf("kv: ping failed: %v", err))
	}

	slog.Info("kv store connected", "addr", cfg.Addr)
	c.KV = store
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

// Session/auth bootstrap.
func (c *Container) initAuth() {
	cfg := c.Config.App.Session
	mgr, err := session.NewManager(session.Config{
		Store:      cfg.Store,
		CookieName: cfg.Name,
		AuthKey:    cfg.AuthKey,
		EncryptKey: cfg.EncryptKey,
		MaxAge:     cfg.MaxAge,
		Secure:     cfg.Secure,
		BlobStore:  authsession.WrapKV(c.KV),
	})
	if err != nil {
		panic(fmt.Sprintf("session: %v", err))
	}
	c.Sessions = mgr
	slog.Info("session manager initialized", "store", cfg.Store)
}

// Validation bootstrap.
func (c *Container) initValidator() {
	c.Validator = validate.New()
	c.Web.Validator = c.Validator
}

// Repository bootstrap.
func (c *Container) initRepos() {
	c.Repos = repository.NewRepositories(c.DB)
}

// Domain services bootstrap.
func (c *Container) initServices() {
	c.Services = service.NewServices(c.Repos, c.Sender, service.ContactConfig{
		SendTo:   c.Config.Service.Send.SendTo,
		SendFrom: send.DefaultRegistry.SendFrom(c.Config.Service.Send.Delivery),
	})

	if err := auth.ValidateConfig(c.Config.Service.Auth); err != nil {
		panic(fmt.Sprintf("app: auth: %v", err))
	}
	c.AuthService = authsvc.NewAuthService(
		c.Repos.User,
		authsvc.Config{
			Method:   c.Config.Service.Auth.Methods.EmailTOTP,
			Notifier: authmail.New(c.Sender, send.DefaultRegistry.SendFrom(c.Config.Service.Send.Delivery)),
			TokenCodec: authsvc.NewEncryptedTokenCodec(
				c.Config.App.Session.AuthKey,
				c.Config.App.Session.EncryptKey,
			),
		},
	)
}
