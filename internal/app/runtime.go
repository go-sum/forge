// Package app provides the application composition root and runtime lifecycle.
package app

import (
	"context"
	"log/slog"

	auth "github.com/go-sum/auth"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/componentry/assets"
	"github.com/go-sum/forge/config"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/kv"
	"github.com/go-sum/queue"
	"github.com/go-sum/send"
	"github.com/go-sum/server"
	"github.com/go-sum/server/validate"
	"github.com/go-sum/session"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type AuthStore interface {
	authrepo.UserStore
	authrepo.AdminStore
}

type BackgroundService interface {
	Start(ctx context.Context)
	Stop() error
}

// Runtime owns long-lived infrastructure and lifecycle resources.
type Runtime struct {
	Config       *config.Config
	PublicPrefix string
	DB           *pgxpool.Pool
	StartupError error
	Assets       *assets.Assets
	Web          *echo.Echo
	ServerConfig server.Config
	RateLimiters *appserver.RateLimiters
	PublicDir    string
	Sessions     session.Manager
	Queue        *queue.Client
	KV           kv.Store
	Validator    *validate.Validator
	AuthService  auth.Service
	AuthStore    AuthStore
	Sender       send.Sender

	background []BackgroundService
}

func NewRuntime() *Runtime {
	r := &Runtime{}
	r.initConfig()
	r.initLogger()
	r.initSender()
	r.initAssets()
	r.initWeb()
	r.initDatabase()
	if r.StartupError != nil {
		return r
	}
	r.initAuthStore()
	r.initQueue()
	r.initKV()
	r.initAuth()
	r.initValidator()
	return r
}

func (r *Runtime) AddBackground(svc BackgroundService) {
	r.background = append(r.background, svc)
}

func (r *Runtime) StartBackground(ctx context.Context) {
	for _, svc := range r.background {
		svc.Start(ctx)
	}
}

func (r *Runtime) Shutdown() {
	for i := len(r.background) - 1; i >= 0; i-- {
		if err := r.background[i].Stop(); err != nil {
			slog.Error("background service shutdown error", "index", i, "error", err)
		}
	}
	if r.KV != nil {
		if err := r.KV.Close(); err != nil {
			slog.Error("kv close error", "error", err)
		}
	}
	if r.DB != nil {
		r.DB.Close()
	}
	slog.Info("runtime shutdown complete")
}
