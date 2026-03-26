package send_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-sum/send"
	"github.com/go-sum/send/adapters/mailchannels"
	"github.com/go-sum/send/adapters/memory"
	"github.com/go-sum/send/adapters/noop"
	"github.com/go-sum/send/adapters/resend"
)

type fakeSender struct {
	cfg send.AdapterConfig
}

func (f *fakeSender) Send(_ context.Context, _ send.Message) error { return nil }

func registerTemp(t *testing.T, name send.SendAdapter, f send.SenderFactory) {
	t.Helper()
	send.Register(name, f)
	t.Cleanup(func() {
		send.Register(name, func(_ send.AdapterConfig) (send.Sender, error) {
			return nil, errors.New("send: unknown adapter (removed by test cleanup)")
		})
	})
}

func TestInitSender_KnownAdapters(t *testing.T) {
	tests := []struct {
		name    string
		adapter send.SendAdapter
		cfg     send.Config
	}{
		{
			name:    "noop",
			adapter: send.SendAdapterNoop,
			cfg:     send.Config{Adapter: string(send.SendAdapterNoop)},
		},
		{
			name:    "memory",
			adapter: send.SendAdapterMemory,
			cfg:     send.Config{Adapter: string(send.SendAdapterMemory)},
		},
		{
			name:    "resend",
			adapter: send.SendAdapterResend,
			cfg: send.Config{
				Adapter:  string(send.SendAdapterResend),
				APIKey:   "key",
				SendFrom: "from@example.com",
			},
		},
		{
			name:    "mailchannels",
			adapter: send.SendAdapterMailchannels,
			cfg: send.Config{
				Adapter:  string(send.SendAdapterMailchannels),
				APIKey:   "key",
				SendFrom: "from@example.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := send.InitSender(tc.cfg)
			if err != nil {
				t.Fatalf("InitSender(%q) returned unexpected error: %v", tc.adapter, err)
			}
			if s == nil {
				t.Fatalf("InitSender(%q) returned nil Sender", tc.adapter)
			}
		})
	}
}

func TestInitSender_NoopSenderType(t *testing.T) {
	s, err := send.InitSender(send.Config{Adapter: string(send.SendAdapterNoop)})
	if err != nil {
		t.Fatalf("InitSender(noop) unexpected error: %v", err)
	}
	if _, ok := s.(*noop.Sender); !ok {
		t.Fatalf("expected *noop.Sender, got %T", s)
	}
}

func TestInitSender_MemorySenderType(t *testing.T) {
	s, err := send.InitSender(send.Config{Adapter: string(send.SendAdapterMemory)})
	if err != nil {
		t.Fatalf("InitSender(memory) unexpected error: %v", err)
	}
	if _, ok := s.(*memory.Sender); !ok {
		t.Fatalf("expected *memory.Sender, got %T", s)
	}
}

func TestInitSender_EmptyAdapterDefaultsToNoop(t *testing.T) {
	s, err := send.InitSender(send.Config{})
	if err != nil {
		t.Fatalf("InitSender with empty Adapter returned unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("InitSender with empty Adapter returned nil Sender")
	}
	if _, ok := s.(*noop.Sender); !ok {
		t.Fatalf("expected *noop.Sender for empty adapter, got %T", s)
	}
}

func TestInitSender_UnknownAdapter(t *testing.T) {
	const unknownName send.SendAdapter = "totally-unknown-adapter-xyz"

	s, err := send.InitSender(send.Config{Adapter: string(unknownName)})
	if err == nil {
		t.Fatal("expected error for unknown adapter, got nil")
	}
	if s != nil {
		t.Fatalf("expected nil Sender for unknown adapter, got %T", s)
	}
	if !strings.Contains(err.Error(), string(unknownName)) {
		t.Errorf("error %q does not contain adapter name %q", err.Error(), unknownName)
	}
}

func TestRegister_CustomAdapter(t *testing.T) {
	const customName send.SendAdapter = "custom-test-adapter-register"

	customSender := &fakeSender{}
	registerTemp(t, customName, func(_ send.AdapterConfig) (send.Sender, error) {
		return customSender, nil
	})

	s, err := send.InitSender(send.Config{Adapter: string(customName)})
	if err != nil {
		t.Fatalf("InitSender(%q) after Register returned unexpected error: %v", customName, err)
	}
	if s != customSender {
		t.Fatalf("expected custom sender instance, got %T", s)
	}
}

func TestRegister_ReplaceExistingAdapter(t *testing.T) {
	const replaceName send.SendAdapter = "custom-test-adapter-replace"

	original := &fakeSender{cfg: send.AdapterConfig{APIKey: "original"}}
	registerTemp(t, replaceName, func(_ send.AdapterConfig) (send.Sender, error) {
		return original, nil
	})

	s1, err := send.InitSender(send.Config{Adapter: string(replaceName)})
	if err != nil {
		t.Fatalf("first InitSender error: %v", err)
	}
	if s1 != original {
		t.Fatalf("expected original sender, got %T", s1)
	}

	replacement := &fakeSender{cfg: send.AdapterConfig{APIKey: "replacement"}}
	send.Register(replaceName, func(_ send.AdapterConfig) (send.Sender, error) {
		return replacement, nil
	})

	s2, err := send.InitSender(send.Config{Adapter: string(replaceName)})
	if err != nil {
		t.Fatalf("second InitSender error: %v", err)
	}
	if s2 != replacement {
		t.Fatalf("expected replacement sender, got %T", s2)
	}
}

func TestInitSender_AdapterConfigRouting(t *testing.T) {
	tests := []struct {
		name     string
		adapter  send.SendAdapter
		apiKey   string
		sendFrom string
	}{
		{
			name:     "resend routes flat config",
			adapter:  send.SendAdapterResend,
			apiKey:   "resend-key",
			sendFrom: "resend@example.com",
		},
		{
			name:     "mailchannels routes flat config",
			adapter:  send.SendAdapterMailchannels,
			apiKey:   "mc-key",
			sendFrom: "mc@example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			customName := send.SendAdapter("capture-" + string(tc.adapter))
			var captured send.AdapterConfig
			registerTemp(t, customName, func(cfg send.AdapterConfig) (send.Sender, error) {
				captured = cfg
				return &fakeSender{cfg: cfg}, nil
			})

			want := send.AdapterConfig{APIKey: tc.apiKey, SendFrom: tc.sendFrom}
			cfg := send.Config{
				Adapter:  string(customName),
				APIKey:   tc.apiKey,
				SendFrom: tc.sendFrom,
			}

			_, err := send.InitSender(cfg)
			if err != nil {
				t.Fatalf("InitSender(%q) error: %v", customName, err)
			}

			if captured != want {
				t.Errorf("factory received AdapterConfig %+v, want %+v", captured, want)
			}
		})
	}
}

func TestInitSender_FactoryError(t *testing.T) {
	const errorAdapterName send.SendAdapter = "custom-test-factory-error"
	wantErr := errors.New("factory intentional error")

	registerTemp(t, errorAdapterName, func(_ send.AdapterConfig) (send.Sender, error) {
		return nil, wantErr
	})

	s, err := send.InitSender(send.Config{Adapter: string(errorAdapterName)})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected factory error %v, got %v", wantErr, err)
	}
	if s != nil {
		t.Fatalf("expected nil Sender on factory error, got %T", s)
	}
}

func TestInitSender_BuiltInProviderTypes(t *testing.T) {
	t.Run("resend sender", func(t *testing.T) {
		s, err := send.InitSender(send.Config{
			Adapter:  string(send.SendAdapterResend),
			APIKey:   "key",
			SendFrom: "from@example.com",
		})
		if err != nil {
			t.Fatalf("InitSender returned unexpected error: %v", err)
		}
		if _, ok := s.(*resend.Sender); !ok {
			t.Fatalf("expected *resend.Sender, got %T", s)
		}
	})

	t.Run("mailchannels sender", func(t *testing.T) {
		s, err := send.InitSender(send.Config{
			Adapter:  string(send.SendAdapterMailchannels),
			APIKey:   "key",
			SendFrom: "from@example.com",
		})
		if err != nil {
			t.Fatalf("InitSender returned unexpected error: %v", err)
		}
		if _, ok := s.(*mailchannels.Sender); !ok {
			t.Fatalf("expected *mailchannels.Sender, got %T", s)
		}
	})
}
