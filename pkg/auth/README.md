---
title: Authentication
description: Passwordless authentication module with email-TOTP verification flows and encrypted session management.
weight: 20
---

# auth

`github.com/go-sum/auth` is a self-contained, passwordless authentication module built on email-delivered TOTP codes. It handles signup, signin, and email-change verification flows with encrypted session cookies and self-contained verification tokens. All sub-packages live under a single `go.mod` and follow the **leaf-node rule**: they import only the standard library and external modules -- never application-specific `internal/` code or sibling `pkg/` packages.

## Dependencies

| Dependency | Version |
|------------|---------|
| [gorilla/sessions] | v1.4 |
| [google/uuid] | v1.6 |
| [validator] | v10.30 |

## Features

- Passwordless authentication via email-delivered 6-digit TOTP codes
- Three verification flow types: signup, signin, and email change
- Anti-enumeration protection on signin (no account existence leakage)
- Self-contained encrypted verification tokens for cross-browser link completion
- Signed and encrypted session cookies via [gorilla/sessions]
- Constant-time code comparison to prevent timing attacks
- Configurable TOTP period with sensible 5-minute default
- Provider-agnostic `Service` interface for future auth method extensibility
- Resend support that generates fresh secrets per cycle

## Sub-packages at a Glance

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `auth` (root) | `github.com/go-sum/auth` | Configuration types, `Service` interface, and config validation |
| `model` | `github.com/go-sum/auth/model` | Domain types for users, verification flows, inputs, and error sentinels |
| `repository` | `github.com/go-sum/auth/repository` | Storage port interfaces (`UserReader`, `UserStore`) |
| `service` | `github.com/go-sum/auth/service` | `AuthService` implementation with TOTP generation, token encryption, and flow orchestration |
| `session` | `github.com/go-sum/auth/session` | Encrypted cookie-based session management for user identity and pending flows |

---

## auth (root package)

Defines the provider-agnostic `Service` interface, auth method configuration, and startup validation.

### Types

**`MethodName`** (`string`) -- names an auth method implementation.

| Constant | Value | Description |
|----------|-------|-------------|
| `MethodEmailTOTP` | `"email_totp"` | Email-delivered TOTP verification |

**`Config`** -- top-level auth configuration for the host application.

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `Selected` | `MethodName` | `selected` | Active auth method. Defaults to `"email_totp"` when empty. |
| `Methods` | `MethodsConfig` | `methods` | Method-specific configuration group. |

Methods:

- `SelectedMethod() MethodName` -- returns `Selected` if non-empty, otherwise `MethodEmailTOTP`

**`MethodsConfig`** -- groups method-specific configuration.

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `EmailTOTP` | `EmailTOTPMethodConfig` | `email_totp` | Configuration for the email-TOTP method. |

**`EmailTOTPMethodConfig`** -- configures the email-backed TOTP auth method.

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `Enabled` | `bool` | `enabled` | Whether this method is active. Must be `true` for the method to function. |
| `Issuer` | `string` | `issuer` | Issuer name used in TOTP generation context. |
| `PeriodSeconds` | `int` | `period_seconds` | TOTP validity window in seconds. Defaults to 300 (5 minutes) when zero or negative. |

### Interface

**`Service`** -- provider-agnostic interface for auth operations. `*service.AuthService` satisfies this interface.

```go
type Service interface {
    BeginSignin(context.Context, model.BeginSigninInput, string) (model.PendingFlow, error)
    BeginSignup(context.Context, model.BeginSignupInput, string) (model.PendingFlow, error)
    BeginEmailChange(context.Context, uuid.UUID, model.BeginEmailChangeInput, string) (model.PendingFlow, error)
    ResendPendingFlow(context.Context, model.PendingFlow, string) (model.PendingFlow, error)
    VerifyPendingFlow(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error)
    VerifyToken(context.Context, string, model.VerifyInput) (model.VerifyResult, error)
    VerifyPageState(string) (model.VerifyPageState, error)
}
```

### Functions

**`ValidateConfig(cfg Config) error`** -- checks that `cfg` selects a supported, enabled method. Returns an error suitable for wrapping in a startup panic if the selected method is unsupported or disabled.

```go
if err := auth.ValidateConfig(cfg.Auth); err != nil {
    panic(fmt.Sprintf("auth config: %v", err))
}
```

---

## model

Defines auth domain types: user identity, verification flows, form inputs, and error sentinels.

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `RoleUser` | `"user"` | Standard user access level |
| `RoleAdmin` | `"admin"` | Administrative access level |

### Types

**`FlowPurpose`** (`string`) -- identifies the verification workflow in progress.

| Constant | Value |
|----------|-------|
| `FlowPurposeSignup` | `"signup"` |
| `FlowPurposeSignin` | `"signin"` |
| `FlowPurposeEmailChange` | `"email_change"` |

**`User`** -- the auth module's view of an application user. Contains only the fields required for authentication and session management.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `uuid.UUID` | Unique user identifier |
| `Email` | `string` | User email address |
| `DisplayName` | `string` | User display name |
| `Role` | `string` | Access level (`RoleUser` or `RoleAdmin`) |
| `Verified` | `bool` | Whether the user has completed email verification |
| `CreatedAt` | `time.Time` | Account creation timestamp |
| `UpdatedAt` | `time.Time` | Last update timestamp |

**`BeginSignupInput`** -- validated data for starting a signup verification flow.

| Field | Type | Form Key | Validation |
|-------|------|----------|------------|
| `Email` | `string` | `email` | Required, valid email, max 255 chars |
| `DisplayName` | `string` | `display_name` | Required, 1-255 chars |
| `Role` | `string` | `role` | Optional, one of `user` or `admin` |

**`BeginSigninInput`** -- email address for starting a signin verification flow.

| Field | Type | Form Key | Validation |
|-------|------|----------|------------|
| `Email` | `string` | `email` | Required, valid email, max 255 chars |

**`BeginEmailChangeInput`** -- target email address for a signed-in user.

| Field | Type | Form Key | Validation |
|-------|------|----------|------------|
| `Email` | `string` | `email` | Required, valid email, max 255 chars |

**`VerifyInput`** -- data from the verification form.

| Field | Type | Form Key | Validation |
|-------|------|----------|------------|
| `Code` | `string` | `code` | Required, exactly 6 numeric digits |
| `Token` | `string` | `token` | Optional, present for cross-browser verification |

**`PendingFlow`** -- browser-bound verification state retained between the begin and verify steps. Stored in the session cookie.

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| `Purpose` | `FlowPurpose` | `purpose` | Which flow type is in progress |
| `Email` | `string` | `email` | Target email address |
| `DisplayName` | `string` | `display_name` | Display name (signup only) |
| `Role` | `string` | `role` | User role (signup only) |
| `UserID` | `uuid.UUID` | `user_id` | Authenticated user ID (email change only) |
| `Secret` | `string` | `secret` | Base32-encoded TOTP secret |
| `IssuedAt` | `time.Time` | `issued_at` | When the flow was created |
| `ExpiresAt` | `time.Time` | `expires_at` | When the flow expires |

**`VerificationToken`** -- self-contained payload embedded in emailed verify links. Shares the same fields as `PendingFlow` and is convertible via `model.VerificationToken(flow)`.

**`DeliveryInput`** -- email payload required for sending a verification message.

| Field | Type | Description |
|-------|------|-------------|
| `Purpose` | `FlowPurpose` | Flow type for email template selection |
| `Email` | `string` | Recipient email address |
| `Code` | `string` | 6-digit TOTP code |
| `VerifyURL` | `string` | Complete verification link with embedded token |
| `ExpiresAt` | `time.Time` | Verification expiration time |

**`VerifyResult`** -- describes a successful verification.

| Field | Type | Description |
|-------|------|-------------|
| `Purpose` | `FlowPurpose` | Which flow type completed |
| `User` | `User` | The authenticated or newly created user |

**`VerifyPageState`** -- supplies the verification screen with a purpose and prefilled code.

| Field | Type | Description |
|-------|------|-------------|
| `Purpose` | `FlowPurpose` | Flow type for page rendering |
| `Code` | `string` | Prefilled TOTP code (from email link) |
| `Token` | `string` | Encrypted token for cross-browser submission |
| `Email` | `string` | Target email address for display |
| `CanResend` | `bool` | Whether the user can request a new code |

### Error Sentinels

| Error | Description |
|-------|-------------|
| `ErrUserNotFound` | No user matches the given identifier |
| `ErrEmailTaken` | The email address is already in use |
| `ErrInvalidCredentials` | Signin failed (user not found or not verified) |
| `ErrInvalidVerificationCode` | The submitted TOTP code does not match |
| `ErrVerificationExpired` | The verification flow has exceeded its time window |
| `ErrVerificationMissing` | No verification token or flow data is available |
| `ErrUnsupportedMethod` | The requested auth method is not supported or enabled |

---

## repository

Defines the storage port interfaces that the auth package requires. The host application provides concrete implementations backed by its database layer.

### Interfaces

**`UserReader`** -- read-only user lookup for middleware and auth flows.

```go
type UserReader interface {
    GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
    GetByEmail(ctx context.Context, email string) (model.User, error)
}
```

**`UserStore`** -- the narrow user storage interface required by auth. Embeds `UserReader`.

```go
type UserStore interface {
    UserReader
    Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error)
    UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error)
}
```

| Method | Parameters | Returns | Description |
|--------|-----------|---------|-------------|
| `GetByID` | `ctx`, `id uuid.UUID` | `(model.User, error)` | Looks up a user by primary key. Returns `model.ErrUserNotFound` if absent. |
| `GetByEmail` | `ctx`, `email string` | `(model.User, error)` | Looks up a user by email. Returns `model.ErrUserNotFound` if absent. |
| `Create` | `ctx`, `email`, `displayName`, `role string`, `verified bool` | `(model.User, error)` | Creates a new user. Returns `model.ErrEmailTaken` on unique constraint violation. |
| `UpdateEmail` | `ctx`, `id uuid.UUID`, `email string` | `(model.User, error)` | Changes a user's email. Returns `model.ErrEmailTaken` on conflict. |

---

## service

Implements `auth.Service` with email-TOTP signup, signin, and email-change flows. Generates RFC 6238-compatible TOTP codes and produces self-contained encrypted verification tokens.

### Interfaces

**`TokenCodec`** -- encodes and decodes self-contained verification link payloads.

```go
type TokenCodec interface {
    Encode(model.VerificationToken) (string, error)
    Decode(string) (model.VerificationToken, error)
}
```

**`Notifier`** -- delivers verification details to the user via the host application.

```go
type Notifier interface {
    SendVerification(context.Context, model.DeliveryInput) error
}
```

### Types

**`Config`** -- parameterises the passwordless auth service.

| Field | Type | Description |
|-------|------|-------------|
| `Method` | `auth.EmailTOTPMethodConfig` | TOTP method configuration (issuer, period, enabled flag) |
| `Notifier` | `Notifier` | Email delivery implementation. Falls back to a no-op that returns an error if nil. |
| `TokenCodec` | `TokenCodec` | Token encoder/decoder for verification links |
| `Clock` | `clock` (unexported interface) | Time source for testing. Defaults to system UTC clock if nil. |

**`AuthService`** -- handles email-TOTP signup, signin, and email-change flows.

**`EncryptedTokenCodec`** -- produces self-contained encrypted verification tokens using AES-GCM encryption with HMAC-SHA256 authentication.

### Functions

**`NewAuthService(users repository.UserStore, cfg Config) *AuthService`** -- constructs an `AuthService` from its repository, notifier, and token codec dependencies.

```go
codec := service.NewEncryptedTokenCodec(authKey, encryptKey)
svc := service.NewAuthService(userRepo, service.Config{
    Method: auth.EmailTOTPMethodConfig{
        Enabled:       true,
        Issuer:        "MyApp",
        PeriodSeconds: 300,
    },
    Notifier:   emailNotifier,
    TokenCodec: codec,
})
```

**`NewEncryptedTokenCodec(authKey, encryptKey string) *EncryptedTokenCodec`** -- constructs a `TokenCodec` backed by signed AES-GCM tokens. The `authKey` is used for HMAC-SHA256 signing and `encryptKey` for AES-GCM encryption.

### Methods (AuthService)

**`BeginSignup(ctx, input, verifyPath) (PendingFlow, error)`** -- starts a signup verification flow. Checks that the email is not already registered, generates a TOTP secret, sends the verification email with a 6-digit code and verification link, and returns a `PendingFlow` for session storage. Returns `model.ErrEmailTaken` if the email is already in use.

**`BeginSignin(ctx, input, verifyPath) (PendingFlow, error)`** -- starts a signin verification flow. Always returns a `PendingFlow` regardless of whether the email exists, preventing account enumeration. Only sends the verification email if the account exists and is verified.

**`BeginEmailChange(ctx, userID, input, verifyPath) (PendingFlow, error)`** -- starts an email-change verification flow for a signed-in user. Validates that the new email differs from the current one and is not already taken.

**`ResendPendingFlow(ctx, flow, verifyPath) (PendingFlow, error)`** -- starts a fresh verification cycle for the current pending flow. Generates a new secret and code while preserving the original flow purpose and parameters.

**`VerifyPendingFlow(ctx, flow, input) (VerifyResult, error)`** -- completes a same-browser verification using pending session state. Validates the TOTP code against the flow secret and, on success, creates the user (signup), authenticates the user (signin), or updates the email (email change).

**`VerifyToken(ctx, token, input) (VerifyResult, error)`** -- completes a cross-browser verification using a self-contained encrypted token from the email link. Decodes the token, validates the code, and completes the flow.

**`VerifyPageState(token) (VerifyPageState, error)`** -- decodes an emailed token so the verification page can prefill the code and render the appropriate UI.

### Methods (EncryptedTokenCodec)

**`Encode(token VerificationToken) (string, error)`** -- serializes a verification token into a compact, URL-safe, encrypted string. The token format is a three-part dot-separated string: `header.body.signature`.

**`Decode(raw string) (VerificationToken, error)`** -- validates the HMAC signature, decrypts the AES-GCM payload, and checks expiration. Returns `model.ErrVerificationMissing` for any tampering or format error, and `model.ErrVerificationExpired` for expired tokens.

---

## session

Encrypted cookie-based session management backed by [gorilla/sessions]. Provides typed operations for storing user identity, display names, and pending verification flows.

### Types

**`SessionConfig`** -- cookie store configuration.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Cookie name. Defaults to `"session"` when empty. |
| `AuthKey` | `string` | HMAC signing key. Must be 32 bytes (SHA-256) or 64 bytes (SHA-512). |
| `EncryptKey` | `string` | AES encryption key. Must be 16 (AES-128), 24 (AES-192), or 32 (AES-256) bytes. |
| `MaxAge` | `int` | Cookie max age in seconds. |
| `Secure` | `bool` | Whether to set the `Secure` flag (HTTPS only). |

**`SessionManager`** -- wraps a [gorilla/sessions] `CookieStore` and provides typed session operations.

### Error Sentinels

| Error | Description |
|-------|-------------|
| `ErrNotAuthenticated` | The session contains no user ID or display name |
| `ErrPendingFlowNotFound` | No pending verification flow is stored in the session |

### Functions

**`NewSessionStore(cfg SessionConfig) (*SessionManager, error)`** -- creates a `SessionManager` backed by a signed and encrypted cookie store. Returns an error if key lengths are invalid. Cookie options are set to `HttpOnly`, `SameSite=Strict`, and `Path=/`.

```go
sm, err := session.NewSessionStore(session.SessionConfig{
    Name:       "app-session",
    AuthKey:    os.Getenv("SESSION_AUTH_KEY"),    // 32 or 64 bytes
    EncryptKey: os.Getenv("SESSION_ENCRYPT_KEY"), // 16, 24, or 32 bytes
    MaxAge:     86400,
    Secure:     true,
})
if err != nil {
    return fmt.Errorf("session store: %w", err)
}
```

### Methods

**`SetUserID(w, r, userID string) error`** -- stores the user ID in the session cookie.

**`GetUserID(r) (string, error)`** -- reads the user ID from the session cookie. Returns `("", ErrNotAuthenticated)` when no user ID is present.

**`SetDisplayName(w, r, name string) error`** -- stores the user's display name in the session cookie.

**`GetDisplayName(r) (string, error)`** -- reads the display name from the session cookie. Returns `("", ErrNotAuthenticated)` when no display name is present.

**`SetPendingFlow(w, r, flow PendingFlow) error`** -- stores the browser-bound verification flow in the session cookie as JSON.

**`GetPendingFlow(r) (PendingFlow, error)`** -- reads the pending verification flow from the session cookie. Returns `(PendingFlow{}, ErrPendingFlowNotFound)` when no flow is stored.

**`ClearPendingFlow(w, r) error`** -- removes the pending verification flow from the session without affecting other session data.

**`Clear(w, r) error`** -- invalidates the entire session by setting `MaxAge` to `-1`.

---

## Configuration

Below is an annotated YAML snippet showing the auth section of `app.yaml`:

```yaml
auth:
  selected: email_totp
  methods:
    email_totp:
      enabled: true
      issuer: MyApp
      period_seconds: 300    # 5 minutes
```

Session keys are provided via environment variables:

```yaml
session:
  name: app-session
  auth_key: ${SESSION_AUTH_KEY}         # 32 or 64 bytes
  encrypt_key: ${SESSION_ENCRYPT_KEY}   # 16, 24, or 32 bytes
  max_age: 86400                        # 24 hours
  secure: true
```

---

## Wiring Example

The following example shows the auth sub-packages wired together in a typical application startup sequence.

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/go-sum/auth"
    "github.com/go-sum/auth/service"
    "github.com/go-sum/auth/session"
)

func main() {
    // 1. Validate auth configuration
    if err := auth.ValidateConfig(cfg.Auth); err != nil {
        slog.Error("invalid auth config", "error", err)
        os.Exit(1)
    }

    // 2. Create session manager
    sm, err := session.NewSessionStore(session.SessionConfig{
        Name:       cfg.Session.Name,
        AuthKey:    os.Getenv("SESSION_AUTH_KEY"),
        EncryptKey: os.Getenv("SESSION_ENCRYPT_KEY"),
        MaxAge:     cfg.Session.MaxAge,
        Secure:     cfg.Session.Secure,
    })
    if err != nil {
        slog.Error("session store failed", "error", err)
        os.Exit(1)
    }

    // 3. Create token codec for verification links
    codec := service.NewEncryptedTokenCodec(
        os.Getenv("AUTH_TOKEN_AUTH_KEY"),
        os.Getenv("AUTH_TOKEN_ENCRYPT_KEY"),
    )

    // 4. Create auth service
    authSvc := service.NewAuthService(userRepo, service.Config{
        Method:     cfg.Auth.Methods.EmailTOTP,
        Notifier:   emailNotifier,
        TokenCodec: codec,
    })

    // 5. Use in handlers
    // authSvc satisfies auth.Service
    // sm provides session operations for login/logout
    _ = authSvc
    _ = sm
}
```

---

## Security Considerations

### Anti-Enumeration

`BeginSignin` always returns a valid `PendingFlow` regardless of whether the email exists. Verification emails are only sent to verified accounts. This prevents attackers from determining which email addresses have accounts.

### Timing Attack Prevention

TOTP code comparison uses constant-time byte comparison (`subtleConstantCompare`) to prevent timing side-channel attacks on the verification code.

### Token Security

Verification tokens use a layered security model:
1. **AES-GCM encryption** protects token payload confidentiality and integrity
2. **HMAC-SHA256 signature** authenticates the header and encrypted body
3. **Expiration check** rejects tokens beyond the configured `PeriodSeconds` window
4. Any tampering, format error, or decryption failure returns `model.ErrVerificationMissing` without leaking details

### Session Cookie Hardening

Session cookies are configured with:
- `HttpOnly` -- prevents JavaScript access
- `SameSite=Strict` -- prevents cross-site request attachment
- `Secure` flag -- configurable, should be `true` in production (HTTPS only)
- Signed and encrypted via [gorilla/sessions] `CookieStore`

### Key Length Validation

`NewSessionStore` rejects invalid key lengths at startup:
- `AuthKey` must be 32 or 64 bytes
- `EncryptKey` must be 16, 24, or 32 bytes

### Secret Management

All cryptographic keys (session auth key, session encrypt key, token auth key, token encrypt key) must be provided via environment variables. Never hardcode keys in source files or configuration.

---

## Leaf-Node Rule

Every package in this module imports only the Go standard library and external modules. There are no imports from application-specific `internal/` packages and no cross-imports between sibling `pkg/` packages. This means the entire `github.com/go-sum/auth` module can be vendored into any Go project without pulling in application-specific code.

[gorilla/sessions]: https://github.com/gorilla/sessions
[google/uuid]: https://github.com/google/uuid
[validator]: https://github.com/go-playground/validator
