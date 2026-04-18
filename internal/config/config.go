package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	HTTPPort       string
	LogLevel       string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	HashIDSalt     string
	HashIDMinLen   int
	RateLimitRPS   int
	RateLimitBurst int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:      getenv("APP_ENV", "development"),
		HTTPPort:    getenv("HTTP_PORT", "8080"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		HashIDSalt:  os.Getenv("HASHID_SALT"),
	}

	var err error
	if cfg.JWTAccessTTL, err = time.ParseDuration(getenv("JWT_ACCESS_TTL", "15m")); err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	if cfg.JWTRefreshTTL, err = time.ParseDuration(getenv("JWT_REFRESH_TTL", "720h")); err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	cfg.HashIDMinLen = atoiOr(os.Getenv("HASHID_MIN_LEN"), 6)
	cfg.RateLimitRPS = atoiOr(os.Getenv("RATE_LIMIT_RPS"), 10)
	cfg.RateLimitBurst = atoiOr(os.Getenv("RATE_LIMIT_BURST"), 20)

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return errors.New("REDIS_URL is required")
	}
	if len(c.JWTSecret) < 16 {
		return errors.New("JWT_SECRET must be at least 16 chars")
	}
	if c.HashIDSalt == "" {
		return errors.New("HASHID_SALT is required")
	}
	return nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
