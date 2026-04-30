package processor

import (
	"testing"

	"github.com/Yokonad/orpdisc/internal/models"
)

func TestCalculateModelHash(t *testing.T) {
	model := &models.Model{
		ID:                  "test/model",
		ContextLength:       1000,
		PricingPrompt:       0.01,
		PricingCompletion:   0.02,
		MaxCompletionTokens: 500,
		Tokenizer:           "test-tokenizer",
	}

	hash1 := CalculateModelHash(model)
	if hash1 == "" {
		t.Error("CalculateModelHash() returned empty hash")
	}

	// Same model should produce same hash
	hash2 := CalculateModelHash(model)
	if hash1 != hash2 {
		t.Errorf("CalculateModelHash() not deterministic: got %s and %s", hash1, hash2)
	}

	// Different model should produce different hash
	model2 := &models.Model{
		ID:                  "test/model",
		ContextLength:       2000, // changed
		PricingPrompt:       0.01,
		PricingCompletion:   0.02,
		MaxCompletionTokens: 500,
		Tokenizer:           "test-tokenizer",
	}
	hash3 := CalculateModelHash(model2)
	if hash1 == hash3 {
		t.Errorf("CalculateModelHash() produced same hash for different models")
	}
}

func TestCalculateCostPer1KTokens(t *testing.T) {
	tests := []struct {
		name     string
		prompt   float64
		complete float64
		expected float64
	}{
		{"zero prices", 0, 0, 0},
		{"same prices", 0.000015, 0.000075, 0.09},
		{"prompt only", 0.00001, 0, 0.01},
		{"completion only", 0, 0.00005, 0.05},
		{"large values", 0.001, 0.002, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCostPer1KTokens(tt.prompt, tt.complete)
			if result != tt.expected {
				t.Errorf("CalculateCostPer1KTokens(%f, %f) = %f, want %f",
					tt.prompt, tt.complete, result, tt.expected)
			}
		})
	}
}

func TestCalculateContextCostRatio(t *testing.T) {
	tests := []struct {
		name     string
		ctxLen   int
		cost     float64
		expected float64
	}{
		{"zero cost", 1000, 0, 0},
		{"normal values", 1000, 0.09, 11111.111111111111},
		{"large context", 200000, 0.001, 200000000},
		{"small context", 100, 1.0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateContextCostRatio(tt.ctxLen, tt.cost)
			// Use approximate comparison for floating point
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.0001 {
				t.Errorf("CalculateContextCostRatio(%d, %f) = %f, want %f",
					tt.ctxLen, tt.cost, result, tt.expected)
			}
		})
	}
}

func TestFormatPriceChange(t *testing.T) {
	result := FormatPriceChange(0.01, 0.02, 0.015, 0.025)

	// Check that it contains expected format elements
	if result == "" {
		t.Error("FormatPriceChange() returned empty string")
	}

	// Check it's not completely broken
	if len(result) < 10 {
		t.Errorf("FormatPriceChange() seems too short: %s", result)
	}
}

func TestGetModelDetails(t *testing.T) {
	model := &models.Model{
		ID:                  "test/model",
		Name:                "Test Model",
		Provider:            "test",
		ContextLength:       1000,
		MaxCompletionTokens: 500,
		PricingPrompt:       0.01,
		PricingCompletion:   0.02,
		Tokenizer:           "test-tokenizer",
	}

	details := GetModelDetails(model)

	// Should contain key information
	expected := []string{
		"Provider: test",
		"Context: 1000 tokens",
		"Cost:",
		"Context/Cost Ratio:",
	}

	for _, exp := range expected {
		if len(details) < len(exp) || !containsSubstring(details, exp) {
			t.Errorf("GetModelDetails() missing expected content: %s", exp)
		}
	}
}

func TestTopByCostPer1K(t *testing.T) {
	models := []models.Model{
		{ID: "a", Name: "Expensive", PricingPrompt: 0.1, PricingCompletion: 0.1},
		{ID: "b", Name: "Cheap", PricingPrompt: 0.001, PricingCompletion: 0.001},
		{ID: "c", Name: "Medium", PricingPrompt: 0.01, PricingCompletion: 0.01},
	}

	result := TopByCostPer1K(models, 2)

	if len(result) != 2 {
		t.Errorf("TopByCostPer1K() returned %d models, want 2", len(result))
	}

	// First should be cheapest (b)
	if result[0].ID != "b" {
		t.Errorf("TopByCostPer1K()[0] = %s, want b (cheap)", result[0].ID)
	}

	// Second should be medium (c)
	if result[1].ID != "c" {
		t.Errorf("TopByCostPer1K()[1] = %s, want c (medium)", result[1].ID)
	}
}

func TestTopByContextCostRatio(t *testing.T) {
	models := []models.Model{
		{ID: "a", Name: "Small Context", ContextLength: 1000, PricingPrompt: 0.01, PricingCompletion: 0.01},
		{ID: "b", Name: "Large Context", ContextLength: 100000, PricingPrompt: 0.01, PricingCompletion: 0.01},
		{ID: "c", Name: "Medium Context", ContextLength: 10000, PricingPrompt: 0.01, PricingCompletion: 0.01},
	}

	result := TopByContextCostRatio(models, 2)

	if len(result) != 2 {
		t.Errorf("TopByContextCostRatio() returned %d models, want 2", len(result))
	}

	// First should be largest context (b)
	if result[0].ID != "b" {
		t.Errorf("TopByContextCostRatio()[0] = %s, want b (large)", result[0].ID)
	}

	// Second should be medium context (c)
	if result[1].ID != "c" {
		t.Errorf("TopByContextCostRatio()[1] = %s, want c (medium)", result[1].ID)
	}
}

func TestTopByCostPer1KWithLessModels(t *testing.T) {
	models := []models.Model{
		{ID: "a", PricingPrompt: 0.1, PricingCompletion: 0.1},
		{ID: "b", PricingPrompt: 0.001, PricingCompletion: 0.001},
	}

	// Request more than available
	result := TopByCostPer1K(models, 5)

	if len(result) != 2 {
		t.Errorf("TopByCostPer1K() returned %d models, want 2 (max available)", len(result))
	}
}

func TestTopByContextCostRatioWithLessModels(t *testing.T) {
	models := []models.Model{
		{ID: "a", ContextLength: 1000, PricingPrompt: 0.01, PricingCompletion: 0.01},
	}

	// Request more than available
	result := TopByContextCostRatio(models, 5)

	if len(result) != 1 {
		t.Errorf("TopByContextCostRatio() returned %d models, want 1 (max available)", len(result))
	}
}

func TestChangesetHasChanges(t *testing.T) {
	tests := []struct {
		name      string
		changeset models.Changeset
		expected  bool
	}{
		{
			name:      "empty changeset",
			changeset: models.Changeset{},
			expected:  false,
		},
		{
			name: "with new models",
			changeset: models.Changeset{
				NewModels: []models.Model{{ID: "a"}},
			},
			expected: true,
		},
		{
			name: "with updated models",
			changeset: models.Changeset{
				UpdatedModels: []models.Model{{ID: "b"}},
			},
			expected: true,
		},
		{
			name: "with removed models",
			changeset: models.Changeset{
				RemovedModels: []models.Model{{ID: "c"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.changeset.HasChanges()
			if result != tt.expected {
				t.Errorf("HasChanges() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChangesetTotalChanges(t *testing.T) {
	changeset := models.Changeset{
		NewModels:     []models.Model{{ID: "a"}, {ID: "b"}},
		UpdatedModels: []models.Model{{ID: "c"}},
		RemovedModels: []models.Model{{ID: "d"}, {ID: "e"}, {ID: "f"}},
	}

	result := changeset.TotalChanges()
	if result != 6 {
		t.Errorf("TotalChanges() = %d, want 6", result)
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
