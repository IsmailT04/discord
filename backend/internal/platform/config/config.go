package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env     string `envconfig:"APP_ENV" default:"development"`
	Port    int    `envconfig:"APP_PORT" default:"8080"`
	BaseURL string `envconfig:"APP_BASE_URL" default:"http://localhost:8080"`
	ServiceVersion string `envconfig:"SERVICE_VERSION" default:"0.0.1"`
	
	Database struct {
		Host            string        `envconfig:"DB_HOST" default:"localhost"`
		Port            int           `envconfig:"DB_PORT" default:"5432"`
		User            string        `envconfig:"DB_USER" required:"true"`
		Password        string        `envconfig:"DB_PASSWORD" required:"true"`
		Name            string        `envconfig:"DB_NAME" required:"true"`
		SSLMode         string        `envconfig:"DB_SSLMODE" default:"disable"`
		MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
		MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
		ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"15m"`
	}

	Redis struct {
		Host     string `envconfig:"REDIS_HOST" default:"localhost"`
		Port     int    `envconfig:"REDIS_PORT" default:"6379"`
		Password string `envconfig:"REDIS_PASSWORD"` // Removed required:"true" for smoother local dev
		DB       int    `envconfig:"REDIS_DB" default:"0"`
	}

	Auth struct {
		CookieSecret    string        `envconfig:"COOKIE_SECRET" required:"true"`
		CookieDomain    string        `envconfig:"COOKIE_DOMAIN" default:"localhost"`
		CookieSecure    bool          `envconfig:"COOKIE_SECURE" default:"false"`
		AccessTokenTTL  time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:"15m"`
		RefreshTokenTTL time.Duration `envconfig:"REFRESH_TOKEN_TTL" default:"168h"`
		CSRFSecret      string        `envconfig:"CSRF_SECRET" required:"true"`
	}

	Chat struct {
		MaxAttachmentSizeBytes int      `envconfig:"MAX_ATTACHMENT_SIZE_BYTES" default:"5242880"`
		AllowedAttachmentTypes []string `envconfig:"ALLOWED_ATTACHMENT_TYPES" default:"image/jpeg,image/png,image/gif,image/webp,application/pdf"`
		// Note: envconfig automatically splits comma-separated strings into []string slices!
	}

	LiveKit struct {
		Host      string `envconfig:"LIVEKIT_HOST" default:"ws://localhost:7880"`
		APIKey    string `envconfig:"LIVEKIT_API_KEY" required:"true"`
		APISecret string `envconfig:"LIVEKIT_API_SECRET" required:"true"`
	}

	Observability struct {
		OTELServiceName          string `envconfig:"OTEL_SERVICE_NAME" default:"discord-backend"`
		OTELExporterOTLPEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4318"`
		OTELTraceSampler         string `envconfig:"OTEL_TRACES_SAMPLER" default:"always_on"` // always_on | always_off
	}
}

// Validate performs business invariant checks on the config
func (c *Config) Validate() error {
	// 1. Environment Check
	switch c.Env {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("APP_ENV must be one of [development, staging, production], got: %q", c.Env)
	}

	// 2. Port Validation
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("APP_PORT must be between 1 and 65535, got: %d", c.Port)
	}

	// 3. URL Validation
	if _, err := url.ParseRequestURI(c.BaseURL); err != nil {
		return fmt.Errorf("APP_BASE_URL is invalid: %w", err)
	}

	// 4. Database Connection Pool Rules
	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be greater than 0")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS (%d) cannot be greater than DB_MAX_OPEN_CONNS (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}

	// 5. Auth Security Invariants
	if len(c.Auth.CookieSecret) < 32 {
		return fmt.Errorf("COOKIE_SECRET must be at least 32 bytes long for security")
	}
	if len(c.Auth.CSRFSecret) < 32 {
		return fmt.Errorf("CSRF_SECRET must be at least 32 bytes long for security")
	}
	if c.Env == "production" && !c.Auth.CookieSecure {
		return fmt.Errorf("COOKIE_SECURE must be set to 'true' when APP_ENV is 'production'")
	}

	// 6. Attachment Rules
	if c.Chat.MaxAttachmentSizeBytes <= 0 {
		return fmt.Errorf("MAX_ATTACHMENT_SIZE_BYTES must be greater than 0")
	}
	if len(c.Chat.AllowedAttachmentTypes) == 0 {
		return fmt.Errorf("ALLOWED_ATTACHMENT_TYPES cannot be empty")
	}

	return nil
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process env vars: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
