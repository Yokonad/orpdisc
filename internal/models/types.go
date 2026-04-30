package models

import (
	"strconv"
	"strings"
	"time"
)

// Model represents a machine learning model from OpenRouter
type Model struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Provider           string    `json:"provider"` // Extracted from ID (e.g., "google" from "google/gemini-2.5-pro")
	Description        string    `json:"description,omitempty"`
	ContextLength      int       `json:"context_length,omitempty"`
	MaxCompletionTokens int       `json:"max_completion_tokens,omitempty"`
	PricingPrompt      float64   `json:"pricing_prompt,omitempty"`
	PricingCompletion  float64   `json:"pricing_completion,omitempty"`
	Tokenizer          string    `json:"tokenizer,omitempty"`
	DataHash           string    `json:"data_hash,omitempty"`
	FirstSeen          time.Time `json:"first_seen,omitempty"`
	LastUpdated        time.Time `json:"last_updated,omitempty"`
}

// CostPer1KTokens calculates the cost per 1K tokens
func (m *Model) CostPer1KTokens() float64 {
	return (m.PricingPrompt + m.PricingCompletion) * 1000
}

// ContextCostRatio calculates the context length to cost ratio
func (m *Model) ContextCostRatio() float64 {
	cost := m.CostPer1KTokens()
	if cost == 0 {
		return 0
	}
	return float64(m.ContextLength) / cost
}

// PriceHistory records historical pricing for a model
type PriceHistory struct {
	ID               int64     `json:"id"`
	ModelID          string    `json:"model_id"`
	PricingPrompt    float64   `json:"pricing_prompt"`
	PricingCompletion float64  `json:"pricing_completion"`
	RecordedAt       time.Time `json:"recorded_at"`
}

// Notification represents a Discord notification that was sent
type Notification struct {
	ID       int64     `json:"id"`
	Type     string    `json:"type"` // "new", "price_change", "removed", "digest"
	ModelID  string    `json:"model_id,omitempty"`
	OldValue string    `json:"old_value,omitempty"`
	NewValue string    `json:"new_value,omitempty"`
	SentAt   time.Time `json:"sent_at,omitempty"`
	Success  bool      `json:"success"`
}

// NotificationType constants
const (
	NotificationTypeNew         = "new"
	NotificationTypePriceChange = "price_change"
	NotificationTypeRemoved     = "removed"
	NotificationTypeDigest      = "digest"
)

// Changeset represents detected changes between two model snapshots
type Changeset struct {
	NewModels     []Model `json:"new_models"`
	UpdatedModels []Model `json:"updated_models"`
	RemovedModels []Model `json:"removed_models"`
	IsDigest      bool    `json:"is_digest"`
}

// HasChanges returns true if there are any changes in the changeset
func (c *Changeset) HasChanges() bool {
	return len(c.NewModels) > 0 || len(c.UpdatedModels) > 0 || len(c.RemovedModels) > 0
}

// TotalChanges returns the total number of changes
func (c *Changeset) TotalChanges() int {
	return len(c.NewModels) + len(c.UpdatedModels) + len(c.RemovedModels)
}

// --- OpenRouter API Response Types ---

// APIResponse represents the OpenRouter API response structure
type APIResponse struct {
	Models []APIModel `json:"data"`
}

// APIModel represents a model in the OpenRouter API response
type APIModel struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	ContextLength    int      `json:"context_length,omitempty"`
	PromptModalities []string `json:"prompt_modalities,omitempty"`
	OutputModalities []string `json:"output_modalities,omitempty"`
	Pricing          Pricing  `json:"pricing,omitempty"`
	TopProvider      Provider `json:"top_provider,omitempty"`
	Architecture     Arch     `json:"architecture,omitempty"`
}

// Pricing represents pricing information from OpenRouter API
type Pricing struct {
	Prompt     string `json:"prompt,omitempty"`
	Completion string `json:"completion,omitempty"`
}

// Provider represents the top provider info
type Provider struct {
	MaxCompletionTokens int `json:"max_completion_tokens,omitempty"`
}

// Arch represents architecture info including tokenizer
type Arch struct {
	Tokenizer string `json:"tokenizer,omitempty"`
}

// ToModel converts an APIModel to a Model
func (am *APIModel) ToModel() *Model {
	// Extract provider by splitting on "/" (e.g., "google/gemini-2.5-pro" -> "google")
	provider := ""
	if idx := strings.Index(am.ID, "/"); idx > 0 {
		provider = am.ID[:idx]
	}

	return &Model{
		ID:                 am.ID,
		Name:               am.Name,
		Provider:           provider,
		Description:        am.Description,
		ContextLength:      am.ContextLength,
		MaxCompletionTokens: am.TopProvider.MaxCompletionTokens,
		PricingPrompt:      parsePrice(am.Pricing.Prompt),
		PricingCompletion:  parsePrice(am.Pricing.Completion),
		Tokenizer:          am.Architecture.Tokenizer,
	}
}

// IsTextModel returns true if the model supports text output
func (am *APIModel) IsTextModel() bool {
	for _, mod := range am.OutputModalities {
		if mod == "text" {
			return true
		}
	}
	return false
}

// parsePrice converts a price string (e.g., "0.000015") to float64
func parsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	price, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return price
}
