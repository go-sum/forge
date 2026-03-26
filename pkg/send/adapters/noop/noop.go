package noop

import (
	"context"
	"log/slog"

	"github.com/go-sum/send/core"
)

// Sender discards all messages and logs them at INFO level.
// Use it as the default adapter in development to inspect outbound email in logs
// without requiring a live mail provider.
type Sender struct{}

// New constructs a no-op sender.
func New() *Sender {
	return &Sender{}
}

// Send logs the message fields and returns nil without delivering anything.
func (s *Sender) Send(ctx context.Context, msg core.Message) error {
	slog.InfoContext(ctx, "send: noop - message discarded",
		"to", msg.To,
		"from", msg.From,
		"subject", msg.Subject,
		"text", msg.Text,
	)
	return nil
}
