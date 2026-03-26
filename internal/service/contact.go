package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-sum/componentry/email"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/send"

	g "maragu.dev/gomponents"
)

// ContactService handles the contact form submission workflow.
type ContactService struct {
	sender send.Sender
	cfg    ContactConfig
}

// ContactConfig contains the delivery addresses the contact workflow needs.
type ContactConfig struct {
	SendTo   string
	SendFrom string
}

// NewContactService constructs a ContactService.
func NewContactService(sender send.Sender, cfg ContactConfig) *ContactService {
	return &ContactService{sender: sender, cfg: cfg}
}

// Submit sends a notification to the configured send_to address and a
// confirmation reply to the submitter's email address.
func (s *ContactService) Submit(ctx context.Context, input model.ContactInput) error {
	if err := s.sender.Send(ctx, send.Message{
		To:      s.cfg.SendTo,
		From:    s.cfg.SendFrom,
		Subject: "New contact form submission from " + input.Name,
		HTML:    renderHTML(notificationBody(input)),
		Text:    notificationText(input),
	}); err != nil {
		return fmt.Errorf("ContactService.Submit: send notification: %w", err)
	}

	if err := s.sender.Send(ctx, send.Message{
		To:      input.Email,
		From:    s.cfg.SendFrom,
		Subject: "Thanks for reaching out",
		HTML:    renderHTML(confirmationBody(input)),
		Text:    confirmationText(input),
	}); err != nil {
		return fmt.Errorf("ContactService.Submit: send confirmation: %w", err)
	}

	return nil
}

func renderHTML(body g.Node) string {
	var buf bytes.Buffer
	_ = body.Render(&buf)
	return buf.String()
}

func notificationBody(input model.ContactInput) g.Node {
	return email.Layout("New Contact Submission", g.Group([]g.Node{
		email.H1("New contact form submission"),
		email.P("Name: " + input.Name),
		email.P("Email: " + input.Email),
		email.P("Message:"),
		email.P(input.Message),
	}))
}

func notificationText(input model.ContactInput) string {
	return email.PlainText(
		"New contact form submission",
		"",
		"Name: "+input.Name,
		"Email: "+input.Email,
		"",
		"Message:",
		input.Message,
	)
}

func confirmationBody(input model.ContactInput) g.Node {
	return email.Layout("Thanks for reaching out", g.Group([]g.Node{
		email.H1("Thanks for reaching out, " + input.Name + "!"),
		email.P("We've received your message and will get back to you soon."),
		email.P("Your message:"),
		email.P(input.Message),
	}))
}

func confirmationText(input model.ContactInput) string {
	return email.PlainText(
		"Thanks for reaching out, "+input.Name+"!",
		"",
		"We've received your message and will get back to you soon.",
		"",
		"Your message:",
		input.Message,
	)
}
