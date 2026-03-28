---
title: HTTP security primitives
description: Reusable HTTP security primitives for Go web applications.
weight: 20
---

# HTTP security primitives

`github.com/go-sum/security` is a collection of reusable HTTP security primitives for Go web applications. Core primitives (`token`, `origin`, `fetchmeta`, `httpsec`, `headers`) are framework-agnostic. [Echo] middleware wrappers live in `csrf`, `ratelimit`, and `middleware`.

## Dependencies

| Dependency | Version |
|------------|---------|
| [Echo] | v5.0 |

## Features

- HMAC-SHA256 signed, stateless, time-limited tokens with scope isolation
- CSRF protection middleware for [Echo] with automatic token issuance and verification
- Per-IP token bucket rate limiting with typed errors and custom identifier support
- Origin and Referer header validation with hostname normalization (RFC compliant)
- W3C Fetch Metadata header validation (`Sec-Fetch-Site`, `Sec-Fetch-Mode`, `Sec-Fetch-Dest`)
- Composed cross-origin guard middleware combining origin and fetch metadata checks
- CSP directive source injection utility for Content-Security-Policy header management

## Sub-packages at a Glance

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `csrf` | `github.com/go-sum/security/csrf` | [Echo] CSRF protection middleware using `token` |
| `fetchmeta` | `github.com/go-sum/security/fetchmeta` | W3C Fetch Metadata header validation |
| `headers` | `github.com/go-sum/security/headers` | CSP directive source injection |
| `httpsec` | `github.com/go-sum/security/httpsec` | HTTP method safety classification (RFC 9110) |
| `middleware` | `github.com/go-sum/security/middleware` | Cross-origin guard (origin + fetchmeta) |
| `origin` | `github.com/go-sum/security/origin` | Origin/Referer header validation |
| `ratelimit` | `github.com/go-sum/security/ratelimit` | [Echo] per-IP token bucket rate limiting |
| `token` | `github.com/go-sum/security/token` | HMAC-SHA256 signed, stateless, time-limited tokens |

---

## csrf

[Echo] middleware. Issues HMAC-signed CSRF tokens on safe requests; verifies on unsafe requests.

### Types

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `[]byte` | HMAC signing key (32+ bytes recommended) |
| `TokenTTL` | `time.Duration` | Token lifetime (default: 1 hour) |
| `ContextKey` | `string` | Echo context key where token is stored |
| `HeaderName` | `string` | Request header to read token (checked first) |
| `FormField` | `string` | Form field to read token (fallback) |

### Functions

**`Middleware(cfg Config) echo.MiddlewareFunc`** -- returns an [Echo] middleware that manages CSRF tokens.

### Behavior

- **Safe methods (GET, HEAD, OPTIONS, TRACE):** issues a fresh token and stores it in `c.Set(cfg.ContextKey, tok)`.
- **Unsafe methods (POST, PUT, PATCH, DELETE):** reads the token from `HeaderName` then `FormField`; returns a typed 403 error on failure; issues a fresh token on success (for re-render on validation errors).

The error type implements `StatusCode() int` (403) and `PublicMessage() string`.

```go
e.Use(csrf.Middleware(csrf.Config{
    Key:        []byte(cfg.Security.CSRF.Key),
    TokenTTL:   1 * time.Hour,
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

- `StatusCode() int` -- returns `Status`
- `PublicMessage() string` -- returns `Message`
- `Unwrap() error` -- returns `Cause`

### Functions

**`CrossOriginGuard(originPolicy origin.Policy, fetchPolicy fetchmeta.Policy) echo.MiddlewareFunc`** -- returns an [Echo] middleware that validates unsafe requests against both policies.

### Behavior

- Safe methods (GET, HEAD, OPTIONS, TRACE) pass through without checks.
- Unsafe methods are validated against both policies; either failure returns `*Error{Status: 403}`.

```go
e.Use(middleware.CrossOriginGuard(
    origin.Policy{
        Enabled:         true,
        CanonicalOrigin: "https://example.com",
        RequireHeader:   true,
    },
    fetchmeta.Policy{
        Enabled:             true,
        AllowedSites:        []string{"same-origin", "same-site"},
        AllowedModes:        []string{"cors", "navigate", "same-origin"},
        FallbackWhenMissing: true,
    },
))
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

[Echo] middleware. Per-identifier token bucket using [Echo]'s built-in in-memory store.

### Types

**`Config`** -- middleware configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Rate` | `float64` | Requests per second (refill rate) |
| `Burst` | `int` | Max burst (default: `ceil(Rate)`) |
| `IdentifierExtractor` | `func(c *echo.Context) (string, error)` | Rate-limit key; default: remote IP |

### Functions

**`Middleware(cfg Config) echo.MiddlewareFunc`** -- returns an [Echo] middleware that enforces per-identifier rate limits.

### Behavior

- Default extractor uses `c.RealIP()`, handling `X-Forwarded-For:IP:port` via `net.SplitHostPort`.
- Requests exceeding the limit return a typed 429 error.
- Each call to `Middleware()` creates an independent in-memory store -- different policies maintain separate per-IP buckets.
- The error type implements `StatusCode() int` (429 or 500) and `PublicMessage() string`.

```go
mw := ratelimit.Middleware(ratelimit.Config{
    Rate:  5.0,
    Burst: 10,
})
signinGroup.Use(mw)
```

---

## token

HMAC-SHA256 signed, stateless tokens. No server-side state required.

**Wire format** (48 bytes, base64url-encoded, no padding):

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
| `origin` | CORS defense | Origin/Referer validation with hostname normalization |
| `fetchmeta` | Browser metadata | W3C `Sec-Fetch-*` header validation |
| `ratelimit` | Brute-force defense | Per-IP token bucket (or custom identifier) |
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

### Rate Limiting by Named Policy

```go
// Each policy maintains an independent per-IP bucket
signin := g.Group("/signin")
signin.Use(server.RateLimitMiddleware(cfg, "signin"))  // 5 req/sec
signin.POST("", h.SignIn)

api := g.Group("/api")
api.Use(server.RateLimitMiddleware(cfg, "api"))        // 100 req/sec
```

### Custom Rate Limit Identifier (User ID)

```go
mw := ratelimit.Middleware(ratelimit.Config{
    Rate:  100,
    Burst: 200,
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
