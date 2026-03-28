package log

import (
	"context"
	"log/slog"

	"github.com/go-sum/send/core"
)

// Sender logs outbound email details at INFO level and performs no network delivery.
type Sender struct{}

// New constructs a logging sender.
func New() *Sender {
	return &Sender{}
}

// Send writes the message fields to the standard structured logger.
func (s *Sender) Send(ctx context.Context, msg core.Message) error {
	slog.InfoContext(ctx, "send: log - message emitted to logger",
		"to", msg.To,
		"from", msg.From,
		"subject", msg.Subject,
		"text", msg.Text,
		"html", msg.HTML,
	)
	return nil
}
