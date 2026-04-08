package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	authpgstore "github.com/go-sum/auth/pgstore"
	"github.com/go-sum/forge/internal/features/contact"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/kv/redisstore"
	"github.com/go-sum/queue"
	"github.com/go-sum/queue/pgstore"
	"github.com/go-sum/send"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/database/migrate"
)

// Database bootstrap.
func (r *Runtime) initDatabase() {
	pool, err := database.Connect(context.Background(), r.Config.Store.Database.URL, r.Config.Store.Database.MaxConns)
	if err != nil {
		r.StartupError = fmt.Errorf("database connect: %w", err)
		slog.Error("database startup check failed", "error", r.StartupError)
		return
	}
	r.DB = pool

	if r.Config.Store.Database.AutoMigrate {
		migrationsDir := r.Config.Store.Database.MigrationsDir
		slog.Info("applying database migrations", "dir", migrationsDir)
		if err := migrate.Up(context.Background(), r.Config.Store.Database.URL, migrationsDir); err != nil {
			r.StartupError = fmt.Errorf("database migrate: %w", err)
			slog.Error("migration failed", "error", r.StartupError)
			return
		}
		slog.Info("database migrations complete")
	}

	if err := repository.VerifyRequiredRelations(context.Background(), pool); err != nil {
		r.StartupError = fmt.Errorf("database verify: %w", err)
		slog.Error("database startup check failed", "error", r.StartupError)
		return
	}
	slog.Info("database connected")
}

// Auth store bootstrap. Requires the users table to already exist via
// migrations. The PgStore implements both authrepo.UserStore (for auth flows)
// and authrepo.AdminStore (for admin user management).
func (r *Runtime) initAuthStore() {
	r.AuthStore = authpgstore.New(authpgstore.Config{Pool: r.DB})
	slog.Info("auth store initialized")
}

// Queue bootstrap. Always creates a queue.Client. When queue.enabled is true,
// uses a PostgreSQL-backed store for async processing with workers. When false,
// creates a sync client that executes handlers inline during dispatch.
func (r *Runtime) initQueue() {
	cfg := r.Config.Store.Queue

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
		store = pgstore.New(pgstore.Config{Pool: r.DB, ReapThreshold: cfg.ReapThreshold})
	}

	r.Queue = queue.New(store, queue.Config{
		Queues:       queueCfgs,
		PollInterval: cfg.PollInterval,
		ShutdownWait: cfg.ShutdownWait,
	})

	r.AddBackground(r.Queue)

	mode := "sync"
	if cfg.Enabled {
		mode = "async"
	}
	slog.Info("queue initialized", "mode", mode, "queues", len(cfg.Queues))
}

// registerQueueHandlers registers job handlers for each configured queue.
func (r *Runtime) registerQueueHandlers() {
	r.Queue.Register("email", func(ctx context.Context, job queue.Job) error {
		var p contact.EmailPayload
		if err := json.Unmarshal(job.Payload, &p); err != nil {
			return fmt.Errorf("email job: unmarshal: %w", err)
		}
		return r.Sender.Send(ctx, send.Message{
			To:      p.To,
			From:    p.From,
			Subject: p.Subject,
			HTML:    p.HTML,
			Text:    p.Text,
		})
	})
}

// KV store bootstrap.
func (r *Runtime) initKV() {
	if !r.Config.Store.KV.Enabled {
		slog.Info("kv store disabled")
		return
	}

	cfg := r.Config.Store.KV.Redis
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
	r.KV = store
}
