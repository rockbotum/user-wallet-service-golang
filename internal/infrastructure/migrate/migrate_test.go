//go:build integration

package migrate

import (
	"fmt"
	"os"
	"testing"

	"user-wallet-service/internal/infrastructure/config"
)

func TestMigrateUp(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../../migrations"
	}

	if err := Run(databaseURL, migrationsPath); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	t.Log("migrations applied successfully")

	// повторный запуск не должен упасть
	if err := Run(databaseURL, migrationsPath); err != nil {
		t.Fatalf("Run() on already applied migrations error = %v", err)
	}

	t.Log("idempotent migration run OK")
}
