package server

import (
	"log/slog"
	"time"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/security/ratelimit"
	"github.com/labstack/echo/v5"
)

type RateLimiters struct {
	stores map[string]ratelimit.Store
}

func NewRateLimiters(cfg *config.Config) *RateLimiters {
	backend := cfg.Security.RateLimitBackend.Selected

	stores := make(map[string]ratelimit.Store, len(cfg.Security.RateLimits))
	for name, policy := range cfg.Security.RateLimits {
		if policy.Rate == 0 {
			continue
		}
		stores[name] = newRateLimitStore(backend, policy)
	}

	return &RateLimiters{stores: stores}
}

func (r *RateLimiters) Middleware(cfg *config.Config, name string) echo.MiddlewareFunc {
	rl, ok := cfg.Security.RateLimits[name]
	if !ok || rl.Rate == 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}

	store := r.stores[name]
	if store == nil {
		store = ratelimit.NewMemoryStoreWithConfig(ratelimit.MemoryStoreConfig{
			Rate:      rl.Rate,
			Burst:     rl.Burst,
			ExpiresIn: time.Duration(rl.ExpiresIn) * time.Second,
		})
	}

	return ratelimit.Middleware(ratelimit.Config{
		Store: store,
	})
}

func newRateLimitStore(backend string, policy config.RateLimitConfig) ratelimit.Store {
	switch backend {
	case "", "memory":
		return ratelimit.NewMemoryStoreWithConfig(ratelimit.MemoryStoreConfig{
			Rate:      policy.Rate,
			Burst:     policy.Burst,
			ExpiresIn: time.Duration(policy.ExpiresIn) * time.Second,
		})
	default:
		slog.Warn("unsupported rate limit backend, falling back to in-memory store", "backend", backend)
		return ratelimit.NewMemoryStoreWithConfig(ratelimit.MemoryStoreConfig{
			Rate:      policy.Rate,
			Burst:     policy.Burst,
			ExpiresIn: time.Duration(policy.ExpiresIn) * time.Second,
		})
	}
}
