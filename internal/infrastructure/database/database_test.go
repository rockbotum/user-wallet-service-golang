//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"user-wallet-service/internal/infrastructure/config"
)

func TestDatabaseConnection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	t.Log("database connection successful")

	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	t.Logf("database version: %s", version)

	var now time.Time
	if err := db.QueryRowContext(ctx, "SELECT NOW()").Scan(&now); err != nil {
		t.Fatalf("query now: %v", err)
	}
	t.Logf("database time: %s", now.Format(time.RFC3339))

	t.Log("all connection tests passed")
}
