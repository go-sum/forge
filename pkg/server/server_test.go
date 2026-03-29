package server

import (
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func TestNewWithConfigInitializesEcho(t *testing.T) {
	e := NewWithConfig(echo.Config{})
	if e == nil {
		t.Fatal("NewWithConfig() returned nil")
	}
}

func TestStartReturnsErrorForInvalidAddress(t *testing.T) {
	e := NewWithConfig(echo.Config{})
	err := Start(e, Config{
		Host:            "127.0.0.1",
		Port:            "not-a-port",
		GracefulTimeout: time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "server:") {
		t.Fatalf("err = %v", err)
	}
}
