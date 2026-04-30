package openrouter

import (
	"testing"
	"time"

	"github.com/Yokonad/orpdisc/internal/config"
	"github.com/Yokonad/orpdisc/internal/models"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		HTTPTimeout:            30 * time.Second,
		OpenRouterBaseURL:      "https://openrouter.ai/api/v1",
		MaxRetries:             5,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:  60 * time.Minute,
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.baseURL != cfg.OpenRouterBaseURL {
		t.Errorf("baseURL = %s, want %s", client.baseURL, cfg.OpenRouterBaseURL)
	}

	if client.maxRetries != cfg.MaxRetries {
		t.Errorf("maxRetries = %d, want %d", client.maxRetries, cfg.MaxRetries)
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"0.000015", 0.000015},
		{"0.000075", 0.000075},
		{"", 0},
		{"invalid", 0},
		{"0.00001", 0.00001},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParsePrice(tt.input)
			if result != tt.expected {
				t.Errorf("ParsePrice(%q) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractProvider(t *testing.T) {
	tests := []struct {
		modelID  string
		expected string
	}{
		{"google/gemini-2.5-pro", "google"},
		{"anthropic/claude-opus-4.7", "anthropic"},
		{"openai/gpt-5.4", "openai"},
		{"singleword", "singleword"},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := ExtractProvider(tt.modelID)
			if result != tt.expected {
				t.Errorf("ExtractProvider(%q) = %q, want %q", tt.modelID, result, tt.expected)
			}
		})
	}
}

func TestValidateResponse(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name       string
		models     []models.APIModel
		wantErrors int
	}{
		{
			name:       "valid models",
			models:     []models.APIModel{{ID: "test/model", Name: "Test Model"}},
			wantErrors: 0,
		},
		{
			name:       "missing ID",
			models:     []models.APIModel{{ID: "", Name: "Test Model"}},
			wantErrors: 1,
		},
		{
			name:       "missing name",
			models:     []models.APIModel{{ID: "test/model", Name: ""}},
			wantErrors: 1,
		},
		{
			name:       "invalid prompt price",
			models:     []models.APIModel{{ID: "test/model", Name: "Test", Pricing: models.Pricing{Prompt: "invalid"}}},
			wantErrors: 1,
		},
		{
			name:       "invalid completion price",
			models:     []models.APIModel{{ID: "test/model", Name: "Test", Pricing: models.Pricing{Completion: "invalid"}}},
			wantErrors: 1,
		},
		{
			name:       "multiple errors",
			models:     []models.APIModel{{ID: "", Name: ""}, {ID: "test", Name: "test", Pricing: models.Pricing{Prompt: "x"}}},
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := client.ValidateResponse(tt.models)
			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateResponse() returned %d errors, want %d", len(errors), tt.wantErrors)
			}
		})
	}
}
