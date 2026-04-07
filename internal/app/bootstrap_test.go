package app

import (
	"testing"

	"github.com/go-sum/forge/config"
)

func TestInitAuthStoreDoesNotInstallSchemaAtRuntime(t *testing.T) {
	c := &Container{}

	c.initAuthStore()

	if c.AuthStore == nil {
		t.Fatal("AuthStore = nil, want initialized store")
	}
}

func TestInitQueueDoesNotInstallSchemaAtRuntime(t *testing.T) {
	c := &Container{
		Config: &config.Config{
			App: config.AppConfig{
				Queue: config.QueueConfig{
					Enabled: true,
					Queues: []config.QueueEntryConfig{
						{Name: "email"},
					},
				},
			},
		},
	}

	c.initQueue()

	if c.Queue == nil {
		t.Fatal("Queue = nil, want initialized client")
	}
	if !c.Queue.Async() {
		t.Fatal("Queue.Async() = false, want true")
	}
	if len(c.background) != 1 {
		t.Fatalf("background services = %d, want 1", len(c.background))
	}
}
