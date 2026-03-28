---
title: Server infrastructure
description: Infrastructure packages for building web applications with Echo v5 and PostgreSQL.
weight: 20
---

# server

`github.com/go-sum/server` is a collection of infrastructure packages for building web applications with [Echo] and PostgreSQL. All sub-packages live under a single `go.mod` and follow the **leaf-node rule**: they import only the standard library and external modules -- never application-specific `internal/` code. This makes the entire module portable to any [Echo] project.

## Dependencies

| Dependency | Version |
|------------|---------|
| [Echo] | v5.0 |
| [koanf] | v2.3 |
| [pgx] | v5.8 |
| [validator] | v10.30 |

## Features

- Lightweight [Echo] instance factory with graceful shutdown driven by OS signals
- Structured application errors with safe public messages and internal cause tracking
- Generic, layered YAML configuration loader with environment variable expansion and struct validation
- PostgreSQL connection pool management via `pgxpool` with safe defaults
- Structured logging backed by `log/slog` with development and production modes
- Static asset cache-control middleware with content-hash-aware immutable headers
- Named route registration and type-safe URL reversal utilities
- Struct and field validation via [validator]

## Sub-packages at a Glance

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `server` | `github.com/go-sum/server` | [Echo] instance factory and graceful-shutdown lifecycle |
| `apperr` | `github.com/go-sum/server/apperr` | Typed application errors with HTTP status codes and safe messages |
| `config` | `github.com/go-sum/server/config` | Generic layered YAML config loader with `${VAR}` expansion |
| `database` | `github.com/go-sum/server/database` | PostgreSQL `pgxpool` connection, health check, and helpers |
| `logging` | `github.com/go-sum/server/logging` | `slog`-based structured logging with dev/prod modes |
| `middleware` | `github.com/go-sum/server/middleware` | Reusable Echo middleware (static asset caching) |
| `route` | `github.com/go-sum/server/route` | Named route registration and URL reversal |
| `validate` | `github.com/go-sum/server/validate` | Struct and field validation wrapper |

---

## server (root package)

Creates a bare [Echo] instance and manages server lifecycle with graceful shutdown.

### Types

**`Config`** -- server startup configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Host` | `string` | Server hostname or IP address |
| `Port` | `string` | Server port |
| `GracefulTimeout` | `time.Duration` | Maximum time to wait for in-flight requests during shutdown |

### Functions

**`New() *echo.Echo`** -- creates a bare [Echo] instance with no middleware attached. Application-specific middleware and routing are wired separately.

**`Start(e *echo.Echo, cfg Config) error`** -- begins listening on `Host:Port` and blocks until `SIGINT` or `SIGTERM` is received. Performs graceful shutdown within `GracefulTimeout`. Returns an error (prefixed with `"server:"`) rather than calling `os.Exit`, so deferred cleanup in `main` (such as closing database pools) runs normally. Logs startup and shutdown events via `slog`.

```go
e := server.New()
// ... register middleware and routes ...
err := server.Start(e, server.Config{
    Host:            "0.0.0.0",
    Port:            "8080",
    GracefulTimeout: 10 * time.Second,
})
if err != nil {
    slog.Error("server exited", "error", err)
}
```

---

## apperr

Provides typed application errors that carry an HTTP status code, a domain-agnostic error code, and a safe user-facing message. The underlying `Cause` error is kept for server-side logging but never exposed to clients.

### Types

**`Code`** (`string`) -- domain-agnostic error classifier.

| Constant | Value |
|----------|-------|
| `CodeBadRequest` | `"bad_request"` |
| `CodeUnauthorized` | `"unauthorized"` |
| `CodeForbidden` | `"forbidden"` |
| `CodeNotFound` | `"not_found"` |
| `CodeConflict` | `"conflict"` |
| `CodeValidationFailed` | `"validation_failed"` |
| `CodeServiceUnavailable` | `"service_unavailable"` |
| `CodeInternal` | `"internal_error"` |

**`Error`** -- implements the `error` interface.

| Field | Type | Description |
|-------|------|-------------|
| `Status` | `int` | HTTP status code |
| `Code` | `Code` | Domain-agnostic error classifier |
| `Title` | `string` | Safe user-facing title (defaults to `http.StatusText` if empty at creation) |
| `Message` | `string` | Safe user-facing detail |
| `Cause` | `error` | Underlying infrastructure error (logged server-side, never sent to clients) |

Methods:

- `Error() string` -- returns `Cause.Error()` if present, otherwise `Message`, otherwise `Title`
- `Unwrap() error` -- returns `Cause` for use with `errors.Is` / `errors.As`
- `StatusCode() int` -- returns `Status`
- `PublicMessage() string` -- returns the safe message suitable for client responses

### Factory Functions

| Function | Signature | HTTP Status |
|----------|-----------|-------------|
| `BadRequest` | `BadRequest(message string) *Error` | 400 |
| `Unauthorized` | `Unauthorized(message string) *Error` | 401 |
| `Forbidden` | `Forbidden(message string) *Error` | 403 |
| `NotFound` | `NotFound(message string) *Error` | 404 |
| `Conflict` | `Conflict(message string) *Error` | 409 |
| `Validation` | `Validation(message string) *Error` | 422 |
| `Unavailable` | `Unavailable(message string, cause error) *Error` | 503 |
| `Internal` | `Internal(cause error) *Error` | 500 |
| `New` | `New(status int, code Code, title, message string, cause error) *Error` | custom |

`Internal` sets a fixed public message: `"Something went wrong on our side. Please try again."`

### Classification Helpers

**`From(err error) *Error`** -- extracts a `*Error` from anywhere in the error chain using `errors.As`. Returns `nil` if no `*Error` is found.

**`Resolve(err error) *Error`** -- like `From`, but falls back to `Internal(err)` instead of returning `nil`. Guarantees a non-nil `*Error` for any non-nil input.

```go
// In an error handler:
appErr := apperr.Resolve(err)
slog.Error("request failed", "cause", appErr.Cause)
c.JSON(appErr.StatusCode(), map[string]string{
    "message": appErr.PublicMessage(),
})
```

### Error Classification Pattern

1. Repositories return domain errors (e.g., `model.ErrUserNotFound`)
2. Handlers map domain errors to `apperr` errors (e.g., `apperr.NotFound(...)`)
3. A centralized error handler calls `apperr.From()` or `apperr.Resolve()` to classify
4. `Cause` is logged server-side; only `PublicMessage()` reaches the client

---

## config

Generic, layered YAML configuration loader built on [koanf]. Supports environment-specific overlays, `${VAR}` expansion, and struct validation.

### Types

**`ConfigFile`** -- a single configuration file entry.

| Field | Type | Description |
|-------|------|-------------|
| `Filepath` | `string` | Path to the YAML file |

**`Options`** -- controls how `Load` discovers and merges configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Files` | `[]ConfigFile` | Ordered list of config files. `Files[0]` is required; the rest are optional. |
| `EnvKey` | `string` | Active environment name (e.g., `"development"`). Triggers an overlay lookup. |

### Functions

**`Load[T any](opts func(*T) Options) (*T, error)`** -- allocates a `*T`, calls the options function to determine files and environment, then loads and validates the configuration.

### Loading Order (last writer wins)

1. `Files[0].Filepath` -- required base config; returns an error if missing
2. `{dir}/{stem}.{EnvKey}.yaml` -- optional environment overlay; silently skipped if absent
3. `Files[1:]` -- optional extra files; silently skipped if absent
4. Unmarshal merged YAML into `*T` via [koanf] struct tags
5. Validate `*T` using [validator] struct tags

### Variable Expansion

Every YAML file supports `${VAR}` and `${VAR:-default}` syntax. Environment variables are expanded before parsing. Unset variables without a default expand to an empty string.

```yaml
# config/app.yaml
app:
  session_key: ${AUTH_SESSION_ENCRYPT_KEY}
  log_level: ${LOG_LEVEL:-info}
```

```go
type AppConfig struct {
    App  AppSettings  `koanf:"app"`
    Site SiteSettings `koanf:"site"`
}

cfg, err := config.Load(func(c *AppConfig) config.Options {
    return config.Options{
        Files: []config.ConfigFile{
            {Filepath: "config/app.yaml"},
            {Filepath: "config/site.yaml"},
        },
        EnvKey: os.Getenv("APP_ENV"),
    }
})
```

---

## database

PostgreSQL connection pool management via `pgxpool` ([pgx]).

### Functions

**`Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error)`** -- parses the DSN, creates a connection pool, and pings the database to verify connectivity. Applies a safe default of `MaxConns=10` when the DSN does not specify `pool_max_conns`. This prevents silently exhausting PostgreSQL's default `max_connections=100` across multiple application instances. The DSN parameter `?pool_max_conns=N` takes precedence.

**`CheckHealth(ctx context.Context, pool *pgxpool.Pool) error`** -- verifies pool health by pinging the database.

**`Close(pool *pgxpool.Pool)`** -- gracefully closes the connection pool.

**`IsUniqueViolation(err error) bool`** -- returns `true` if the error is a PostgreSQL unique constraint violation (SQLSTATE `23505`). Use this in repository code to map database errors to domain errors without duplicating the error code string.

```go
pool, err := database.Connect(ctx, "postgres://user:pass@localhost:5432/mydb?pool_max_conns=20")
if err != nil {
    return fmt.Errorf("startup: %w", err)
}
defer database.Close(pool)

// In a repository:
_, err = q.InsertUser(ctx, params)
if database.IsUniqueViolation(err) {
    return model.ErrEmailTaken
}
```

---

## logging

Structured logging backed by `log/slog`. Produces human-readable text output in development and structured JSON in production.

### Types

**`Config`** -- controls logger construction.

| Field | Type | Description |
|-------|------|-------------|
| `Development` | `bool` | `true` = `slog.TextHandler` (stderr); `false` = `slog.JSONHandler` (stdout) |
| `Level` | `string` | `"debug"`, `"info"`, `"warn"`, or `"error"` (defaults to `"info"`) |
| `TextOutput` | `io.Writer` | Output destination for text mode (defaults to `os.Stderr`) |
| `JSONOutput` | `io.Writer` | Output destination for JSON mode (defaults to `os.Stdout`) |

### Functions

**`New(cfg Config) *slog.Logger`** -- creates a logger without modifying global state.

**`Init(cfg Config) *slog.Logger`** -- creates a logger and installs it as the global `slog` default via `slog.SetDefault`. Returns the logger for optional direct use.

```go
// Development: human-readable text on stderr
logging.Init(logging.Config{
    Development: true,
    Level:       "debug",
})

// Production: structured JSON on stdout
logging.Init(logging.Config{
    Level: "info",
})
```

---

## middleware

Reusable [Echo] middleware functions. Each middleware carries no application-specific imports.

### Functions

**`StaticCacheControl(prefix string) echo.MiddlewareFunc`** -- sets `Cache-Control` headers for static asset paths under `prefix`.

| Condition | Cache-Control Value |
|-----------|-------------------|
| Path matches `prefix/` and URL has `?v=<hash>` | `public, max-age=31536000, immutable` (1 year) |
| Path matches `prefix/` and URL has no `?v` param | `no-cache` |
| Path does not match `prefix/` | No header set |
| Empty `prefix` | Middleware is silently disabled |

Content-hashed URLs (e.g., `/public/app.css?v=abc12345`) receive an immutable one-year cache header, eliminating revalidation requests for assets that change URL on every build.

```go
e.Use(middleware.StaticCacheControl("/public"))
```

```html
<!-- Versioned asset: cached for 1 year -->
<link href="/public/app.css?v=abc12345" rel="stylesheet">

<!-- Unversioned asset: always revalidated -->
<link href="/public/favicon.ico" rel="icon">
```

---

## route

Named route registration and type-safe URL reversal for [Echo]. Enforces the convention that every route has a `Name`.

### Types

**`RouteAdder`** -- interface satisfied by `*echo.Echo` and `*echo.Group`.

```go
type RouteAdder interface {
    AddRoute(route echo.Route) (echo.RouteInfo, error)
}
```

### Functions

**`Add(target RouteAdder, route echo.Route) echo.RouteInfo`** -- registers a route on the target. Panics if `Name` is empty or registration fails. This enforces the naming convention at startup rather than allowing unnamed routes to slip through.

**`Reverse(routes echo.Routes, name string, pathValues ...any) string`** -- resolves a named route to its URL path. Panics if the name is unknown. Use this when the route is guaranteed to exist.

**`ReverseWithQuery(routes echo.Routes, name string, query url.Values, pathValues ...any) string`** -- resolves a named route and appends the given query parameters.

**`SafeReverse(routes echo.Routes, name string) (string, bool)`** -- resolves a named route without panicking. Returns `("", false)` if the name is unknown or the resolved path still contains unfilled `:param` segments. Useful for sitemap generation where parameterized routes should be skipped.

### Naming Convention

Routes follow the pattern `<resource>.<action>`:

```go
route.Add(g, echo.Route{
    Method:  http.MethodGet,
    Path:    "/users",
    Name:    "user.list",
    Handler: h.UserList,
})

route.Add(g, echo.Route{
    Method:  http.MethodGet,
    Path:    "/users/:id/edit",
    Name:    "user.edit",
    Handler: h.UserEditForm,
})

// URL reversal
path := route.Reverse(e.Routes(), "user.edit", userID)
// => "/users/550e8400-e29b-41d4-a716-446655440000/edit"

// Safe reversal for sitemap generation
if path, ok := route.SafeReverse(e.Routes(), "home.show"); ok {
    // use path
}
```

---

## validate

Thin wrapper around [validator] providing a reusable `Validator` type. Construct a single instance at startup and pass it to handlers and form helpers.

### Types

**`Validator`** -- wraps `*validator.Validate`.

### Functions

**`New() *Validator`** -- creates a ready-to-use validator.

### Methods

**`Struct(s any) error`** -- validates a struct using its `validate` struct tags. Returns `validator.ValidationErrors` on failure.

**`Var(field any, tag string) error`** -- validates a single variable against a tag expression.

**`Validate() *validator.Validate`** -- returns the underlying [validator] instance for callers that need direct access (e.g., form submission helpers).

```go
v := validate.New()

type CreateUserInput struct {
    Email string `validate:"required,email"`
    Name  string `validate:"required,min=2,max=100"`
}

input := CreateUserInput{Email: "alice@example.com", Name: "Alice"}
if err := v.Struct(input); err != nil {
    // err is validator.ValidationErrors
}
```

---

## Bootstrap Example

The following example shows all sub-packages wired together in a typical application startup sequence.

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/go-sum/server"
    "github.com/go-sum/server/config"
    "github.com/go-sum/server/database"
    "github.com/go-sum/server/logging"
    "github.com/go-sum/server/validate"
)

func main() {
    ctx := context.Background()

    // 1. Load layered configuration
    cfg, err := config.Load(func(c *AppConfig) config.Options {
        return config.Options{
            Files: []config.ConfigFile{
                {Filepath: "config/app.yaml"},
                {Filepath: "config/site.yaml"},
            },
            EnvKey: os.Getenv("APP_ENV"),
        }
    })
    if err != nil {
        slog.Error("config load failed", "error", err)
        os.Exit(1)
    }

    // 2. Initialize structured logger
    logging.Init(logging.Config{
        Development: cfg.IsDevelopment(),
        Level:       cfg.App.Log.Level,
    })

    // 3. Connect database pool
    pool, err := database.Connect(ctx, cfg.DSN())
    if err != nil {
        slog.Error("database connect failed", "error", err)
        os.Exit(1)
    }
    defer database.Close(pool)

    // 4. Create Echo instance
    e := server.New()

    // 5. Create validator
    v := validate.New()
    _ = v // pass to handlers

    // 6. Register middleware and routes
    // ...

    // 7. Start with graceful shutdown
    if err := server.Start(e, server.Config{
        Host:            cfg.App.Host,
        Port:            cfg.App.Port,
        GracefulTimeout: cfg.App.GracefulTimeout,
    }); err != nil {
        slog.Error("server exited with error", "error", err)
        os.Exit(1)
    }
}
```

---

## Leaf-Node Rule

Every package in this module imports only the Go standard library and external modules. There are no imports from application-specific `internal/` packages and no cross-imports between sibling `pkg/` packages. This means the entire `github.com/go-sum/server` module can be vendored into any [Echo] project without pulling in application-specific code.

[Echo]: https://echo.labstack.com/
[pgx]: https://github.com/jackc/pgx
[koanf]: https://github.com/knadh/koanf
[validator]: https://github.com/go-playground/validator
