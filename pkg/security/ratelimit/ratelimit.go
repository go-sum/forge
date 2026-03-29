// Package ratelimit provides IP-based rate limiting middleware for Echo v5.
//
// It wraps a configurable Store behind an Echo-idiomatic Config type with
// Skipper, BeforeFunc, DenyHandler, and ErrorHandler hooks. All failure paths
// return typed errors that implement StatusCode() int and PublicMessage() string
// so that the application's error-classification pipeline emits the correct HTTP
// status (429 Too Many Requests) rather than falling through to a 500.
package ratelimit

import (
	"errors"
	"log/slog"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

// Skipper defines a function to skip middleware. Returning true skips processing.
type Skipper func(c *echo.Context) bool

// BeforeFunc defines a function executed just before the middleware.
type BeforeFunc func(c *echo.Context)

// Store is the interface for custom rate limit backends.
// Allow returns true if the identifier is within its rate limit.
type Store interface {
	Allow(identifier string) (bool, error)
}

// Config defines the rate limiting configuration aligned with Echo v5 conventions.
type Config struct {
	// Skipper defines a function to skip middleware. Defaults to never skipping.
	Skipper Skipper

	// BeforeFunc defines a function executed before the rate limit check.
	BeforeFunc BeforeFunc

	// IdentifierExtractor extracts the rate limit key from the request context.
	// Defaults to the remote IP address when nil.
	IdentifierExtractor func(c *echo.Context) (string, error)

	// Store applies the rate limit policy. Required — use NewMemoryStoreWithConfig to build one.
	Store Store

	// DenyHandler is called when the store denies a request.
	// Defaults to returning a typed 429 violation.
	DenyHandler func(c *echo.Context, identifier string, err error) error

	// ErrorHandler is called when IdentifierExtractor returns an error.
	// Defaults to returning a typed 500 violation.
	ErrorHandler func(c *echo.Context, err error) error
}

// ToMiddleware converts Config to an echo.MiddlewareFunc, returning an error
// for invalid configuration (e.g. nil Store) rather than panicking.
// This satisfies the echo.MiddlewareConfigurator interface.
func (cfg Config) ToMiddleware() (echo.MiddlewareFunc, error) {
	if cfg.Skipper == nil {
		cfg.Skipper = func(*echo.Context) bool { return false }
	}
	if cfg.IdentifierExtractor == nil {
		cfg.IdentifierExtractor = defaultIPExtractor
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *echo.Context, err error) error {
			return &violation{code: http.StatusInternalServerError, msg: "Service temporarily unavailable"}
		}
	}
	if cfg.DenyHandler == nil {
		cfg.DenyHandler = func(c *echo.Context, identifier string, err error) error {
			slog.Default().Debug("rate limit exceeded",
				"ip", identifier,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
			)
			return &violation{code: http.StatusTooManyRequests, msg: denyMessage}
		}
	}
	if cfg.Store == nil {
		return nil, errors.New("ratelimit: Store must be provided")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}
			if cfg.BeforeFunc != nil {
				cfg.BeforeFunc(c)
			}

			identifier, err := cfg.IdentifierExtractor(c)
			if err != nil {
				return cfg.ErrorHandler(c, err)
			}

			allowed, allowErr := cfg.Store.Allow(identifier)
			if !allowed {
				return cfg.DenyHandler(c, identifier, allowErr)
			}

			return next(c)
		}
	}, nil
}

// Middleware returns an Echo middleware that enforces a per-identifier request
// rate using the configured store. Panics on invalid configuration.
//
// Requests exceeding the configured rate return a typed 429 Too Many Requests.
func Middleware(cfg Config) echo.MiddlewareFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// violation is a typed rate-limit failure value. It implements StatusCode() and
// PublicMessage() so that internal/server.classify() emits the correct HTTP
// status response rather than falling through to a 500 Internal Server Error.
type violation struct {
	code int
	msg  string
}

func (v *violation) Error() string         { return v.msg }
func (v *violation) StatusCode() int       { return v.code }
func (v *violation) PublicMessage() string { return v.msg }

const denyMessage = "Too many requests. Please wait and try again."

// defaultIPExtractor reads the real IP from the context, stripping any port suffix.
func defaultIPExtractor(c *echo.Context) (string, error) {
	ip := c.RealIP()
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host, nil
	}
	return ip, nil
}

// MemoryStoreConfig holds construction parameters for MemoryStore.
type MemoryStoreConfig struct {
	Rate      float64       // requests per second (token bucket refill)
	Burst     int           // maximum burst size; 0 → ceil(Rate), min 1
	ExpiresIn time.Duration // duration before stale visitors are eligible for cleanup; 0 → 3m
}

// MemoryStore is an in-process token-bucket store keyed by identifier.
// Stale entries are lazily swept when the time since the last cleanup
// exceeds ExpiresIn, preventing unbounded memory growth.
type MemoryStore struct {
	visitors    map[string]*visitor
	mu          sync.Mutex
	rate        float64
	burst       int
	expiresIn   time.Duration
	lastCleanup time.Time
}

// visitor tracks a unique identifier's rate limiter and last activity time.
type visitor struct {
	*rate.Limiter
	lastSeen time.Time
}

// NewMemoryStoreWithConfig returns a MemoryStore configured with the provided
// MemoryStoreConfig. Burst defaults to ceil(Rate) (min 1) when zero.
// ExpiresIn defaults to 3 minutes when zero.
func NewMemoryStoreWithConfig(cfg MemoryStoreConfig) *MemoryStore {
	expiresIn := cfg.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3 * time.Minute
	}
	burst := cfg.Burst
	if burst <= 0 {
		burst = int(math.Max(1, math.Ceil(cfg.Rate)))
	}
	now := time.Now()
	return &MemoryStore{
		visitors:    make(map[string]*visitor),
		rate:        cfg.Rate,
		burst:       burst,
		expiresIn:   expiresIn,
		lastCleanup: now,
	}
}

// NewMemoryStore returns a MemoryStore with the given rate (req/s).
// Burst defaults to ceil(Rate) (min 1); ExpiresIn defaults to 3 minutes.
func NewMemoryStore(rateLimit float64) *MemoryStore {
	return NewMemoryStoreWithConfig(MemoryStoreConfig{Rate: rateLimit})
}

// Allow implements Store. It returns true if the identifier is within its rate
// limit. Stale visitor entries are lazily cleaned up after ExpiresIn elapses.
func (s *MemoryStore) Allow(identifier string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	v, ok := s.visitors[identifier]
	if !ok {
		v = &visitor{
			Limiter: rate.NewLimiter(rate.Limit(s.rate), s.burst),
		}
		s.visitors[identifier] = v
	}
	v.lastSeen = now

	if now.Sub(s.lastCleanup) > s.expiresIn {
		s.cleanupStaleVisitors(now)
	}

	return v.AllowN(now, 1), nil
}

// cleanupStaleVisitors removes visitor entries that have not been seen within
// the expiry window. Must be called with s.mu held.
func (s *MemoryStore) cleanupStaleVisitors(now time.Time) {
	for id, v := range s.visitors {
		if now.Sub(v.lastSeen) > s.expiresIn {
			delete(s.visitors, id)
		}
	}
	s.lastCleanup = now
}
