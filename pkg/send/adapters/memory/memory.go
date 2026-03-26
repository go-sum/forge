package memory

import (
	"context"
	"sync"

	"github.com/go-sum/send/core"
)

// Sender captures sent messages in memory. It is safe for concurrent use.
// Use it in tests to assert on outbound messages without a real mail provider.
type Sender struct {
	mu       sync.Mutex
	Messages []core.Message
}

// New constructs an in-memory sender.
func New() *Sender {
	return &Sender{}
}

// Send appends msg to Messages and returns nil.
func (s *Sender) Send(_ context.Context, msg core.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
	return nil
}

// Reset clears all captured messages.
func (s *Sender) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = nil
}
