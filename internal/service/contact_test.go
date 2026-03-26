package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/send"
	"github.com/go-sum/send/adapters/memory"
)

func TestContactService_Submit_sendsNotificationAndConfirmation(t *testing.T) {
	sender := memory.New()
	cfg := &config.SendConfig{
		Adapter:  "noop",
		SendTo:   "admin@example.com",
		SendFrom: "no-reply@example.com",
	}
	svc := service.NewContactService(sender, cfg)

	input := model.ContactInput{
		Name:    "Alice",
		Email:   "alice@example.com",
		Message: "Hello there!",
	}
	if err := svc.Submit(context.Background(), input); err != nil {
		t.Fatalf("Submit returned unexpected error: %v", err)
	}

	if len(sender.Messages) != 2 {
		t.Fatalf("expected 2 messages sent, got %d", len(sender.Messages))
	}

	// First message: admin notification
	notify := sender.Messages[0]
	if notify.To != "admin@example.com" {
		t.Errorf("notification To: expected admin@example.com, got %q", notify.To)
	}
	if notify.From != "no-reply@example.com" {
		t.Errorf("notification From: expected no-reply@example.com, got %q", notify.From)
	}

	// Second message: submitter confirmation
	confirm := sender.Messages[1]
	if confirm.To != "alice@example.com" {
		t.Errorf("confirmation To: expected alice@example.com, got %q", confirm.To)
	}
	if confirm.From != "no-reply@example.com" {
		t.Errorf("confirmation From: expected no-reply@example.com, got %q", confirm.From)
	}
}

func TestContactService_Submit_propagatesSenderError(t *testing.T) {
	wantErr := errors.New("send failed")
	sender := &failSender{err: wantErr}
	cfg := &config.SendConfig{
		Adapter: "noop",
		SendTo:  "admin@example.com",
	}
	svc := service.NewContactService(sender, cfg)

	err := svc.Submit(context.Background(), model.ContactInput{
		Name: "Bob", Email: "bob@example.com", Message: "Hi",
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected wrapped send error, got: %v", err)
	}
}

type failSender struct{ err error }

func (f *failSender) Send(_ context.Context, _ send.Message) error { return f.err }
