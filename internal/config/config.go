package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v10"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	DiscordWebhookURL    string        `env:"DISCORD_WEBHOOK_URL,required"`
	PollInterval         time.Duration `env:"POLL_INTERVAL_MINUTES" envDefault:"4h"`
	DatabasePath         string        `env:"DB_PATH" envDefault:"./data.db"`
	LogLevel             string        `env:"LOG_LEVEL" envDefault:"info"`
	HTTPTimeout          time.Duration `env:"HTTP_TIMEOUT_SECONDS" envDefault:"30s"`
	OpenRouterBaseURL    string        `env:"OPENROUTER_BASE_URL" envDefault:"https://openrouter.ai/api/v1"`
	MaxRetries           int           `env:"MAX_RETRIES" envDefault:"5"`
	CircuitBreakerThreshold int        `env:"CIRCUIT_BREAKER_THRESHOLD" envDefault:"5"`
	CircuitBreakerTimeout   time.Duration `env:"CIRCUIT_BREAKER_TIMEOUT_MINUTES" envDefault:"60m"`
	HealthCheckPort         string        `env:"HEALTH_CHECK_PORT" envDefault:":9090"`
	ActiveStartHour        int           `env:"ACTIVE_START_HOUR" envDefault:"9"`
	ActiveEndHour          int           `env:"ACTIVE_END_HOUR" envDefault:"19"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	// Validate Discord webhook URL - must be a valid URL starting with https://
	if _, err := url.Parse(c.DiscordWebhookURL); err != nil {
		return fmt.Errorf("invalid DISCORD_WEBHOOK_URL: %w", err)
	}

	// Additional webhook URL validation - must be https
	u, _ := url.Parse(c.DiscordWebhookURL)
	if u.Scheme != "https" {
		return fmt.Errorf("invalid DISCORD_WEBHOOK_URL: must use https")
	}

	// Validate durations are positive
	if c.PollInterval <= 0 {
		return fmt.Errorf("POLL_INTERVAL_MINUTES must be positive")
	}

	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("HTTP_TIMEOUT_SECONDS must be positive")
	}

	// Validate log level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid LOG_LEVEL: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}

// RedactedWebhookURL returns the webhook URL with the token redacted for logging
func (c *Config) RedactedWebhookURL() string {
	u, err := url.Parse(c.DiscordWebhookURL)
	if err != nil {
		return "[invalid URL]"
	}
	// Keep only the first 8 chars of the token
	if len(u.Path) > 8 {
		u.Path = u.Path[:8] + "..."
	}
	return u.String()
}
