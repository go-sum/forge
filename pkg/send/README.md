---
title: Email delivery
description: Provider-agnostic email delivery using a factory and adapter pattern.
weight: 20
---

# Email delivery

`github.com/go-sum/send` is a provider-agnostic email delivery package built on a factory+adapter pattern. It ships with adapters for development, testing, and two production providers. The module is standalone -- it imports only the standard library and `net/http`, with no `internal/` or cross-`pkg/` dependencies.

## External Services

| Service | API Docs |
|---------|----------|
| [Resend] | https://resend.com/docs/api-reference/emails/send-email |
| [MailChannels] | https://api.mailchannels.net/tx/v1/documentation |

## Features

- Single `Sender` interface with one method -- minimal surface for loose coupling
- Factory registry decouples adapter selection from instantiation
- Built-in adapters for [Resend] and [MailChannels] HTTP APIs
- No-op adapter for development -- logs all fields via `log/slog` without delivering
- Thread-safe in-memory adapter for test assertions on outbound messages
- `NewWithURL` constructors on HTTP adapters enable test injection via `httptest.Server`
- Empty configuration defaults to the no-op adapter for safe local development
- Custom adapter registration via `Register` for third-party providers

## Package Structure

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `send` (root) | `github.com/go-sum/send` | Factory, registry, public API, re-exported core types |
| `core` | `github.com/go-sum/send/core` | Domain types: `Message` struct, `Sender` interface |
| `noop` | `github.com/go-sum/send/adapters/noop` | Development adapter: logs and discards |
| `memory` | `github.com/go-sum/send/adapters/memory` | Test adapter: captures messages in-memory |
| `resend` | `github.com/go-sum/send/adapters/resend` | [Resend] HTTP API adapter |
| `mailchannels` | `github.com/go-sum/send/adapters/mailchannels` | [MailChannels] HTTP API adapter |

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

## Factory API

### `Config`

| Field | Type | Description |
|-------|------|-------------|
| `Adapter` | `string` | Adapter name: `"noop"`, `"memory"`, `"resend"`, or `"mailchannels"`. Empty defaults to `"noop"`. |
| `SendFrom` | `string` | Default sender address passed to the adapter |
| `APIKey` | `string` | API key for external adapters |

### `AdapterConfig`

Passed to factory functions when constructing an adapter.

| Field | Type | Description |
|-------|------|-------------|
| `APIKey` | `string` | API key credential |
| `SendFrom` | `string` | Default sender address |

### Adapter Constants

| Constant | Value |
|----------|-------|
| `SendAdapterNoop` | `"noop"` |
| `SendAdapterMemory` | `"memory"` |
| `SendAdapterResend` | `"resend"` |
| `SendAdapterMailchannels` | `"mailchannels"` |

### Functions

**`InitSender(cfg Config) (Sender, error)`** -- looks up the registered factory for `cfg.Adapter` and returns a configured `Sender`. An empty `Adapter` defaults to `"noop"`. Returns an error if the adapter name is not registered.

**`Register(adapter SendAdapter, factory SenderFactory)`** -- adds or replaces a factory in the global registry. Use this to register third-party adapters not bundled with the package.

## Adapters

### noop

Logs all message fields at INFO level via `log/slog` and returns `nil` without delivering. Serves as the default adapter when `Config.Adapter` is empty.

**Constructor:** `func New() *Sender`

No configuration required.

### memory

Captures all sent messages in a thread-safe slice. Intended for test use only.

**Constructor:** `func New() *Sender`

**Fields:**

| Name | Type | Description |
|------|------|-------------|
| `Messages` | `[]core.Message` | All messages passed to `Send`, in order |

**Methods:**

- `Send(ctx context.Context, msg Message) error` -- appends `msg` to `Messages`; always returns `nil`
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

## Configuration

| Adapter | `APIKey` | `SendFrom` | Notes |
|---------|----------|------------|-------|
| `noop` | -- | -- | Safe dev default; messages only logged |
| `memory` | -- | -- | Test use only; messages captured in-memory |
| `resend` | Required | Required | [Resend] account key |
| `mailchannels` | Required | Required | [MailChannels] API key |

## Usage

### Init from config (production with Resend)

```go
import "github.com/go-sum/send"

sender, err := send.InitSender(send.Config{
    Adapter:  "resend",
    SendFrom: "noreply@example.com",
    APIKey:   os.Getenv("RESEND_API_KEY"),
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

// Empty Adapter falls back to "noop" -- logs messages, delivers nothing.
sender, err := send.InitSender(send.Config{})
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

    if len(mem.Messages) != 1 {
        t.Fatalf("expected 1 message, got %d", len(mem.Messages))
    }
    if mem.Messages[0].To != "alice@example.com" {
        t.Errorf("expected To=alice@example.com, got %s", mem.Messages[0].To)
    }
    if mem.Messages[0].Subject != "Welcome" {
        t.Errorf("expected Subject=Welcome, got %s", mem.Messages[0].Subject)
    }

    // Reset for subsequent test phases.
    mem.Reset()
    if len(mem.Messages) != 0 {
        t.Fatalf("expected 0 messages after reset, got %d", len(mem.Messages))
    }
}
```

### Custom adapter registration

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

// 2. Register a factory before calling InitSender.
func init() {
    send.Register("ses", func(cfg send.AdapterConfig) (send.Sender, error) {
        return &sesSender{region: "us-east-1"}, nil
    })
}

// 3. Select the custom adapter via Config.
sender, err := send.InitSender(send.Config{
    Adapter:  "ses",
    SendFrom: "noreply@example.com",
    APIKey:   os.Getenv("AWS_SES_API_KEY"),
})
```

## Design Notes

- The factory registry pattern decouples adapter selection from instantiation. Application code depends only on `Config` and the `Sender` interface, never on concrete adapter types.
- The `Sender` interface has a single method -- this keeps the contract minimal and makes it trivial to implement custom adapters or test fakes.
- The `memory` adapter is thread-safe via an internal mutex, so it can be used safely in concurrent test scenarios.
- `NewWithURL` constructors on the `resend` and `mailchannels` adapters allow pointing the HTTP client at an `httptest.Server` for integration tests without hitting live APIs.
- This module follows the `pkg/` leaf-node rule: it imports only the standard library and `net/http`. There are no imports from `internal/` or other `pkg/` packages.

[Resend]: https://resend.com/
[MailChannels]: https://api.mailchannels.net/tx/v1/documentation
