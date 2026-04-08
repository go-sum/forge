package database

import (
	"context"
	"strings"
	"testing"
)

func TestConnectRejectsInvalidDSN(t *testing.T) {
	if _, err := Connect(context.Background(), "not a dsn", 0); err == nil || !strings.Contains(err.Error(), "parsing database config") {
		t.Fatalf("err = %v", err)
	}
}
