// Package send provides provider-agnostic email delivery.
package send

import (
	"fmt"
	"sync"

	"github.com/go-sum/send/adapters/log"
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

// ProviderName names a sender backend.
type ProviderName string

const (
	ProviderLog          ProviderName = "log"
	ProviderNoop         ProviderName = "noop"
	ProviderMemory       ProviderName = "memory"
	ProviderResend       ProviderName = "resend"
	ProviderMailchannels ProviderName = "mailchannels"
)

// HTTPProviderConfig carries the credentials and defaults for HTTP-backed providers.
type HTTPProviderConfig struct {
	APIKey   string `validate:"required"`
	SendFrom string `validate:"required"`
	Timeout  int    // HTTP client timeout in seconds; 0 → 10s default
}

// ProvidersConfig holds config for all built-in providers.
type ProvidersConfig struct {
	Log          struct{}
	Noop         struct{}
	Memory       struct{}
	Resend       HTTPProviderConfig
	Mailchannels HTTPProviderConfig
}

// Config holds the selected provider and its configured provider blocks.
type Config struct {
	Selected  ProviderName
	Providers ProvidersConfig
}

// Factory constructs a Sender for the selected provider.
type Factory func(Config) (Sender, error)

// SendFromResolver returns the provider-level default sender address.
type SendFromResolver func(Config) string

// Provider bundles the construction and config accessors for a provider.
type Provider struct {
	Factory  Factory
	SendFrom SendFromResolver
}

// Registry stores provider factories by name.
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderName]Provider
}

// NewRegistry returns a registry preloaded with the built-in providers.
func NewRegistry() *Registry {
	r := &Registry{providers: make(map[ProviderName]Provider)}
	r.Register(ProviderLog, Provider{
		Factory: func(Config) (Sender, error) {
			return log.New(), nil
		},
	})
	r.Register(ProviderNoop, Provider{
		Factory: func(Config) (Sender, error) {
			return noop.New(), nil
		},
	})
	r.Register(ProviderMemory, Provider{
		Factory: func(Config) (Sender, error) {
			return memory.New(), nil
		},
	})
	r.Register(ProviderResend, Provider{
		Factory: func(cfg Config) (Sender, error) {
			if err := requireHTTPProvider(ProviderResend, cfg.Providers.Resend); err != nil {
				return nil, err
			}
			p := cfg.Providers.Resend
			return resend.New(p.APIKey, p.SendFrom, p.Timeout), nil
		},
		SendFrom: func(cfg Config) string {
			return cfg.Providers.Resend.SendFrom
		},
	})
	r.Register(ProviderMailchannels, Provider{
		Factory: func(cfg Config) (Sender, error) {
			if err := requireHTTPProvider(ProviderMailchannels, cfg.Providers.Mailchannels); err != nil {
				return nil, err
			}
			p := cfg.Providers.Mailchannels
			return mailchannels.New(p.APIKey, p.SendFrom, p.Timeout), nil
		},
		SendFrom: func(cfg Config) string {
			return cfg.Providers.Mailchannels.SendFrom
		},
	})
	return r
}

// DefaultRegistry is the package-level registry used by New.
var DefaultRegistry = NewRegistry()

// SelectedProvider resolves the configured provider, defaulting to noop.
func (c Config) SelectedProvider() ProviderName {
	if c.Selected == "" {
		return ProviderNoop
	}
	return c.Selected
}

// Register adds or replaces a provider in the default registry.
func Register(name ProviderName, provider Provider) {
	DefaultRegistry.Register(name, provider)
}

// Register adds or replaces a provider in the registry.
func (r *Registry) Register(name ProviderName, provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

// New constructs a configured Sender using the registry.
func (r *Registry) New(cfg Config) (Sender, error) {
	providerName := cfg.SelectedProvider()

	r.mu.RLock()
	provider, ok := r.providers[providerName]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("send: unknown provider %q", providerName)
	}
	if provider.Factory == nil {
		return nil, fmt.Errorf("send: provider %q has no factory", providerName)
	}
	return provider.Factory(cfg)
}

// SendFrom resolves the default sender address for the selected provider.
func (r *Registry) SendFrom(cfg Config) string {
	providerName := cfg.SelectedProvider()

	r.mu.RLock()
	provider, ok := r.providers[providerName]
	r.mu.RUnlock()
	if !ok || provider.SendFrom == nil {
		return ""
	}
	return provider.SendFrom(cfg)
}

// New constructs a configured Sender from the selected built-in provider.
func New(cfg Config) (Sender, error) {
	return DefaultRegistry.New(cfg)
}

func requireHTTPProvider(provider ProviderName, cfg HTTPProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("send: provider %q requires api_key", provider)
	}
	if cfg.SendFrom == "" {
		return fmt.Errorf("send: provider %q requires send_from", provider)
	}
	return nil
}
