package memory_test

import (
	"context"
	"testing"

	"github.com/go-sum/send"
	"github.com/go-sum/send/adapters/memory"
)

func TestSender_Send(t *testing.T) {
	s := memory.New()

	msg1 := send.Message{To: "a@example.com", Subject: "First"}
	msg2 := send.Message{To: "b@example.com", Subject: "Second"}

	if err := s.Send(context.Background(), msg1); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}
	if err := s.Send(context.Background(), msg2); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	if got := len(s.Messages); got != 2 {
		t.Fatalf("expected 2 messages, got %d", got)
	}
	if s.Messages[0].To != "a@example.com" {
		t.Errorf("expected first message To=a@example.com, got %q", s.Messages[0].To)
	}
	if s.Messages[1].Subject != "Second" {
		t.Errorf("expected second message Subject=Second, got %q", s.Messages[1].Subject)
	}
}

func TestSender_Reset(t *testing.T) {
	s := memory.New()
	_ = s.Send(context.Background(), send.Message{To: "x@example.com"})
	s.Reset()
	if len(s.Messages) != 0 {
		t.Fatalf("expected 0 messages after Reset, got %d", len(s.Messages))
	}
}
