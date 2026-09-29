// Package config отвечает за загрузку и хранение конфигурации приложения.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPPort          = 8080
	defaultHTTPReadTimeout   = 10 * time.Second
	defaultHTTPWriteTimeout  = 10 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultDBMaxOpenConns    = 25
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 5 * time.Minute

	// minJWTSecretLen — минимальная длина секрета для HMAC-SHA256 подписи токенов.
	minJWTSecretLen = 32
)

// Config — конфигурация приложения.
type Config struct {
	HTTPPort         int
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	ShutdownTimeout  time.Duration

	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	RateLimitRPS   float64
	RateLimitBurst int
	RateLimitTTL   time.Duration
}

// Load читает конфигурацию из переменных окружения, применяя значения по умолчанию.
func Load() (Config, error) {
	port, err := envInt("HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return Config{}, err
	}

	dbPort, err := envInt("DB_PORT", 5432)
	if err != nil {
		return Config{}, err
	}

	dbMaxOpenConns, err := envInt("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns)
	if err != nil {
		return Config{}, err
	}

	dbMaxIdleConns, err := envInt("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns)
	if err != nil {
		return Config{}, err
	}

	jwtAccessTTL, err := envDuration("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	jwtRefreshTTL, err := envDuration("JWT_REFRESH_TTL", 168*time.Hour)
	if err != nil {
		return Config{}, err
	}

	dbConnMaxLifetime, err := envDuration("DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}

	// Секрет JWT обязателен: дефолтное значение позволило бы подделывать токены.
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < minJWTSecretLen {
		return Config{}, fmt.Errorf("config: JWT_SECRET must be set and be at least %d characters", minJWTSecretLen)
	}

	rateLimitRPS, err := envFloat("RATE_LIMIT_RPS", 10)
	if err != nil {
		return Config{}, err
	}
	rateLimitBurst, err := envInt("RATE_LIMIT_BURST", 20)
	if err != nil {
		return Config{}, err
	}
	rateLimitTTL, err := envDuration("RATE_LIMIT_TTL", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPPort:          port,
		HTTPReadTimeout:   defaultHTTPReadTimeout,
		HTTPWriteTimeout:  defaultHTTPWriteTimeout,
		ShutdownTimeout:   defaultShutdownTimeout,
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            dbPort,
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres67"),
		DBName:            getEnv("DB_NAME", "user_wallet_servicedb"),
		DBMaxOpenConns:    dbMaxOpenConns,
		DBMaxIdleConns:    dbMaxIdleConns,
		DBConnMaxLifetime: dbConnMaxLifetime,
		JWTSecret:         jwtSecret,
		JWTAccessTTL:      jwtAccessTTL,
		JWTRefreshTTL:     jwtRefreshTTL,
		RateLimitRPS:      rateLimitRPS,
		RateLimitBurst:    rateLimitBurst,
		RateLimitTTL:      rateLimitTTL,
	}, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer, got %q", key, raw)
	}
	if v <= 0 {
		return 0, fmt.Errorf("config: %s must be positive, got %d", key, v)
	}
	return v, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a valid duration, got %q", key, raw)
	}
	return d, nil
}

func envFloat(key string, fallback float64) (float64, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a float, got %q", key, raw)
	}
	if f <= 0 {
		return 0, fmt.Errorf("config: %s must be positive, got %f", key, f)
	}
	return f, nil
}
