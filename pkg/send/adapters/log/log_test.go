package log_test

import (
	"context"
	"testing"

	"github.com/go-sum/send"
	sendlog "github.com/go-sum/send/adapters/log"
)

func TestSender_Send(t *testing.T) {
	s := sendlog.New()
	err := s.Send(context.Background(), send.Message{
		To:      "user@example.com",
		From:    "no-reply@example.com",
		Subject: "Hello",
		HTML:    "<p>Hello</p>",
		Text:    "Hello",
	})
	if err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}
}
