package core

import "context"

// Message is a provider-agnostic email message.
type Message struct {
	// To is the recipient address.
	To string
	// From is the sender address. When empty, the active adapter's configured
	// send_from address is used.
	From string
	// Subject is the email subject line.
	Subject string
	// HTML is the HTML body. Recommended for email clients that support it.
	HTML string
	// Text is the plain-text fallback body (RFC 5322 CRLF line endings).
	Text string
}

// Sender delivers email messages.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}
