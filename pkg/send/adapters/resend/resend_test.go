package resend_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/send"
	"github.com/go-sum/send/adapters/resend"
)

func TestSender_Send_success(t *testing.T) {
	var capturedAuth, capturedCT string
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		capturedCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := resend.NewWithURL("test-key", "no-reply@example.com", srv.URL)
	err := s.Send(context.Background(), send.Message{
		To:      "user@example.com",
		Subject: "Hello",
		HTML:    "<p>Hi</p>",
		Text:    "Hi",
	})
	if err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}
	if capturedAuth != "Bearer test-key" {
		t.Errorf("expected Authorization=Bearer test-key, got %q", capturedAuth)
	}
	if capturedCT != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", capturedCT)
	}
	if got, ok := capturedBody["from"].(string); !ok || got != "no-reply@example.com" {
		t.Errorf("expected from=no-reply@example.com, got %v", capturedBody["from"])
	}
	if to, ok := capturedBody["to"].([]any); !ok || len(to) != 1 || to[0] != "user@example.com" {
		t.Errorf("unexpected to field: %v", capturedBody["to"])
	}
}

func TestSender_Send_nonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	s := resend.NewWithURL("bad-key", "no-reply@example.com", srv.URL)
	err := s.Send(context.Background(), send.Message{To: "x@example.com", Subject: "Test"})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
