package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	authpgstore "github.com/go-sum/auth/pgstore"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/kv/redisstore"
	"github.com/go-sum/queue"
	"github.com/go-sum/queue/pgstore"
	"github.com/go-sum/send"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/database/migrate"
)

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

// Auth store bootstrap. Requires the users table to already exist via
// migrations. The PgStore implements both authrepo.UserStore (for auth flows)
// and authrepo.AdminStore (for admin user management).
func (c *Container) initAuthStore() {
	c.AuthStore = authpgstore.New(authpgstore.Config{Pool: c.DB})
	slog.Info("auth store initialized")
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
		store = pgstore.New(pgstore.Config{Pool: c.DB})
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
