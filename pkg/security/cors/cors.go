// Package cors provides CORS middleware for Echo v5 with three origin-matching
// strategies: exact string list, compiled regexp set, and custom function.
//
// All preflight handling, Vary:Origin, Access-Control-Allow-Credentials, and
// Access-Control-Max-Age logic is delegated to Echo's built-in CORSConfig.
// This package adds regex origin matching and a unified Config surface over the
// three modes.
package cors

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

// OriginMode selects the origin-matching strategy used by Config.
type OriginMode int

const (
	// OriginModeExact matches origins against AllowOrigins using case-insensitive
	// string comparison. The special value "*" permits all origins (AllowCredentials
	// must be false when "*" is used).
	OriginModeExact OriginMode = iota

	// OriginModeRegex matches the request Origin header against each compiled
	// pattern in RegexOrigins. Patterns are compiled once at middleware creation
	// time; an invalid pattern causes ToMiddleware to return an error.
	OriginModeRegex

	// OriginModeFunc delegates origin validation to AllowOriginFunc. The function
	// receives the Echo context and the raw Origin header value and returns the
	// allowed origin string, whether it is permitted, and any error.
	OriginModeFunc
)

// Config defines the CORS middleware configuration.
type Config struct {
	// Skipper defines a function to skip the middleware. Defaults to never skip.
	Skipper func(c *echo.Context) bool

	// Mode selects the origin-matching strategy. See OriginMode constants.
	Mode OriginMode

	// AllowOrigins is the list of allowed origins for OriginModeExact.
	// Required when Mode == OriginModeExact. Supports "*" for wildcard (cannot
	// be combined with AllowCredentials == true).
	AllowOrigins []string

	// RegexOrigins is the list of regexp patterns evaluated against the request
	// Origin for OriginModeRegex. Patterns are compiled at middleware creation
	// time. Required when Mode == OriginModeRegex.
	RegexOrigins []string

	// AllowOriginFunc is the custom origin validator for OriginModeFunc.
	// Returns the allowed origin, whether it is permitted, and any error.
	// Required when Mode == OriginModeFunc.
	AllowOriginFunc func(c *echo.Context, origin string) (string, bool, error)

	// AllowMethods is the list of HTTP methods permitted cross-origin.
	// Defaults to GET, HEAD, PUT, PATCH, POST, DELETE when empty.
	AllowMethods []string

	// AllowHeaders is the list of request headers permitted cross-origin.
	// Defaults to empty (browser's requested headers are echoed back for
	// preflight when this is empty).
	AllowHeaders []string

	// AllowCredentials permits cookies and authorization headers cross-origin.
	// Cannot be combined with AllowOrigins = ["*"].
	AllowCredentials bool

	// ExposeHeaders lists response headers that browsers are permitted to read.
	ExposeHeaders []string

	// MaxAge is the preflight cache duration in seconds.
	// 0 means the header is not sent; negative value sends "0".
	MaxAge int
}

// Middleware returns an Echo middleware configured with cfg.
// Panics if cfg fails validation (matches the ratelimit/csrf pattern).
func Middleware(cfg Config) echo.MiddlewareFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// ToMiddleware converts Config to an echo.MiddlewareFunc or returns an error
// for invalid configuration. Satisfies the echo.MiddlewareConfigurator interface.
//
// For OriginModeRegex, all patterns are compiled here. An invalid pattern
// causes an immediate error rather than a per-request panic.
func (cfg Config) ToMiddleware() (echo.MiddlewareFunc, error) {
	if cfg.Skipper == nil {
		cfg.Skipper = func(*echo.Context) bool { return false }
	}

	echoCfg := echomw.CORSConfig{
		Skipper:          cfg.Skipper,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		AllowCredentials: cfg.AllowCredentials,
		ExposeHeaders:    cfg.ExposeHeaders,
		MaxAge:           cfg.MaxAge,
	}

	switch cfg.Mode {
	case OriginModeExact:
		if len(cfg.AllowOrigins) == 0 {
			return nil, errors.New("cors: OriginModeExact requires at least one AllowOrigins entry")
		}
		// Delegate exact matching (including "*" and credentials guard) to Echo.
		echoCfg.AllowOrigins = cfg.AllowOrigins

	case OriginModeRegex:
		if len(cfg.RegexOrigins) == 0 {
			return nil, errors.New("cors: OriginModeRegex requires at least one RegexOrigins entry")
		}
		compiled := make([]*regexp.Regexp, 0, len(cfg.RegexOrigins))
		for _, pat := range cfg.RegexOrigins {
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, fmt.Errorf("cors: invalid RegexOrigins pattern %q: %w", pat, err)
			}
			compiled = append(compiled, re)
		}
		echoCfg.UnsafeAllowOriginFunc = func(_ *echo.Context, origin string) (string, bool, error) {
			for _, re := range compiled {
				if re.MatchString(origin) {
					return origin, true, nil
				}
			}
			return "", false, nil
		}

	case OriginModeFunc:
		if cfg.AllowOriginFunc == nil {
			return nil, errors.New("cors: OriginModeFunc requires AllowOriginFunc to be set")
		}
		echoCfg.UnsafeAllowOriginFunc = cfg.AllowOriginFunc

	default:
		return nil, fmt.Errorf("cors: unknown OriginMode %d", cfg.Mode)
	}

	return echoCfg.ToMiddleware()
}
