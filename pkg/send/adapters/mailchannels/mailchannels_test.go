package mailchannels_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/send"
	"github.com/go-sum/send/adapters/mailchannels"
)

func TestSender_Send_success(t *testing.T) {
	var capturedAuth, capturedCT string
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("X-API-Key")
		capturedCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &capturedBody)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	s := mailchannels.NewWithURL("test-key", "no-reply@example.com", srv.URL)
	err := s.Send(context.Background(), send.Message{
		To:      "user@example.com",
		Subject: "Hello",
		HTML:    "<p>Hi</p>",
		Text:    "Hi",
	})
	if err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}
	if capturedAuth != "test-key" {
		t.Errorf("expected X-API-Key=test-key, got %q", capturedAuth)
	}
	if capturedCT != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", capturedCT)
	}
	from, _ := capturedBody["from"].(map[string]any)
	if from["email"] != "no-reply@example.com" {
		t.Errorf("expected from.email=no-reply@example.com, got %v", from["email"])
	}
}

func TestSender_Send_nonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s := mailchannels.NewWithURL("bad-key", "no-reply@example.com", srv.URL)
	err := s.Send(context.Background(), send.Message{To: "x@example.com", Subject: "Test"})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
