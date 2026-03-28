package send_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-sum/send"
	sendlog "github.com/go-sum/send/adapters/log"
	"github.com/go-sum/send/adapters/mailchannels"
	"github.com/go-sum/send/adapters/memory"
	"github.com/go-sum/send/adapters/noop"
	"github.com/go-sum/send/adapters/resend"
)

type fakeSender struct{}

func (fakeSender) Send(context.Context, send.Message) error { return nil }

func testConfig(selected send.ProviderName) send.Config {
	return send.Config{
		Selected: selected,
		Providers: send.ProvidersConfig{
			Resend: send.HTTPProviderConfig{
				APIKey:   "resend-key",
				SendFrom: "resend@example.com",
			},
			Mailchannels: send.HTTPProviderConfig{
				APIKey:   "mailchannels-key",
				SendFrom: "mailchannels@example.com",
			},
		},
	}
}

func TestDefaultRegistryKnownProviders(t *testing.T) {
	tests := []struct {
		name     string
		selected send.ProviderName
	}{
		{name: "log", selected: send.ProviderLog},
		{name: "noop", selected: send.ProviderNoop},
		{name: "memory", selected: send.ProviderMemory},
		{name: "resend", selected: send.ProviderResend},
		{name: "mailchannels", selected: send.ProviderMailchannels},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sender, err := send.DefaultRegistry.New(testConfig(tc.selected))
			if err != nil {
				t.Fatalf("DefaultRegistry.New() error = %v", err)
			}
			if sender == nil {
				t.Fatal("DefaultRegistry.New() returned nil sender")
			}
		})
	}
}

func TestNew_DefaultsToNoop(t *testing.T) {
	sender, err := send.New(send.Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, ok := sender.(*noop.Sender); !ok {
		t.Fatalf("expected *noop.Sender, got %T", sender)
	}
}

func TestNew_RejectsUnknownProvider(t *testing.T) {
	const unknown send.ProviderName = "totally-unknown"

	sender, err := send.New(send.Config{Selected: unknown})
	if err == nil {
		t.Fatal("New() error = nil, want unknown provider error")
	}
	if sender != nil {
		t.Fatalf("expected nil sender, got %T", sender)
	}
	if !strings.Contains(err.Error(), string(unknown)) {
		t.Fatalf("error %q does not mention provider %q", err, unknown)
	}
}

func TestNew_RequiresHTTPProviderConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     send.Config
		wantErr string
	}{
		{
			name: "resend missing api key",
			cfg: send.Config{
				Selected: send.ProviderResend,
				Providers: send.ProvidersConfig{
					Resend: send.HTTPProviderConfig{SendFrom: "from@example.com"},
				},
			},
			wantErr: "api_key",
		},
		{
			name: "mailchannels missing send_from",
			cfg: send.Config{
				Selected: send.ProviderMailchannels,
				Providers: send.ProvidersConfig{
					Mailchannels: send.HTTPProviderConfig{APIKey: "key"},
				},
			},
			wantErr: "send_from",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sender, err := send.New(tc.cfg)
			if err == nil {
				t.Fatal("New() error = nil, want validation error")
			}
			if sender != nil {
				t.Fatalf("expected nil sender, got %T", sender)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestRegistryRegisterCustomProvider(t *testing.T) {
	registry := send.NewRegistry()
	const custom send.ProviderName = "custom"

	want := fakeSender{}
	registry.Register(custom, send.Provider{
		Factory: func(send.Config) (send.Sender, error) {
			return want, nil
		},
		SendFrom: func(send.Config) string {
			return "custom@example.com"
		},
	})

	sender, err := registry.New(send.Config{Selected: custom})
	if err != nil {
		t.Fatalf("registry.New() error = %v", err)
	}
	if sender != want {
		t.Fatalf("registry.New() sender = %#v, want %#v", sender, want)
	}
	if got := registry.SendFrom(send.Config{Selected: custom}); got != "custom@example.com" {
		t.Fatalf("registry.SendFrom() = %q, want %q", got, "custom@example.com")
	}
}

func TestRegisterUpdatesDefaultRegistry(t *testing.T) {
	oldRegistry := send.DefaultRegistry
	send.DefaultRegistry = send.NewRegistry()
	t.Cleanup(func() { send.DefaultRegistry = oldRegistry })

	const custom send.ProviderName = "custom-default"
	send.Register(custom, send.Provider{
		Factory: func(send.Config) (send.Sender, error) {
			return fakeSender{}, nil
		},
	})

	sender, err := send.New(send.Config{Selected: custom})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, ok := sender.(fakeSender); !ok {
		t.Fatalf("expected fakeSender, got %T", sender)
	}
}

func TestConfigSendFrom(t *testing.T) {
	cfg := testConfig(send.ProviderMailchannels)
	if got := send.DefaultRegistry.SendFrom(cfg); got != "mailchannels@example.com" {
		t.Fatalf("SendFrom() = %q, want %q", got, "mailchannels@example.com")
	}

	cfg = testConfig(send.ProviderResend)
	if got := send.DefaultRegistry.SendFrom(cfg); got != "resend@example.com" {
		t.Fatalf("SendFrom() = %q, want %q", got, "resend@example.com")
	}

	if got := send.DefaultRegistry.SendFrom(send.Config{}); got != "" {
		t.Fatalf("SendFrom() for noop default = %q, want empty", got)
	}

	cfg = testConfig(send.ProviderLog)
	if got := send.DefaultRegistry.SendFrom(cfg); got != "" {
		t.Fatalf("SendFrom() for log = %q, want empty", got)
	}
}

func TestNew_BuiltInProviderTypes(t *testing.T) {
	t.Run("noop", func(t *testing.T) {
		sender, err := send.New(testConfig(send.ProviderNoop))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, ok := sender.(*noop.Sender); !ok {
			t.Fatalf("expected *noop.Sender, got %T", sender)
		}
	})

	t.Run("log", func(t *testing.T) {
		sender, err := send.New(testConfig(send.ProviderLog))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, ok := sender.(*sendlog.Sender); !ok {
			t.Fatalf("expected *log.Sender, got %T", sender)
		}
	})

	t.Run("memory", func(t *testing.T) {
		sender, err := send.New(testConfig(send.ProviderMemory))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, ok := sender.(*memory.Sender); !ok {
			t.Fatalf("expected *memory.Sender, got %T", sender)
		}
	})

	t.Run("resend", func(t *testing.T) {
		sender, err := send.New(testConfig(send.ProviderResend))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, ok := sender.(*resend.Sender); !ok {
			t.Fatalf("expected *resend.Sender, got %T", sender)
		}
	})

	t.Run("mailchannels", func(t *testing.T) {
		sender, err := send.New(testConfig(send.ProviderMailchannels))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if _, ok := sender.(*mailchannels.Sender); !ok {
			t.Fatalf("expected *mailchannels.Sender, got %T", sender)
		}
	})
}
