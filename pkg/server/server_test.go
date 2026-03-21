package server

import (
	"strings"
	"testing"
	"time"
)

func TestNewInitializesEcho(t *testing.T) {
	e := New(Config{Debug: true})
	if e == nil {
		t.Fatal("New() returned nil")
	}
	if e.HTTPErrorHandler == nil {
		t.Fatal("HTTPErrorHandler was not configured")
	}
}

func TestStartReturnsErrorForInvalidAddress(t *testing.T) {
	e := New(Config{})
	err := Start(e, Config{
		Host:            "127.0.0.1",
		Port:            "not-a-port",
		GracefulTimeout: time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "server:") {
		t.Fatalf("err = %v", err)
	}
}
