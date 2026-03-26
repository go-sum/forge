// Package send provides provider-agnostic email delivery.
//
// Public API:
//   - [Message] and [Sender] are the provider-agnostic email contract.
//   - [InitSender] constructs a configured Sender via the adapter registry.
//   - Built-in adapters live under [github.com/go-sum/send/adapters/...].
package send

import (
	"fmt"

	"github.com/go-sum/send/adapters/mailchannels"
	"github.com/go-sum/send/adapters/memory"
	"github.com/go-sum/send/adapters/noop"
	"github.com/go-sum/send/adapters/resend"
	"github.com/go-sum/send/core"
)

// Message is a provider-agnostic email message.
type Message = core.Message

// Sender delivers email messages.
type Sender = core.Sender

// SendAdapter names a sender backend.
type SendAdapter string

const (
	SendAdapterNoop         SendAdapter = "noop"
	SendAdapterMemory       SendAdapter = "memory"
	SendAdapterResend       SendAdapter = "resend"
	SendAdapterMailchannels SendAdapter = "mailchannels"
)

// AdapterConfig carries the credentials and defaults for a single adapter.
type AdapterConfig struct {
	APIKey   string
	SendFrom string
}

// Config holds the selected adapter name and the credentials for the active adapter.
type Config struct {
	Adapter  string
	SendFrom string
	APIKey   string
}

// SenderFactory constructs a Sender from an AdapterConfig.
type SenderFactory func(cfg AdapterConfig) (Sender, error)

// senderFactories is the registry of all known adapter constructors.
var senderFactories = map[SendAdapter]SenderFactory{
	SendAdapterNoop: func(_ AdapterConfig) (Sender, error) {
		return noop.New(), nil
	},
	SendAdapterMemory: func(_ AdapterConfig) (Sender, error) {
		return memory.New(), nil
	},
	SendAdapterResend: func(cfg AdapterConfig) (Sender, error) {
		return resend.New(cfg.APIKey, cfg.SendFrom), nil
	},
	SendAdapterMailchannels: func(cfg AdapterConfig) (Sender, error) {
		return mailchannels.New(cfg.APIKey, cfg.SendFrom), nil
	},
}

// Register adds or replaces a factory for the given adapter name.
// Intended for adapters not bundled with this package.
func Register(adapter SendAdapter, factory SenderFactory) {
	senderFactories[adapter] = factory
}

// InitSender looks up the registered factory for cfg.Adapter and returns a configured Sender.
// An empty Adapter name defaults to noop.
// Returns an error if the adapter name is not registered.
func InitSender(cfg Config) (Sender, error) {
	adapter := SendAdapter(cfg.Adapter)
	if adapter == "" {
		adapter = SendAdapterNoop
	}
	factory, ok := senderFactories[adapter]
	if !ok {
		return nil, fmt.Errorf("send: unknown adapter %q", cfg.Adapter)
	}
	return factory(AdapterConfig{APIKey: cfg.APIKey, SendFrom: cfg.SendFrom})
}
