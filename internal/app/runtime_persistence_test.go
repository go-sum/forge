package app

import (
	"testing"

	"github.com/go-sum/forge/config"
)

func TestInitAuthStoreDoesNotInstallSchemaAtRuntime(t *testing.T) {
	r := &Runtime{}

	r.initAuthStore()

	if r.AuthStore == nil {
		t.Fatal("AuthStore = nil, want initialized store")
	}
}

func TestInitQueueDoesNotInstallSchemaAtRuntime(t *testing.T) {
	r := &Runtime{
		Config: &config.Config{
			Store: config.StoreConfig{
				Queue: config.QueueConfig{
					Enabled: true,
					Queues: []config.QueueEntryConfig{
						{Name: "email"},
					},
				},
			},
		},
	}

	r.initQueue()

	if r.Queue == nil {
		t.Fatal("Queue = nil, want initialized client")
	}
	if !r.Queue.Async() {
		t.Fatal("Queue.Async() = false, want true")
	}
	if len(r.background) != 1 {
		t.Fatalf("background services = %d, want 1", len(r.background))
	}
}
