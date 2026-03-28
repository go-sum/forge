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
	messages []core.Message
}

// New constructs an in-memory sender.
func New() *Sender {
	return &Sender{}
}

// Send appends msg to the captured message list and returns nil.
func (s *Sender) Send(_ context.Context, msg core.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
	return nil
}

// Sent returns a copy of all captured messages. Safe for concurrent access.
func (s *Sender) Sent() []core.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]core.Message, len(s.messages))
	copy(out, s.messages)
	return out
}

// Reset clears all captured messages.
func (s *Sender) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = nil
}
