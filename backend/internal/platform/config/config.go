package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppPort      string
	AppEnv       string
	DatabaseURL  string
	RedisURL     string
	PublicBaseURL string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	GotenbergURL string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	SMTPFrom   string
	OrgRequestTo string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppPort:      envOr("APP_PORT", "8080"),
		AppEnv:       envOr("APP_ENV", "development"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     envOr("REDIS_URL", "redis://localhost:6379/0"),
		PublicBaseURL: envOr("PUBLIC_BASE_URL", "http://localhost:5173"),

		MinioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    envOr("MINIO_BUCKET", "jetistik"),
		MinioUseSSL:    os.Getenv("MINIO_USE_SSL") == "true",

		GotenbergURL: envOr("GOTENBERG_URL", "http://localhost:3000"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		SMTPHost:     envOr("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     envOr("SMTP_PORT", "587"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     envOr("SMTP_FROM", "Jetistik <noreply@jetistik.kz>"),
		OrgRequestTo: envOr("ORGANIZER_REQUEST_TO", "yerzhan@blackboard.kz"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	var err error
	cfg.JWTAccessTTL, err = time.ParseDuration(envOr("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWTRefreshTTL, err = time.ParseDuration(envOr("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	return cfg, nil
}

func (c *Config) IsDev() bool {
	return c.AppEnv == "development"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
