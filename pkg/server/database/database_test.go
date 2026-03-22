package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConnectRejectsInvalidDSN(t *testing.T) {
	if _, err := Connect(context.Background(), "not a dsn"); err == nil || !strings.Contains(err.Error(), "parsing database config") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckHealthReportsPingFailure(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/test?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	defer Close(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := CheckHealth(ctx, pool); err == nil {
		t.Fatal("CheckHealth() unexpectedly succeeded")
	}
}
