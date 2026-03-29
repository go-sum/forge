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
- Declarative `Cache-Control` and `Vary` header middleware for route groups
- HTTP method override middleware for HTML forms (`PUT`, `PATCH`, `DELETE`)
- ETag-based conditional response middleware with automatic `304 Not Modified`
- Typed HTTP header parsing and construction for `Accept`, `Accept-Language`, `Cache-Control`, and `Vary`
- ETag generation, `Last-Modified` helpers, and conditional-request checking
- Named route registration and type-safe URL reversal utilities
- Struct and field validation via [validator]

## Sub-packages at a Glance

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `server` | `github.com/go-sum/server` | [Echo] instance factory and graceful-shutdown lifecycle |
| `apperr` | `github.com/go-sum/server/apperr` | Typed application errors with HTTP status codes and safe messages |
| `cache` | `github.com/go-sum/server/cache` | ETag generation and HTTP conditional-request helpers |
| `config` | `github.com/go-sum/server/config` | Generic layered YAML config loader with `${VAR}` expansion |
| `database` | `github.com/go-sum/server/database` | PostgreSQL `pgxpool` connection and helpers |
| `headers` | `github.com/go-sum/server/headers` | Typed HTTP header parsing for `Accept`, `Accept-Language`, `Cache-Control`, `Vary` |
| `logging` | `github.com/go-sum/server/logging` | `slog`-based structured logging with dev/prod modes |
| `middleware` | `github.com/go-sum/server/middleware` | Reusable Echo middleware (static asset caching, cache headers) |
| `middleware/etag` | `github.com/go-sum/server/middleware/etag` | ETag conditional-response middleware |
| `middleware/override` | `github.com/go-sum/server/middleware/override` | HTTP method override middleware for HTML forms |
| `route` | `github.com/go-sum/server/route` | Named route registration and URL reversal |
| `validate` | `github.com/go-sum/server/validate` | Struct and field validation wrapper |

---

## server (root package)

Creates an [Echo] instance and manages server lifecycle with graceful shutdown.

### Types

**`Config`** -- server startup configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Host` | `string` | Server hostname or IP address |
| `Port` | `string` | Server port |
| `GracefulTimeout` | `time.Duration` | Maximum time to wait for in-flight requests during shutdown |
| `BeforeServeFunc` | `func(*http.Server) error` | Called with the underlying `*http.Server` before it starts accepting connections. Use to set `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`, etc. Optional. |
| `ListenerAddrFunc` | `func(net.Addr)` | Called with the resolved `net.Addr` after the listener is bound. Use to discover the actual port when `Port` is `"0"`. Optional. |

### Functions

**`NewWithConfig(cfg echo.Config) *echo.Echo`** -- creates an [Echo] instance with construction-time configuration. Pass `echo.Config{HTTPErrorHandler: ...}` to install the error handler at construction time rather than post-construction via field assignment.

**`Start(e *echo.Echo, cfg Config) error`** -- begins listening on `Host:Port` and blocks until `SIGINT` or `SIGTERM` is received. Performs graceful shutdown within `GracefulTimeout`. Returns an error (prefixed with `"server:"`) rather than calling `os.Exit`, so deferred cleanup in `main` (such as closing database pools) runs normally. Logs startup and shutdown events via `slog`.

```go
e := server.NewWithConfig(echo.Config{})
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

## cache

ETag generation and HTTP conditional-request helpers for use with response caching middleware. All functions depend only on the standard library.

### Functions

**`WeakETag(content []byte) string`** -- returns a weak ETag of the form `W/"<len>-<crc32hex>"`. Uses the content length and IEEE CRC32 checksum -- fast and non-cryptographic. Suitable for fragment caching where exact byte-identity is not required.

**`StrongETag(content []byte) string`** -- returns a strong ETag of the form `"<sha256hex>"`. Uses a full SHA-256 hash -- cryptographically strong and suitable when exact byte-identity must be guaranteed.

**`SetETag(h http.Header, tag string)`** -- writes the `ETag` response header. `tag` must already be a correctly formatted ETag value (e.g. the return value of `WeakETag` or `StrongETag`).

**`SetLastModified(h http.Header, t time.Time)`** -- writes the `Last-Modified` response header using the HTTP date format defined by RFC 7231.

**`CheckIfNoneMatch(r *http.Request, etag string) bool`** -- reports whether the request's `If-None-Match` header matches `etag`, indicating that a `304 Not Modified` response is appropriate. Handles the wildcard value `"*"` and comma-separated lists of quoted ETag strings.

**`CheckIfModifiedSince(r *http.Request, t time.Time) bool`** -- reports whether the content has NOT been modified since the time in the request's `If-Modified-Since` header, indicating that a `304 Not Modified` response is appropriate. Returns `false` (treat as modified) if the header is absent or unparseable.

```go
import "github.com/go-sum/server/cache"

// Generate ETags
body := []byte("<div>Hello</div>")
weak := cache.WeakETag(body)     // W/"16-a1b2c3d4"
strong := cache.StrongETag(body) // "e3b0c44298fc1c14..."

// Set response headers
cache.SetETag(w.Header(), weak)
cache.SetLastModified(w.Header(), time.Now())

// Check conditional request headers
if cache.CheckIfNoneMatch(r, weak) {
    w.WriteHeader(http.StatusNotModified)
    return
}
```

---

## config

Generic, layered YAML configuration loader built on [koanf]. Supports environment-specific overlays, `${VAR}` expansion, and struct validation.

### Types

**`ConfigFile`** -- a single configuration file entry.

| Field | Type | Description |
|-------|------|-------------|
| `Filepath` | `string` | Path to the YAML file |
| `Required` | `bool` | When `true`, a missing file fails startup. Parse, read, and permission errors are always fatal regardless of this flag. |

**`Options`** -- controls how `Load` discovers and merges configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Files` | `[]ConfigFile` | Ordered list of config files. Each file declares whether it is required via the `Required` field. |
| `EnvKey` | `string` | Active environment name (e.g., `"development"`). Triggers an overlay lookup. |

### Functions

**`Load[T any](opts func(*T) Options) (*T, error)`** -- allocates a `*T`, calls the options function to determine files and environment, then loads and validates the configuration.

### Loading Order (last writer wins)

1. `Files[0].Filepath` -- usually the required base config; returns an error if missing when `Required` is `true`
2. `{dir}/{stem}.{EnvKey}.yaml` -- optional environment overlay; silently skipped if absent
3. `Files[1:]` -- loaded in order; missing optional files (where `Required` is `false`) are silently skipped
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
            {Filepath: "config/app.yaml", Required: true},
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

**`IsUniqueViolation(err error) bool`** -- returns `true` if the error is a PostgreSQL unique constraint violation (SQLSTATE `23505`). Use this in repository code to map database errors to domain errors without duplicating the error code string.

```go
pool, err := database.Connect(ctx, "postgres://user:pass@localhost:5432/mydb?pool_max_conns=20")
if err != nil {
    return fmt.Errorf("startup: %w", err)
}
defer pool.Close()

// In a repository:
_, err = q.InsertUser(ctx, params)
if database.IsUniqueViolation(err) {
    return model.ErrEmailTaken
}
```

---

## headers

Typed parsing, construction, and round-trip serialisation for common HTTP headers: `Accept`, `Accept-Language`, `Cache-Control`, and `Vary`. All parsers accept raw header string values as returned by `http.Header.Get` and return typed values with structured accessor methods and `fmt.Stringer` implementations for round-trip serialisation. This package depends only on the Go standard library.

### Accept-Language

**`LanguageItem`** -- a single parsed entry from an `Accept-Language` header value.

| Field | Type | Description |
|-------|------|-------------|
| `Tag` | `string` | Language tag, e.g. `"en-US"`, `"fr"`, `"*"` |
| `Quality` | `float64` | Quality value in `[0.0, 1.0]`; defaults to `1.0` |

**`AcceptLanguage`** (`[]LanguageItem`) -- a quality-weighted list of language tags, sorted by `Quality` descending.

**`ParseAcceptLanguage(header string) AcceptLanguage`** -- parses a raw `Accept-Language` header value and returns a quality-sorted `AcceptLanguage` slice. Items with `q=0` are dropped. An empty or blank header returns an empty (non-nil) slice.

Methods:

- `Preferred(candidates []string) string` -- returns the best matching candidate for the client's language preferences via exact case-insensitive match or subtag prefix match. Returns `""` if no match.
- `String() string` -- serialises back to a header value. Items with `q==1.0` omit the `q` parameter.

```go
import "github.com/go-sum/server/headers"

al := headers.ParseAcceptLanguage("en-US,fr;q=0.8,de;q=0.5")
best := al.Preferred([]string{"fr", "en"}) // "en" (subtag prefix match for en-US)
```

### Accept

**`ContentItem`** -- a single parsed entry from an `Accept` header value.

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `string` | Media type, e.g. `"text/html"`, `"application/*"`, `"*/*"` |
| `Quality` | `float64` | Quality value in `[0.0, 1.0]`; defaults to `1.0` |
| `Params` | `map[string]string` | Non-`q` extension parameters |

**`Accept`** (`[]ContentItem`) -- a quality-weighted list of media types, sorted by `Quality` descending with more-specific types ranked higher at equal quality.

**`ParseAccept(header string) Accept`** -- parses a raw `Accept` header value and returns a sorted `Accept` slice. Items with `q=0` are dropped. An empty or blank header returns an empty (non-nil) slice.

Methods:

- `Preferred(candidates []string) string` -- returns the best matching candidate for the client's `Accept` preferences. Supports exact matches, type wildcards (`text/*`), and the global wildcard (`*/*`). Returns `""` if no match.
- `String() string` -- serialises back to a header value. Items with `q==1.0` omit the `q` parameter.

```go
import "github.com/go-sum/server/headers"

accept := headers.ParseAccept("text/html, application/json;q=0.9")
best := accept.Preferred([]string{"application/json", "text/html"}) // "text/html"
```

### Cache-Control

**`CacheControl`** -- holds the directives parsed from a `Cache-Control` header value. Directive names are stored lower-cased. Flag directives (no value) are stored with an empty string value.

**`ParseCacheControl(header string) CacheControl`** -- parses a raw `Cache-Control` header value. Directive names are lower-cased; the first occurrence of a duplicate directive wins. An empty or blank header returns a zero `CacheControl`.

Methods:

| Method | Return | Description |
|--------|--------|-------------|
| `MaxAge()` | `(seconds int, ok bool)` | Returns the `max-age` directive value |
| `SMaxAge()` | `(seconds int, ok bool)` | Returns the `s-maxage` directive value |
| `NoStore()` | `bool` | Reports whether `no-store` is set |
| `NoCache()` | `bool` | Reports whether `no-cache` is set |
| `Private()` | `bool` | Reports whether `private` is set |
| `Public()` | `bool` | Reports whether `public` is set |
| `MustRevalidate()` | `bool` | Reports whether `must-revalidate` is set |
| `Immutable()` | `bool` | Reports whether `immutable` is set |
| `Has(directive string)` | `bool` | Reports whether the named directive (case-insensitive) is present |
| `String()` | `string` | Serialises directives back to a header value in sorted order |

```go
import "github.com/go-sum/server/headers"

cc := headers.ParseCacheControl("public, max-age=3600, immutable")
if maxAge, ok := cc.MaxAge(); ok {
    fmt.Println(maxAge) // 3600
}
fmt.Println(cc.Immutable()) // true
fmt.Println(cc.NoStore())   // false
```

**`Builder`** -- builds a `Cache-Control` header value programmatically. Directives are output in the order they were added. Calling the same directive setter more than once is a no-op after the first call.

**`NewCacheControl() *Builder`** -- returns a new `Builder`.

Builder methods (all return `*Builder` for chaining): `MaxAge(seconds int)`, `SMaxAge(seconds int)`, `NoStore()`, `NoCache()`, `Private()`, `Public()`, `MustRevalidate()`, `Immutable()`.

`String() string` -- returns the final `Cache-Control` header value.

```go
import "github.com/go-sum/server/headers"

value := headers.NewCacheControl().Public().MaxAge(86400).Immutable().String()
// "public, max-age=86400, immutable"
```

### Vary

**`AppendVary(h http.Header, values ...string)`** -- adds each value to the `Vary` header without creating duplicates. Comparison is case-insensitive. Uses `http.Header.Set` (not `Add`) to avoid the duplicate entries that arise when multiple middleware each call `h.Add("Vary", ...)` independently. If `values` is empty or all values are already present, `h` is unchanged.

```go
import "github.com/go-sum/server/headers"

h := make(http.Header)
headers.AppendVary(h, "Accept", "Accept-Language")
headers.AppendVary(h, "Accept") // no-op — already present
// h.Get("Vary") == "Accept, Accept-Language"
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

**`CacheHeaders(cacheControl string, vary ...string) echo.MiddlewareFunc`** -- sets the `Cache-Control` header and appends `Vary` values before calling the next handler. Use this to apply cache directives to an entire route group declaratively.

```go
import "github.com/go-sum/server/middleware"

// Apply to a route group: cache for 1 hour, vary on Accept
fragmentGroup := e.Group("/fragments")
fragmentGroup.Use(middleware.CacheHeaders("public, max-age=3600", "Accept"))
```

---

## middleware/etag

Conditional-response middleware for [Echo]. Buffers the response body from downstream handlers, computes a weak ETag over the buffered content, and short-circuits with a `304 Not Modified` response when the request's `If-None-Match` header matches. Only GET requests are buffered; all other methods pass through unchanged.

### Types

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Skipper` | `func(c *echo.Context) bool` | Skips the middleware when it returns `true`. Defaults to never skip. |

### Functions

**`Middleware() echo.MiddlewareFunc`** -- returns an ETag middleware with `DefaultConfig`.

**`NewWithConfig(cfg Config) echo.MiddlewareFunc`** -- returns an ETag middleware with the given config. Panics if the config is invalid.

**`(Config) ToMiddleware() (echo.MiddlewareFunc, error)`** -- converts `Config` to an `echo.MiddlewareFunc` or returns an error for invalid configuration.

```go
import "github.com/go-sum/server/middleware/etag"

// Apply to a route group
fragmentGroup := e.Group("/fragments")
fragmentGroup.Use(etag.Middleware())
```

### Behavior

| Condition | Action |
|-----------|--------|
| Non-GET request | Passes through unchanged |
| GET response with status 200 and non-empty body | Computes weak ETag, attaches `ETag` header |
| `If-None-Match` matches computed ETag | Returns `304 Not Modified` with no body |
| Non-200 status or empty body | Flushes response as-is without an ETag |
| Handler returns an error | Error propagated to [Echo] error handler normally |

---

## middleware/override

HTTP method override middleware for [Echo]. HTML forms only emit `GET` and `POST`. This middleware reads a configurable form field (default `"_method"`) from `POST` request bodies and promotes the request method to the specified value before routing. Only `PUT`, `PATCH`, and `DELETE` are permitted override targets -- any other value results in a `400 Bad Request`.

The middleware is designed for `e.Pre()` registration so the promoted method is visible to the router before dispatch.

### Types

**`Config`** -- method override middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Skipper` | `func(c *echo.Context) bool` | Skips the middleware when it returns `true`. Defaults to never skip. |
| `FormField` | `string` | POST body field name that carries the override verb. Defaults to `"_method"`. |

### Functions

**`Middleware() echo.MiddlewareFunc`** -- returns a method override middleware with `DefaultConfig`.

**`NewWithConfig(cfg Config) echo.MiddlewareFunc`** -- returns a method override middleware with the given config. Panics if the config is invalid.

**`(Config) ToMiddleware() (echo.MiddlewareFunc, error)`** -- converts `Config` to an `echo.MiddlewareFunc` or returns an error for invalid configuration.

### Behavior

| Condition | Action |
|-----------|--------|
| Non-POST request | Passes through unchanged |
| POST with empty `_method` field | Passes through unchanged |
| POST with `_method` = `PUT`, `PATCH`, or `DELETE` | Promotes request method |
| POST with `_method` = any other value | Returns `400 Bad Request` |

```go
import "github.com/go-sum/server/middleware/override"

// Register before routing so the promoted method is visible to the router
e.Pre(override.Middleware())
```

```html
<!-- HTML form submitting a DELETE request -->
<form method="POST" action="/users/123">
    <input type="hidden" name="_method" value="DELETE">
    <button type="submit">Delete</button>
</form>
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

Thin wrapper around [validator] providing a reusable `Validator` type. Construct a single instance at startup and pass it to handlers and form helpers. `*Validator` satisfies Echo v5's `Validator` interface via its `Validate` method.

### Types

**`Validator`** -- wraps `*validator.Validate`.

### Functions

**`New() *Validator`** -- creates a ready-to-use validator.

### Methods

**`Validate(i any) error`** -- validates a struct using its `validate` struct tags. Returns `validator.ValidationErrors` on failure. Satisfies Echo v5's `Validator` interface and the `form.StructValidator` interface via structural typing.

**`Var(field any, tag string) error`** -- validates a single variable against a tag expression.

```go
v := validate.New()

type CreateUserInput struct {
    Email string `validate:"required,email"`
    Name  string `validate:"required,min=2,max=100"`
}

input := CreateUserInput{Email: "alice@example.com", Name: "Alice"}
if err := v.Validate(input); err != nil {
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

    "github.com/labstack/echo/v5"

    "github.com/go-sum/server"
    "github.com/go-sum/server/config"
    "github.com/go-sum/server/database"
    "github.com/go-sum/server/logging"
    "github.com/go-sum/server/middleware/override"
    "github.com/go-sum/server/validate"
)

func main() {
    ctx := context.Background()

    // 1. Load layered configuration
    cfg, err := config.Load(func(c *AppConfig) config.Options {
        return config.Options{
            Files: []config.ConfigFile{
                {Filepath: "config/app.yaml", Required: true},
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
    defer pool.Close()

    // 4. Create Echo instance
    e := server.NewWithConfig(echo.Config{})

    // 5. Register pre-routing middleware
    e.Pre(override.Middleware())

    // 6. Create validator
    v := validate.New()
    _ = v // pass to handlers

    // 7. Register middleware and routes
    // ...

    // 8. Start with graceful shutdown
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
