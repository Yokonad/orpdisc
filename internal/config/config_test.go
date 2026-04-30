package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original env and restore after test
	origEnv := map[string]string{
		"DISCORD_WEBHOOK_URL":             "https://discord.com/api/webhooks/123/abc",
		"POLL_INTERVAL_MINUTES":           "15m",
		"DB_PATH":                          "./test.db",
		"LOG_LEVEL":                        "debug",
		"HTTP_TIMEOUT_SECONDS":             "60s",
		"OPENROUTER_BASE_URL":              "https://custom.api/v1",
		"MAX_RETRIES":                      "3",
		"CIRCUIT_BREAKER_THRESHOLD":        "10",
		"CIRCUIT_BREAKER_TIMEOUT_MINUTES":  "30m",
	}
	defer func() {
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	for k, v := range origEnv {
		os.Setenv(k, v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DiscordWebhookURL != origEnv["DISCORD_WEBHOOK_URL"] {
		t.Errorf("DiscordWebhookURL = %s, want %s", cfg.DiscordWebhookURL, origEnv["DISCORD_WEBHOOK_URL"])
	}

	if cfg.PollInterval != 15*time.Minute {
		t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, 15*time.Minute)
	}

	if cfg.DatabasePath != origEnv["DB_PATH"] {
		t.Errorf("DatabasePath = %s, want %s", cfg.DatabasePath, origEnv["DB_PATH"])
	}

	if cfg.LogLevel != origEnv["LOG_LEVEL"] {
		t.Errorf("LogLevel = %s, want %s", cfg.LogLevel, origEnv["LOG_LEVEL"])
	}

	if cfg.HTTPTimeout != 60*time.Second {
		t.Errorf("HTTPTimeout = %v, want %v", cfg.HTTPTimeout, 60*time.Second)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, 3)
	}

	if cfg.CircuitBreakerThreshold != 10 {
		t.Errorf("CircuitBreakerThreshold = %d, want %d", cfg.CircuitBreakerThreshold, 10)
	}

	if cfg.CircuitBreakerTimeout != 30*time.Minute {
		t.Errorf("CircuitBreakerTimeout = %v, want %v", cfg.CircuitBreakerTimeout, 30*time.Minute)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Unset all optional env vars
	os.Unsetenv("POLL_INTERVAL_MINUTES")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("HTTP_TIMEOUT_SECONDS")
	os.Unsetenv("OPENROUTER_BASE_URL")
	os.Unsetenv("MAX_RETRIES")
	os.Unsetenv("CIRCUIT_BREAKER_THRESHOLD")
	os.Unsetenv("CIRCUIT_BREAKER_TIMEOUT_MINUTES")

	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/123/abc")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.PollInterval != 30*time.Minute {
		t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, 30*time.Minute)
	}

	if cfg.DatabasePath != "./data.db" {
		t.Errorf("DatabasePath = %s, want ./data.db", cfg.DatabasePath)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %s, want info", cfg.LogLevel)
	}

	if cfg.HTTPTimeout != 30*time.Second {
		t.Errorf("HTTPTimeout = %v, want %v", cfg.HTTPTimeout, 30*time.Second)
	}

	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, 5)
	}

	if cfg.CircuitBreakerThreshold != 5 {
		t.Errorf("CircuitBreakerThreshold = %d, want %d", cfg.CircuitBreakerThreshold, 5)
	}

	if cfg.CircuitBreakerTimeout != 60*time.Minute {
		t.Errorf("CircuitBreakerTimeout = %v, want %v", cfg.CircuitBreakerTimeout, 60*time.Minute)
	}
}

func TestLoadRequiredEnv(t *testing.T) {
	os.Unsetenv("DISCORD_WEBHOOK_URL")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for missing DISCORD_WEBHOOK_URL, got nil")
	}
}

func TestLoadInvalidWebhookURL(t *testing.T) {
	os.Setenv("DISCORD_WEBHOOK_URL", "not-a-valid-url")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid webhook URL, got nil")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/123/abc")
	os.Setenv("LOG_LEVEL", "invalid")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid log level, got nil")
	}
}

func TestLoadInvalidPollInterval(t *testing.T) {
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/123/abc")
	os.Setenv("POLL_INTERVAL_MINUTES", "-5")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for negative poll interval, got nil")
	}
}

func TestRedactedWebhookURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		contains string
	}{
		{
			name:     "valid webhook",
			url:      "https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOhy0elTv2ZcwNe4uHKqNJj8nd",
			contains: "...",
		},
		{
			name:     "invalid url (not a valid absolute URL)",
			url:      "not-a-url",
			contains: "not-a-ur", // url.Parse parses it successfully as a path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{DiscordWebhookURL: tt.url}
			result := c.RedactedWebhookURL()

			// Should contain the expected marker or indicator
			if !containsSubstring(result, tt.contains) {
				t.Errorf("RedactedWebhookURL() = %s, want to contain %s", result, tt.contains)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
