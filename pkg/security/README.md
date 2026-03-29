---
title: HTTP security primitives
description: Reusable HTTP security primitives for Go web applications.
weight: 20
---

# HTTP security primitives

`github.com/go-sum/security` is a collection of reusable HTTP security primitives for Go web applications. Core primitives (`token`, `origin`, `fetchmeta`, `httpsec`, `headers`) are framework-agnostic. [Echo] middleware wrappers live in `cors`, `csrf`, `ratelimit`, and `middleware`.

## Dependencies

| Dependency | Version |
|------------|---------|
| [Echo] | v5.0 |
| [x/time] | v0.14 |

## Features

- HMAC-SHA256 signed, stateless, time-limited tokens with scope isolation
- CSRF protection middleware for [Echo] with automatic token issuance and verification
- CORS middleware for [Echo] with exact, regex, and custom function origin matching
- Per-IP token bucket rate limiting with typed errors, custom identifier support, and lazy stale-entry cleanup
- Origin and Referer header validation with hostname normalization (RFC compliant)
- W3C Fetch Metadata header validation (`Sec-Fetch-Site`, `Sec-Fetch-Mode`, `Sec-Fetch-Dest`)
- Composed cross-origin guard middleware combining origin and fetch metadata checks
- CSP directive source injection utility for Content-Security-Policy header management
- HTTP method safety classification per RFC 9110

## Sub-packages at a Glance

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `cors` | `github.com/go-sum/security/cors` | [Echo] CORS middleware with exact, regex, and custom origin matching |
| `csrf` | `github.com/go-sum/security/csrf` | [Echo] CSRF protection middleware using `token` |
| `fetchmeta` | `github.com/go-sum/security/fetchmeta` | W3C Fetch Metadata header validation |
| `headers` | `github.com/go-sum/security/headers` | CSP directive source injection |
| `httpsec` | `github.com/go-sum/security/httpsec` | HTTP method safety classification (RFC 9110) |
| `middleware` | `github.com/go-sum/security/middleware` | Cross-origin guard (origin + fetchmeta) |
| `origin` | `github.com/go-sum/security/origin` | Origin/Referer header validation |
| `ratelimit` | `github.com/go-sum/security/ratelimit` | [Echo] per-IP token bucket rate limiting |
| `token` | `github.com/go-sum/security/token` | HMAC-SHA256 signed, stateless, time-limited tokens |

---

## cors

[Echo] middleware providing CORS support with three origin-matching strategies: exact string list, compiled regexp set, and custom function. All preflight handling, `Vary: Origin`, `Access-Control-Allow-Credentials`, and `Access-Control-Max-Age` logic is delegated to Echo's built-in `CORSConfig`. This package adds regex origin matching and a unified `Config` surface over the three modes.

### Types

**`OriginMode`** (`int`) -- selects the origin-matching strategy.

| Constant | Value | Description |
|----------|-------|-------------|
| `OriginModeExact` | `0` | Case-insensitive string comparison against `AllowOrigins`. Supports `"*"` wildcard. |
| `OriginModeRegex` | `1` | Matches request Origin against compiled `RegexOrigins` patterns. |
| `OriginModeFunc` | `2` | Delegates to `AllowOriginFunc` for custom validation logic. |

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Skipper` | `func(c *echo.Context) bool` | Skip middleware when returning `true`. Defaults to never skip. |
| `Mode` | `OriginMode` | Origin-matching strategy. See `OriginMode` constants. |
| `AllowOrigins` | `[]string` | Allowed origins for `OriginModeExact`. Supports `"*"` for wildcard (cannot combine with `AllowCredentials`). |
| `RegexOrigins` | `[]string` | Regexp patterns for `OriginModeRegex`. Compiled once at middleware creation time. |
| `AllowOriginFunc` | `func(c *echo.Context, origin string) (string, bool, error)` | Custom origin validator for `OriginModeFunc`. Returns allowed origin, whether permitted, and any error. |
| `AllowMethods` | `[]string` | HTTP methods permitted cross-origin. Defaults to GET, HEAD, PUT, PATCH, POST, DELETE. |
| `AllowHeaders` | `[]string` | Request headers permitted cross-origin. Empty echoes back the browser's requested headers on preflight. |
| `AllowCredentials` | `bool` | Permits cookies and authorization headers cross-origin. Cannot combine with `AllowOrigins = ["*"]`. |
| `ExposeHeaders` | `[]string` | Response headers that browsers are permitted to read. |
| `MaxAge` | `int` | Preflight cache duration in seconds. `0` omits the header; negative sends `"0"`. |

### Functions

**`Middleware(cfg Config) echo.MiddlewareFunc`** -- returns an [Echo] middleware configured with `cfg`. Panics if `cfg` fails validation. Use `ToMiddleware` for non-panicking construction.

### Methods

**`(cfg Config) ToMiddleware() (echo.MiddlewareFunc, error)`** -- converts `Config` to an `echo.MiddlewareFunc`. Returns an error for invalid configuration (missing required fields, invalid regex patterns, unknown `OriginMode`). For `OriginModeRegex`, all patterns are compiled here; an invalid pattern causes an immediate error rather than a per-request panic.

```go
// Exact origin matching
e.Use(cors.Middleware(cors.Config{
    Mode:             cors.OriginModeExact,
    AllowOrigins:     []string{"https://example.com", "https://admin.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           3600,
}))

// Regex origin matching
e.Use(cors.Middleware(cors.Config{
    Mode:         cors.OriginModeRegex,
    RegexOrigins: []string{`^https://.*\.example\.com$`},
    MaxAge:       3600,
}))

// Custom function origin matching
e.Use(cors.Middleware(cors.Config{
    Mode: cors.OriginModeFunc,
    AllowOriginFunc: func(c *echo.Context, origin string) (string, bool, error) {
        // custom logic
        return origin, true, nil
    },
}))
```

---

## csrf

[Echo] middleware. Issues HMAC-signed CSRF tokens on safe requests; verifies on unsafe requests.

### Types

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `[]byte` | HMAC signing key. Must be at least 32 bytes. |
| `TokenTTL` | `int` | Token lifetime in seconds. Defaults to 3600. |
| `ContextKey` | `string` | Echo context key where token is stored. Defaults to `"csrf"`. |
| `HeaderName` | `string` | Request header to read token (checked first). Defaults to `"X-CSRF-Token"`. |
| `FormField` | `string` | Form field to read token (fallback). Defaults to `"_csrf"`. |
| `Skipper` | `func(c *echo.Context) bool` | Skip middleware when returning `true`. Defaults to never skip. |
| `TokenExtractor` | `func(c *echo.Context) (string, error)` | Custom token lookup. When nil, reads `HeaderName` then `FormField`. |

### Functions

**`Middleware(cfg Config) echo.MiddlewareFunc`** -- returns an [Echo] middleware that manages CSRF tokens. Panics on invalid configuration (e.g. key shorter than 32 bytes). Use `ToMiddleware` for non-panicking construction.

### Methods

**`(cfg Config) ToMiddleware() (echo.MiddlewareFunc, error)`** -- converts `Config` to an `echo.MiddlewareFunc`. Returns an error for invalid configuration (key too short). Applies defaults for empty fields.

### Behavior

- **Safe methods (GET, HEAD, OPTIONS, TRACE):** issues a fresh token and stores it in `c.Set(cfg.ContextKey, tok)`.
- **Unsafe methods (POST, PUT, PATCH, DELETE):** reads the token from `HeaderName` then `FormField`; returns a typed 403 error on failure; issues a fresh token on success (for re-render on validation errors).

The error type implements `StatusCode() int` (403) and `PublicMessage() string`.

```go
e.Use(csrf.Middleware(csrf.Config{
    Key:        []byte(cfg.Security.CSRF.Key),
    TokenTTL:   3600,
    ContextKey: "csrf_token",
    HeaderName: "X-CSRF-Token",
    FormField:  "_csrf",
}))
```

---

## fetchmeta

Framework-agnostic W3C Fetch Metadata header validation (`Sec-Fetch-Site`, `Sec-Fetch-Mode`, `Sec-Fetch-Dest`).

### Types

**`Policy`** -- validation configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Enabled` | `bool` | Whether fetch metadata validation is active |
| `AllowedSites` | `[]string` | e.g. `["same-origin", "same-site"]` |
| `AllowedModes` | `[]string` | e.g. `["cors", "navigate", "same-origin"]` |
| `AllowedDestinations` | `[]string` | e.g. `["iframe", "empty"]` |
| `FallbackWhenMissing` | `bool` | Allow requests from legacy browsers without headers |
| `RejectCrossSiteNavigate` | `bool` | Reject `site=="cross-site"` && `mode=="navigate"` |

**`Result`** -- validation outcome.

| Field | Type | Description |
|-------|------|-------------|
| `Valid` | `bool` | Whether the request passed validation |
| `Reason` | `string` | Failure reason (empty on success) |
| `HeadersMissing` | `bool` | All three `Sec-Fetch-*` headers absent |

### Functions

**`Validate(r *http.Request, policy Policy) Result`** -- validates the request against the policy.

### Behavior

- When all three headers are absent and `FallbackWhenMissing` is `true`, returns `Valid: true` with `HeadersMissing: true` (compatibility for older browsers).
- `RejectCrossSiteNavigate` blocks top-level form POSTs from attacker-controlled pages.

```go
result := fetchmeta.Validate(r, fetchmeta.Policy{
    Enabled:                 true,
    AllowedSites:            []string{"same-origin", "same-site"},
    AllowedModes:            []string{"cors", "navigate", "same-origin"},
    FallbackWhenMissing:     true,
    RejectCrossSiteNavigate: true,
})
if !result.Valid {
    slog.Warn("fetch metadata check failed", "reason", result.Reason)
}
```

---

## headers

CSP directive source injection utility. Use at startup to pre-process CSP strings before passing to HTTP middleware.

### Functions

**`InjectDirectiveSources(csp, directive string, sources []string) string`** -- prepends `sources` into the named directive (space-separated). Returns the original CSP unchanged if the directive is not found or `sources` is empty. Filters blank and whitespace-only source strings.

```go
csp := "default-src 'self'; style-src 'self'; font-src 'self'"
out := headers.InjectDirectiveSources(csp, "style-src",
    []string{"https://fonts.googleapis.com", "'sha256-abc123'"})
// "default-src 'self'; style-src https://fonts.googleapis.com 'sha256-abc123' 'self'; font-src 'self'"
```

---

## httpsec

HTTP method safety classification per RFC 9110. Used internally by `csrf` and `middleware` to decide whether checks apply.

### Functions

**`IsSafeMethod(method string) bool`** -- returns `true` for GET, HEAD, OPTIONS, and TRACE.

**`IsUnsafeMethod(method string) bool`** -- inverse of `IsSafeMethod`.

```go
if httpsec.IsSafeMethod(r.Method) {
    // read-only request -- skip CSRF check
}
```

---

## middleware

Composed [Echo] middleware combining `origin` and `fetchmeta` for a complete cross-origin guard.

### Types

**`Error`** -- typed error returned on validation failure.

| Field | Type | Description |
|-------|------|-------------|
| `Status` | `int` | HTTP status code (403) |
| `Message` | `string` | Safe public message |
| `Cause` | `error` | Underlying validation error |

Methods:

- `Error() string` -- returns `Cause.Error()` if present, otherwise `Message`
- `StatusCode() int` -- returns `Status`
- `PublicMessage() string` -- returns `Message`
- `Unwrap() error` -- returns `Cause`

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Skipper` | `func(c *echo.Context) bool` | Skip middleware when returning `true`. Defaults to never skip. |
| `OriginPolicy` | `origin.Policy` | Origin header validation configuration. |
| `FetchPolicy` | `fetchmeta.Policy` | Fetch Metadata validation configuration. |

### Functions

**`Middleware(cfg Config) echo.MiddlewareFunc`** -- returns an [Echo] middleware from `cfg`. Panics on invalid configuration (both policies disabled). Use `ToMiddleware` for non-panicking construction.

**`CrossOriginGuard(originPolicy origin.Policy, fetchPolicy fetchmeta.Policy) echo.MiddlewareFunc`** -- convenience wrapper. Deprecated: use `Middleware(Config{...})` instead.

### Methods

**`(cfg Config) ToMiddleware() (echo.MiddlewareFunc, error)`** -- converts `Config` to an `echo.MiddlewareFunc`. Returns an error when both `OriginPolicy` and `FetchPolicy` are disabled.

### Behavior

- Safe methods (GET, HEAD, OPTIONS, TRACE) pass through without checks.
- Unsafe methods are validated against enabled policies; either failure returns `*Error{Status: 403}`.
- When `Skipper` returns `true`, the request passes through without checks regardless of method.

```go
e.Use(middleware.Middleware(middleware.Config{
    OriginPolicy: origin.Policy{
        Enabled:         true,
        CanonicalOrigin: "https://example.com",
        RequireHeader:   true,
    },
    FetchPolicy: fetchmeta.Policy{
        Enabled:             true,
        AllowedSites:        []string{"same-origin", "same-site"},
        AllowedModes:        []string{"cors", "navigate", "same-origin"},
        FallbackWhenMissing: true,
    },
}))
```

---

## origin

Framework-agnostic Origin/Referer header validation for CSRF defense.

### Types

**`Policy`** -- validation configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Enabled` | `bool` | Whether origin validation is active |
| `CanonicalOrigin` | `string` | Expected origin, e.g. `"https://example.com"` |
| `RequireHeader` | `bool` | Reject when both Origin and Referer are absent |
| `AllowedOrigins` | `[]string` | Additional accepted origins (e.g. subdomains) |

**`Result`** -- validation outcome.

| Field | Type | Description |
|-------|------|-------------|
| `Valid` | `bool` | Whether the request passed validation |
| `Reason` | `string` | Failure reason (empty on success) |
| `Source` | `string` | `"Origin"` or `"Referer"` -- which header was used |
| `HeadersMissing` | `bool` | Both Origin and Referer absent |

### Functions

**`Validate(r *http.Request, policy Policy) Result`** -- validates the request against the policy.

### Normalization Rules

Lowercases hostname; removes default ports (`:80` for http, `:443` for https); keeps non-default ports. `Origin` header is checked first; falls back to `Referer`.

```go
result := origin.Validate(r, origin.Policy{
    Enabled:         true,
    CanonicalOrigin: "https://example.com",
    RequireHeader:   true,
})
if !result.Valid {
    slog.Warn("origin check failed", "reason", result.Reason)
}
```

---

## ratelimit

[Echo] middleware. Per-identifier token bucket rate limiting with an in-memory store and lazy stale-entry cleanup.

### Types

**`Skipper`** (`func(c *echo.Context) bool`) -- defines a function to skip middleware. Returning `true` skips processing.

**`BeforeFunc`** (`func(c *echo.Context)`) -- defines a function executed just before the middleware check.

**`Store`** -- interface for custom rate limit backends.

```go
type Store interface {
    Allow(identifier string) (bool, error)
}
```

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Skipper` | `Skipper` | Skip middleware when returning `true`. Defaults to never skip. |
| `BeforeFunc` | `BeforeFunc` | Function executed before the rate limit check. |
| `IdentifierExtractor` | `func(c *echo.Context) (string, error)` | Rate-limit key extractor. Defaults to remote IP via `c.RealIP()`. |
| `Store` | `Store` | Rate limit backend. Required -- use `NewMemoryStore` or `NewMemoryStoreWithConfig` to build one. |
| `DenyHandler` | `func(c *echo.Context, identifier string, err error) error` | Called when the store denies a request. Defaults to returning a typed 429 error. |
| `ErrorHandler` | `func(c *echo.Context, err error) error` | Called when `IdentifierExtractor` returns an error. Defaults to returning a typed 500 error. |

**`MemoryStoreConfig`** -- construction parameters for `MemoryStore`.

| Field | Type | Description |
|-------|------|-------------|
| `Rate` | `float64` | Requests per second (token bucket refill rate) |
| `Burst` | `int` | Maximum burst size. `0` defaults to `ceil(Rate)` (min 1). |
| `ExpiresIn` | `time.Duration` | Duration before stale visitors are eligible for cleanup. `0` defaults to 3 minutes. |

**`MemoryStore`** -- in-process token-bucket store keyed by identifier. Stale entries are lazily swept when the time since the last cleanup exceeds `ExpiresIn`, preventing unbounded memory growth.

### Functions

**`Middleware(cfg Config) echo.MiddlewareFunc`** -- returns an [Echo] middleware that enforces per-identifier rate limits. Panics on invalid configuration (e.g. nil `Store`). Use `ToMiddleware` for non-panicking construction.

**`NewMemoryStore(rate float64) *MemoryStore`** -- returns a `MemoryStore` with the given rate (req/s). Burst defaults to `ceil(rate)` (min 1); `ExpiresIn` defaults to 3 minutes.

**`NewMemoryStoreWithConfig(cfg MemoryStoreConfig) *MemoryStore`** -- returns a `MemoryStore` configured with the provided `MemoryStoreConfig`.

### Methods

**`(cfg Config) ToMiddleware() (echo.MiddlewareFunc, error)`** -- converts `Config` to an `echo.MiddlewareFunc`. Returns an error for invalid configuration rather than panicking.

**`(s *MemoryStore) Allow(identifier string) (bool, error)`** -- implements `Store`. Returns `true` if the identifier is within its rate limit.

### Behavior

- Default extractor uses `c.RealIP()`, handling `X-Forwarded-For:IP:port` via `net.SplitHostPort`.
- Requests exceeding the limit return a typed 429 error.
- Each call to `Middleware()` creates an independent middleware instance -- different policies maintain separate per-IP buckets when using separate stores.
- The error type implements `StatusCode() int` (429 or 500) and `PublicMessage() string`.

```go
store := ratelimit.NewMemoryStoreWithConfig(ratelimit.MemoryStoreConfig{
    Rate:      5.0,
    Burst:     10,
    ExpiresIn: 5 * time.Minute,
})

mw := ratelimit.Middleware(ratelimit.Config{
    Store: store,
})
signinGroup.Use(mw)
```

---

## token

HMAC-SHA256 signed, stateless tokens. No server-side state required.

**Wire format** (64 bytes, base64url-encoded, no padding):

- Bytes 0--15: random nonce
- Bytes 16--23: `iat` (unix seconds, big-endian int64)
- Bytes 24--31: `exp` (unix seconds, big-endian int64)
- Bytes 32--63: HMAC-SHA256(key, scope || `\x00` || nonce || iat || exp)

### Errors

```go
var ErrInvalid error  // malformed, tampered, or wrong-scope token
var ErrExpired error  // token past expiry
```

### Functions

**`Issue(key []byte, scope string, ttl time.Duration) (string, error)`** -- generates a new signed token with the given scope and time-to-live. Returns the base64url-encoded token string.

**`Verify(key []byte, scope string, raw string) error`** -- validates a token against the key and scope. Returns `ErrInvalid` for malformed, tampered, or wrong-scope tokens. Returns `ErrExpired` for tokens past expiry. MAC is always verified before expiry is checked to prevent timing leaks.

```go
tok, err := token.Issue(key, "csrf", 1*time.Hour)
if err != nil {
    return fmt.Errorf("token issue: %w", err)
}

err = token.Verify(key, "csrf", tok)
if errors.Is(err, token.ErrExpired) {
    // token was valid but has expired
}
if errors.Is(err, token.ErrInvalid) {
    // token is malformed, tampered, or wrong scope
}
```

### Security Properties

- `scope` is mixed into the HMAC but not stored in the token -- prevents replay across purposes
- Uses `hmac.Equal()` for constant-time comparison
- Nonce generated via `crypto/rand` for each token

---

## Security Properties

| Component | Property | Mechanism |
|-----------|----------|-----------|
| `token` | Integrity | HMAC-SHA256; constant-time `hmac.Equal()` |
| `token` | Nonce uniqueness | `crypto/rand` per token |
| `token` | Scope isolation | Scope mixed into MAC; cross-scope replay fails |
| `token` | Timing safety | MAC verified before expiry check |
| `csrf` | State-change protection | HMAC-signed tokens verified on unsafe methods |
| `csrf` | Token refresh | Fresh token issued on every request |
| `cors` | Cross-origin access control | Exact, regex, or custom origin matching delegated to Echo |
| `cors` | Credential safety | `AllowCredentials` cannot combine with wildcard `"*"` origin |
| `origin` | CORS defense | Origin/Referer validation with hostname normalization |
| `fetchmeta` | Browser metadata | W3C `Sec-Fetch-*` header validation |
| `ratelimit` | Brute-force defense | Per-IP token bucket (or custom identifier) with stale-entry cleanup |
| `middleware` | Composed defense | Origin + Fetch Metadata combined for unsafe requests |

---

## Integration Patterns

### CSRF in HTML Forms

```go
// Handler (safe method -- token stored in context)
func (h *Handler) EditForm(c *echo.Context) error {
    csrfToken := c.Get(h.cfg.Keys.CSRF).(string)
    return render.Component(c, http.StatusOK, view.EditPage(csrfToken))
}
```

```html
<!-- View -->
<form method="POST" action="/user/update">
    <input type="hidden" name="_csrf" value="{{ .CSRFToken }}">
    <input type="text" name="email">
    <button type="submit">Update</button>
</form>
```

### CSRF with HTMX

```html
<div hx-post="/api/items"
     hx-headers='{"X-CSRF-Token": "{{ .CSRFToken }}"}'>
    <input type="text" name="name">
    <button>Add</button>
</div>
```

### CORS for API Routes

```go
apiGroup := e.Group("/api")
apiGroup.Use(cors.Middleware(cors.Config{
    Mode:             cors.OriginModeExact,
    AllowOrigins:     []string{"https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "Authorization", "X-CSRF-Token"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

### Rate Limiting by Named Policy

```go
// Each policy uses an independent store with separate per-IP buckets
signinStore := ratelimit.NewMemoryStoreWithConfig(ratelimit.MemoryStoreConfig{
    Rate:  5.0,
    Burst: 10,
})
signin := g.Group("/signin")
signin.Use(ratelimit.Middleware(ratelimit.Config{Store: signinStore}))
signin.POST("", h.SignIn)

apiStore := ratelimit.NewMemoryStoreWithConfig(ratelimit.MemoryStoreConfig{
    Rate:  100.0,
    Burst: 200,
})
api := g.Group("/api")
api.Use(ratelimit.Middleware(ratelimit.Config{Store: apiStore}))
```

### Custom Rate Limit Identifier (User ID)

```go
store := ratelimit.NewMemoryStore(100)
mw := ratelimit.Middleware(ratelimit.Config{
    Store: store,
    IdentifierExtractor: func(c *echo.Context) (string, error) {
        uid, ok := c.Get("user_id").(string)
        if !ok || uid == "" {
            return "", errors.New("user_id not in context")
        }
        return "user:" + uid, nil
    },
})
```

### Strict vs Lenient Fetch Metadata

```go
// Development / legacy browsers -- allow missing headers
fetchPolicy := fetchmeta.Policy{
    Enabled:             true,
    FallbackWhenMissing: true,
}

// Production -- strict enforcement
fetchPolicy := fetchmeta.Policy{
    Enabled:                 true,
    AllowedSites:            []string{"same-origin", "same-site"},
    AllowedModes:            []string{"cors", "navigate", "same-origin"},
    FallbackWhenMissing:     false,
    RejectCrossSiteNavigate: true,
}
```

---

## Middleware Composition

`CrossOriginGuard` and `CSRF` protect against different threat vectors and are designed to compose.

| Middleware | Question it answers | Mechanism |
|------------|-------------------|-----------|
| `CrossOriginGuard` | Did this request come from a trusted browsing context? | Origin header + `Sec-Fetch-*` metadata |
| `CSRF` | Does the submitter hold a token our server issued? | HMAC-signed token from a legitimate page render |

**Browser-facing mutations** (form submissions, HTMX actions) should apply both in order:

```
CrossOriginGuard → CSRF → Handler
```

`CrossOriginGuard` rejects cross-origin requests at the network level. `CSRF` validates the token even for same-origin requests, guarding against subdomain or browser-extension attacks.

**API endpoints** authenticated via bearer tokens or CORS do not need CSRF tokens. Use CORS middleware + bearer auth instead. Omit `CrossOriginGuard` for API groups that intentionally accept cross-origin requests.

**Error typing:** `CrossOriginGuard` returns `*middleware.Error` (403). `CSRF` returns its own violation type (403). Both implement `StatusCode() int` and `PublicMessage() string`.

---

## Configuration

All security settings map to this YAML structure (consumed by your application's `config.App.Security`):

```yaml
app:
  security:
    externalOrigin: "https://example.com"
    csrf:
      key: "${CSRF_KEY}"         # 32+ byte secret from env var
      headerName: "X-CSRF-Token"
      formField: "_csrf"
    cors:
      mode: exact                # exact | regex | func
      allowOrigins:
        - "https://example.com"
      allowMethods:
        - GET
        - POST
        - PUT
        - DELETE
      allowHeaders:
        - Content-Type
        - Authorization
      allowCredentials: true
      maxAge: 3600
    origin:
      enabled: true
      requireHeader: true
      allowedOrigins: []          # e.g., ["https://admin.example.com"]
    fetchMetadata:
      enabled: true
      allowedSites: ["same-origin", "same-site"]
      allowedModes: ["cors", "navigate", "same-origin"]
      allowedDestinations: []
      fallbackWhenMissing: true
      rejectCrossSiteNavigate: true
    rateLimits:
      signin:                     # policy name used in RateLimitMiddleware
        rate: 5.0
        burst: 10
      api:
        rate: 100.0
        burst: 200
```

---

## Leaf-Node Rule

Every package in this module imports only the Go standard library and external modules. There are no imports from application-specific `internal/` packages and no cross-imports between sibling `pkg/` packages. This means the entire `github.com/go-sum/security` module can be vendored into any [Echo] project without pulling in application-specific code.

[Echo]: https://echo.labstack.com/
[x/time]: https://pkg.go.dev/golang.org/x/time
