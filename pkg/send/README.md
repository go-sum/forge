---
title: Email delivery
description: Provider-agnostic email delivery using a registry and adapter pattern.
weight: 20
---

# Email delivery

`github.com/go-sum/send` is a provider-agnostic email delivery package built on a registry+adapter pattern. It ships with adapters for development, testing, and two production providers. The module is standalone -- it imports only the standard library and `net/http`, with no `internal/` or cross-`pkg/` dependencies.

## External Services

| Service | API Docs |
|---------|----------|
| [Resend] | https://resend.com/docs/api-reference/emails/send-email |
| [MailChannels] | https://api.mailchannels.net/tx/v1/documentation |

## Features

- Single `Sender` interface with one method -- minimal surface for loose coupling
- Thread-safe `Registry` decouples provider selection from instantiation
- Built-in adapters for [Resend] and [MailChannels] HTTP APIs
- Log adapter for development -- logs all fields (including HTML) via `log/slog` without delivering
- No-op adapter for silent discard -- logs fields at INFO level and returns `nil`
- Thread-safe in-memory adapter for test assertions on outbound messages
- `NewWithURL` constructors on HTTP adapters enable test injection via `httptest.Server`
- Empty configuration defaults to the no-op adapter for safe local development
- Custom provider registration via `Register` for third-party providers
- `SendFrom` resolution per provider for retrieving the configured sender address

## Sub-packages

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `send` (root) | `github.com/go-sum/send` | Registry, factory, public API, re-exported core types |
| `core` | `github.com/go-sum/send/core` | Domain types: `Message` struct, `Sender` interface |
| `log` | `github.com/go-sum/send/adapters/log` | Development adapter: logs all fields including HTML body |
| `noop` | `github.com/go-sum/send/adapters/noop` | Development adapter: logs and discards |
| `memory` | `github.com/go-sum/send/adapters/memory` | Test adapter: captures messages in-memory |
| `resend` | `github.com/go-sum/send/adapters/resend` | [Resend] HTTP API adapter |
| `mailchannels` | `github.com/go-sum/send/adapters/mailchannels` | [MailChannels] HTTP API adapter |

---

## Core Types

### `Message`

A provider-agnostic email message. Re-exported from `core` at the root package level.

| Field | Type | Description |
|-------|------|-------------|
| `To` | `string` | Recipient address (required) |
| `From` | `string` | Sender address. When empty, the active adapter's configured `SendFrom` address is used. |
| `Subject` | `string` | Email subject line (required) |
| `HTML` | `string` | HTML body (optional, recommended) |
| `Text` | `string` | Plain-text fallback body (optional) |

### `Sender`

The delivery contract. Re-exported from `core` at the root package level.

```go
type Sender interface {
    Send(ctx context.Context, msg Message) error
}
```

---

## Registry API

### `ProviderName`

`ProviderName` (`string`) names a sender backend.

| Constant | Value |
|----------|-------|
| `ProviderLog` | `"log"` |
| `ProviderNoop` | `"noop"` |
| `ProviderMemory` | `"memory"` |
| `ProviderResend` | `"resend"` |
| `ProviderMailchannels` | `"mailchannels"` |

### `Config`

Top-level configuration that selects a provider and holds per-provider settings.

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `Selected` | `ProviderName` | `selected` | Provider to use. Empty defaults to `"noop"`. |
| `Providers` | `ProvidersConfig` | `providers` | Per-provider configuration blocks. |

Methods:

- `SelectedProvider() ProviderName` -- returns `Selected`, defaulting to `ProviderNoop` when empty.

### `ProvidersConfig`

Holds configuration for all built-in providers.

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `Log` | `struct{}` | `log` | No configuration required. |
| `Noop` | `struct{}` | `noop` | No configuration required. |
| `Memory` | `struct{}` | `memory` | No configuration required. |
| `Resend` | `HTTPProviderConfig` | `resend` | [Resend] credentials and sender address. |
| `Mailchannels` | `HTTPProviderConfig` | `mailchannels` | [MailChannels] credentials and sender address. |

### `HTTPProviderConfig`

Carries the credentials and defaults for HTTP-backed providers.

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `APIKey` | `string` | `api_key` | API key credential (required for HTTP providers). |
| `SendFrom` | `string` | `send_from` | Default sender address (required for HTTP providers). |

### `Provider`

Bundles the construction and config accessors for a provider.

| Field | Type | Description |
|-------|------|-------------|
| `Factory` | `Factory` | Constructs a `Sender` from the full `Config`. |
| `SendFrom` | `SendFromResolver` | Returns the provider-level default sender address. May be `nil`. |

### `Factory`

```go
type Factory func(Config) (Sender, error)
```

Constructs a `Sender` for the selected provider.

### `SendFromResolver`

```go
type SendFromResolver func(Config) string
```

Returns the provider-level default sender address.

### `Registry`

Thread-safe store of provider factories by name.

**`NewRegistry() *Registry`** -- returns a registry preloaded with all built-in providers (`log`, `noop`, `memory`, `resend`, `mailchannels`).

**`(r *Registry) Register(name ProviderName, provider Provider)`** -- adds or replaces a provider in the registry.

**`(r *Registry) New(cfg Config) (Sender, error)`** -- resolves `cfg.SelectedProvider()` in the registry, calls the provider's `Factory`, and returns the configured `Sender`. Returns an error if the provider name is not registered, the factory is `nil`, or the factory itself returns an error.

**`(r *Registry) SendFrom(cfg Config) string`** -- resolves the default sender address for the selected provider. Returns an empty string if the provider is unknown or has no `SendFrom` resolver.

### Package-Level Functions

**`DefaultRegistry`** -- the package-level `*Registry` used by `New` and `Register`. Preloaded with all built-in providers.

**`New(cfg Config) (Sender, error)`** -- constructs a configured `Sender` using `DefaultRegistry`. Equivalent to `DefaultRegistry.New(cfg)`.

**`Register(name ProviderName, provider Provider)`** -- adds or replaces a provider in `DefaultRegistry`.

---

## Adapters

### log

Logs all message fields (including HTML body) at INFO level via `log/slog` and returns `nil` without delivering. Use this adapter in development when you want to inspect the full email content in structured logs.

**Constructor:** `func New() *Sender`

No configuration required.

### noop

Logs message fields (excluding HTML body) at INFO level via `log/slog` and returns `nil` without delivering. Serves as the default adapter when `Config.Selected` is empty.

**Constructor:** `func New() *Sender`

No configuration required.

### memory

Captures all sent messages in a thread-safe slice. Intended for test use only.

**Constructor:** `func New() *Sender`

**Methods:**

- `Send(ctx context.Context, msg Message) error` -- appends `msg` to the internal message list; always returns `nil`
- `Sent() []core.Message` -- returns a copy of all captured messages; safe for concurrent access
- `Reset()` -- clears all captured messages

### resend

Delivers email via the [Resend] HTTP API.

**Constructors:**

- `func New(apiKey, sendFrom string) *Sender` -- uses the production endpoint `https://api.resend.com/emails`
- `func NewWithURL(apiKey, sendFrom, url string) *Sender` -- uses a custom endpoint for test injection

**Behavior:**

- Sends an HTTPS POST with `Authorization: Bearer <apiKey>` header
- Falls back to `sendFrom` when `msg.From` is empty
- Returns an error prefixed with `"resend:"` on non-2xx responses or transport failures

### mailchannels

Delivers email via the [MailChannels] HTTP API.

**Constructors:**

- `func New(apiKey, sendFrom string) *Sender` -- uses the production endpoint `https://api.mailchannels.net/tx/v1/send`
- `func NewWithURL(apiKey, sendFrom, url string) *Sender` -- uses a custom endpoint for test injection

**Behavior:**

- Sends an HTTPS POST with `X-API-Key: <apiKey>` header
- Falls back to `sendFrom` when `msg.From` is empty
- Returns an error prefixed with `"mailchannels:"` on non-2xx responses or transport failures

---

## Configuration

| Provider | `APIKey` | `SendFrom` | Notes |
|----------|----------|------------|-------|
| `log` | -- | -- | Logs all fields including HTML body |
| `noop` | -- | -- | Safe dev default; messages logged without HTML body |
| `memory` | -- | -- | Test use only; messages captured in-memory |
| `resend` | Required | Required | [Resend] account key |
| `mailchannels` | Required | Required | [MailChannels] API key |

### YAML Example

```yaml
send:
  selected: resend
  providers:
    resend:
      api_key: ${RESEND_API_KEY}
      send_from: "noreply@example.com"
    mailchannels:
      api_key: ${MAILCHANNELS_API_KEY}
      send_from: "noreply@example.com"
```

---

## Usage

### Production with Resend

```go
import "github.com/go-sum/send"

sender, err := send.New(send.Config{
    Selected: send.ProviderResend,
    Providers: send.ProvidersConfig{
        Resend: send.HTTPProviderConfig{
            APIKey:   os.Getenv("RESEND_API_KEY"),
            SendFrom: "noreply@example.com",
        },
    },
})
if err != nil {
    return fmt.Errorf("email setup: %w", err)
}

err = sender.Send(ctx, send.Message{
    To:      "alice@example.com",
    Subject: "Welcome",
    HTML:    "<h1>Hello, Alice</h1>",
    Text:    "Hello, Alice",
})
```

### Dev default (empty config defaults to noop)

```go
import "github.com/go-sum/send"

// Empty Selected falls back to "noop" -- logs messages, delivers nothing.
sender, err := send.New(send.Config{})
if err != nil {
    return fmt.Errorf("email setup: %w", err)
}

// This message is logged at INFO level and discarded.
_ = sender.Send(ctx, send.Message{
    To:      "dev@example.com",
    Subject: "Test email",
    Text:    "This will appear in structured logs only.",
})
```

### Testing with the memory adapter

```go
import (
    "testing"

    "github.com/go-sum/send"
    "github.com/go-sum/send/adapters/memory"
)

func TestWelcomeEmail(t *testing.T) {
    mem := memory.New()

    // Inject mem as the Sender dependency in the system under test.
    svc := NewUserService(mem)
    err := svc.Register(ctx, newUserInput)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    sent := mem.Sent()
    if len(sent) != 1 {
        t.Fatalf("expected 1 message, got %d", len(sent))
    }
    if sent[0].To != "alice@example.com" {
        t.Errorf("expected To=alice@example.com, got %s", sent[0].To)
    }
    if sent[0].Subject != "Welcome" {
        t.Errorf("expected Subject=Welcome, got %s", sent[0].Subject)
    }

    // Reset for subsequent test phases.
    mem.Reset()
    if len(mem.Sent()) != 0 {
        t.Fatalf("expected 0 messages after reset, got %d", len(mem.Sent()))
    }
}
```

### Resolving the default sender address

```go
import "github.com/go-sum/send"

cfg := send.Config{
    Selected: send.ProviderResend,
    Providers: send.ProvidersConfig{
        Resend: send.HTTPProviderConfig{
            APIKey:   os.Getenv("RESEND_API_KEY"),
            SendFrom: "noreply@example.com",
        },
    },
}

// Retrieve the provider-level default sender address.
from := send.DefaultRegistry.SendFrom(cfg) // "noreply@example.com"
```

### Custom provider registration

```go
import (
    "context"

    "github.com/go-sum/send"
    "github.com/go-sum/send/core"
)

// 1. Implement the Sender interface.
type sesSender struct {
    region string
}

func (s *sesSender) Send(ctx context.Context, msg core.Message) error {
    // ... deliver via AWS SES SDK ...
    return nil
}

// 2. Register a provider before calling New.
func init() {
    send.Register("ses", send.Provider{
        Factory: func(cfg send.Config) (send.Sender, error) {
            return &sesSender{region: "us-east-1"}, nil
        },
        SendFrom: func(cfg send.Config) string {
            return "noreply@example.com"
        },
    })
}

// 3. Select the custom provider via Config.
sender, err := send.New(send.Config{
    Selected: "ses",
})
```

### Using a custom registry

```go
import "github.com/go-sum/send"

// Create an isolated registry (does not affect DefaultRegistry).
registry := send.NewRegistry()

registry.Register("custom", send.Provider{
    Factory: func(cfg send.Config) (send.Sender, error) {
        return myCustomSender{}, nil
    },
})

sender, err := registry.New(send.Config{Selected: "custom"})
```

---

## Design Notes

- The registry pattern decouples provider selection from instantiation. Application code depends only on `Config` and the `Sender` interface, never on concrete adapter types.
- The `Sender` interface has a single method -- this keeps the contract minimal and makes it trivial to implement custom adapters or test fakes.
- The `memory` adapter is thread-safe via an internal mutex, so it can be used safely in concurrent test scenarios. Messages are accessed via the `Sent()` method, which returns a defensive copy.
- `NewWithURL` constructors on the `resend` and `mailchannels` adapters allow pointing the HTTP client at an `httptest.Server` for integration tests without hitting live APIs.
- The `log` adapter differs from `noop` in that it also emits the HTML body to structured logs, making it useful when inspecting rendered email templates during development.
- This module follows the `pkg/` leaf-node rule: it imports only the standard library and `net/http`. There are no imports from `internal/` or other `pkg/` packages.

[Resend]: https://resend.com/
[MailChannels]: https://api.mailchannels.net/tx/v1/documentation
