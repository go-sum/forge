package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	auth "github.com/go-sum/auth"
	authsvc "github.com/go-sum/auth/service"
	"github.com/go-sum/componentry/assets"
	icons "github.com/go-sum/componentry/icons"
	install "github.com/go-sum/componentry/install"
	"github.com/go-sum/componentry/interactive"
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/adapters/authmail"
	"github.com/go-sum/forge/internal/adapters/authsession"
	"github.com/go-sum/forge/internal/repository"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/kv/redisstore"
	"github.com/go-sum/queue"
	"github.com/go-sum/queue/pgstore"
	secheaders "github.com/go-sum/security/headers"
	"github.com/go-sum/send"
	"github.com/go-sum/server"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/database/migrate"
	"github.com/go-sum/server/logging"
	"github.com/go-sum/server/validate"
	"github.com/go-sum/session"

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

// Database bootstrap.
func (c *Container) initDatabase() {
	pool, err := database.Connect(context.Background(), c.Config.DSN())
	if err != nil {
		c.StartupError = fmt.Errorf("database connect: %w", err)
		slog.Error("database startup check failed", "error", c.StartupError)
		return
	}
	c.DB = pool

	if c.Config.App.Database.AutoMigrate {
		migrationsDir := c.Config.App.Database.MigrationsDir
		if migrationsDir == "" {
			migrationsDir = "db/migrations"
		}
		slog.Info("applying database migrations", "dir", migrationsDir)
		if err := migrate.Up(context.Background(), c.Config.DSN(), migrationsDir); err != nil {
			c.StartupError = fmt.Errorf("database migrate: %w", err)
			slog.Error("migration failed", "error", c.StartupError)
			return
		}
		slog.Info("database migrations complete")
	}

	if err := repository.VerifyRequiredRelations(context.Background(), pool); err != nil {
		c.StartupError = fmt.Errorf("database verify: %w", err)
		slog.Error("database startup check failed", "error", c.StartupError)
		return
	}
	slog.Info("database connected")
}

// Queue bootstrap. Always creates a queue.Client. When queue.enabled is true,
// uses a PostgreSQL-backed store for async processing with workers. When false,
// creates a sync client that executes handlers inline during dispatch.
func (c *Container) initQueue() {
	cfg := c.Config.App.Queue

	queueCfgs := make([]queue.QueueConfig, len(cfg.Queues))
	for i, q := range cfg.Queues {
		queueCfgs[i] = queue.QueueConfig{
			Name:        q.Name,
			Priority:    queue.Priority(q.Priority),
			Workers:     q.Workers,
			MaxAttempts: q.MaxAttempts,
			Timeout:     q.Timeout,
			Backoff:     q.Backoff,
		}
	}

	var store queue.Store
	if cfg.Enabled {
		pg := pgstore.New(pgstore.Config{Pool: c.DB})
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := pg.Install(ctx); err != nil {
			panic(fmt.Sprintf("queue: install schema: %v", err))
		}
		store = pg
	}

	c.Queue = queue.New(store, queue.Config{
		Queues:       queueCfgs,
		PollInterval: cfg.PollInterval,
		ShutdownWait: cfg.ShutdownWait,
	})

	c.AddBackground(c.Queue)

	mode := "sync"
	if cfg.Enabled {
		mode = "async"
	}
	slog.Info("queue initialized", "mode", mode, "queues", len(cfg.Queues))
}

// registerQueueHandlers registers job handlers for each configured queue.
func (c *Container) registerQueueHandlers() {
	c.Queue.Register("email", func(ctx context.Context, job queue.Job) error {
		var p service.EmailPayload
		if err := json.Unmarshal(job.Payload, &p); err != nil {
			return fmt.Errorf("email job: unmarshal: %w", err)
		}
		return c.Sender.Send(ctx, send.Message{
			To:      p.To,
			From:    p.From,
			Subject: p.Subject,
			HTML:    p.HTML,
			Text:    p.Text,
		})
	})
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
	c.registerQueueHandlers()
	c.Services = service.NewServices(c.Repos, c.Queue, service.ContactConfig{
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
