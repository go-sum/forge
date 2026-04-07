package contact

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/queue"
	"github.com/go-sum/send"
	"github.com/go-sum/send/adapters/memory"
)

func newEmailQueue(sender send.Sender) *queue.Client {
	c := queue.New(nil, queue.Config{
		Queues: []queue.QueueConfig{
			{Name: "email", MaxAttempts: 1, Timeout: 5},
		},
	})
	c.Register("email", func(ctx context.Context, job queue.Job) error {
		var p EmailPayload
		if err := json.Unmarshal(job.Payload, &p); err != nil {
			return err
		}
		return sender.Send(ctx, send.Message{
			To: p.To, From: p.From, Subject: p.Subject, HTML: p.HTML, Text: p.Text,
		})
	})
	return c
}

func TestServiceSubmitSendsNotificationAndConfirmation(t *testing.T) {
	sender := memory.New()
	q := newEmailQueue(sender)
	cfg := Config{
		SendTo:   "admin@example.com",
		SendFrom: "no-reply@example.com",
	}
	svc := NewService(q, cfg)

	input := model.ContactInput{
		Name:    "Alice",
		Email:   "alice@example.com",
		Message: "Hello there!",
	}
	if err := svc.Submit(context.Background(), input); err != nil {
		t.Fatalf("Submit returned unexpected error: %v", err)
	}

	sent := sender.Sent()
	if len(sent) != 2 {
		t.Fatalf("expected 2 messages sent, got %d", len(sent))
	}

	notify := sent[0]
	if notify.To != "admin@example.com" {
		t.Errorf("notification To: expected admin@example.com, got %q", notify.To)
	}
	if notify.From != "no-reply@example.com" {
		t.Errorf("notification From: expected no-reply@example.com, got %q", notify.From)
	}

	confirm := sent[1]
	if confirm.To != "alice@example.com" {
		t.Errorf("confirmation To: expected alice@example.com, got %q", confirm.To)
	}
	if confirm.From != "no-reply@example.com" {
		t.Errorf("confirmation From: expected no-reply@example.com, got %q", confirm.From)
	}
}

func TestServiceSubmitPropagatesSenderError(t *testing.T) {
	wantErr := errors.New("send failed")
	sender := &failSender{err: wantErr}
	q := newEmailQueue(sender)
	svc := NewService(q, Config{SendTo: "admin@example.com"})

	err := svc.Submit(context.Background(), model.ContactInput{
		Name: "Bob", Email: "bob@example.com", Message: "Hi",
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected wrapped send error, got: %v", err)
	}
}

type failSender struct{ err error }

func (f *failSender) Send(_ context.Context, _ send.Message) error { return f.err }
