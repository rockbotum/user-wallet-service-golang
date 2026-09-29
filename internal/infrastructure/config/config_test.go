package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("JWT_ACCESS_TTL", "")
	t.Setenv("JWT_REFRESH_TTL", "")
	t.Setenv("DB_CONN_MAX_LIFETIME", "")
	t.Setenv("JWT_SECRET", "test-secret-that-is-long-enough-32ch")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != defaultHTTPPort {
		t.Fatalf("HTTPPort = %d, want %d", cfg.HTTPPort, defaultHTTPPort)
	}
	if cfg.HTTPReadTimeout != defaultHTTPReadTimeout {
		t.Fatalf("HTTPReadTimeout = %v, want %v", cfg.HTTPReadTimeout, defaultHTTPReadTimeout)
	}
	if cfg.DBPort != 5432 {
		t.Fatalf("DBPort = %d, want 5432", cfg.DBPort)
	}
	if cfg.DBMaxOpenConns != defaultDBMaxOpenConns {
		t.Fatalf("DBMaxOpenConns = %d, want %d", cfg.DBMaxOpenConns, defaultDBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != defaultDBMaxIdleConns {
		t.Fatalf("DBMaxIdleConns = %d, want %d", cfg.DBMaxIdleConns, defaultDBMaxIdleConns)
	}
	if cfg.JWTAccessTTL != 15*time.Minute {
		t.Fatalf("JWTAccessTTL = %v, want 15m", cfg.JWTAccessTTL)
	}
	if cfg.JWTRefreshTTL != 168*time.Hour {
		t.Fatalf("JWTRefreshTTL = %v, want 168h", cfg.JWTRefreshTTL)
	}
	if cfg.DBConnMaxLifetime != defaultDBConnMaxLifetime {
		t.Fatalf("DBConnMaxLifetime = %v, want %v", cfg.DBConnMaxLifetime, defaultDBConnMaxLifetime)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("DB_HOST", "custom-host")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "custom_user")
	t.Setenv("DB_PASSWORD", "custom_pass")
	t.Setenv("DB_NAME", "custom_db")
	t.Setenv("DB_MAX_OPEN_CONNS", "50")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("JWT_SECRET", "test-secret-that-is-long-enough-32ch")
	t.Setenv("JWT_ACCESS_TTL", "30m")
	t.Setenv("JWT_REFRESH_TTL", "720h")
	t.Setenv("DB_CONN_MAX_LIFETIME", "10m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != 9000 {
		t.Fatalf("HTTPPort = %d, want 9000", cfg.HTTPPort)
	}
	if cfg.DBHost != "custom-host" {
		t.Fatalf("DBHost = %s, want custom-host", cfg.DBHost)
	}
	if cfg.DBPort != 5433 {
		t.Fatalf("DBPort = %d, want 5433", cfg.DBPort)
	}
	if cfg.DBUser != "custom_user" {
		t.Fatalf("DBUser = %s, want custom_user", cfg.DBUser)
	}
	if cfg.DBPassword != "custom_pass" {
		t.Fatalf("DBPassword = %s, want custom_pass", cfg.DBPassword)
	}
	if cfg.DBName != "custom_db" {
		t.Fatalf("DBName = %s, want custom_db", cfg.DBName)
	}
	if cfg.DBMaxOpenConns != 50 {
		t.Fatalf("DBMaxOpenConns = %d, want 50", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 10 {
		t.Fatalf("DBMaxIdleConns = %d, want 10", cfg.DBMaxIdleConns)
	}
	if cfg.JWTSecret != "test-secret-that-is-long-enough-32ch" {
		t.Fatalf("JWTSecret = %s, want test-secret-that-is-long-enough-32ch", cfg.JWTSecret)
	}
	if cfg.JWTAccessTTL != 30*time.Minute {
		t.Fatalf("JWTAccessTTL = %v, want 30m", cfg.JWTAccessTTL)
	}
	if cfg.JWTRefreshTTL != 720*time.Hour {
		t.Fatalf("JWTRefreshTTL = %v, want 720h", cfg.JWTRefreshTTL)
	}
	if cfg.DBConnMaxLifetime != 10*time.Minute {
		t.Fatalf("DBConnMaxLifetime = %v, want 10m", cfg.DBConnMaxLifetime)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "abc")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for non-numeric port")
	}
}

func TestLoadNonPositivePort(t *testing.T) {
	t.Setenv("HTTP_PORT", "0")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for non-positive port")
	}
}

func TestLoadInvalidDBPort(t *testing.T) {
	t.Setenv("DB_PORT", "abc")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for non-numeric DB port")
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	t.Setenv("JWT_ACCESS_TTL", "invalid")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid duration")
	}
}

func TestLoadMissingJWTSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for missing JWT_SECRET")
	}
}

func TestLoadShortJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "short-secret")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for short JWT_SECRET")
	}
}
