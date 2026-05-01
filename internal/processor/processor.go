package processor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Yokonad/orpdisc/internal/database"
	"github.com/Yokonad/orpdisc/internal/models"
)

// Processor handles model change detection and metric calculations
type Processor struct {
	db *database.DB
}

// NewProcessor creates a new Processor instance
func NewProcessor(db *database.DB) *Processor {
	return &Processor{db: db}
}

// CalculateModelHash generates a consistent hash for a model based on its key attributes
// This hash is used for change detection
func CalculateModelHash(model *models.Model) string {
	// Create a normalized representation of the model for hashing
	hashData := struct {
		ID                 string  `json:"id"`
		ContextLength      int     `json:"context_length"`
		PricingPrompt      float64 `json:"pricing_prompt"`
		PricingCompletion   float64 `json:"pricing_completion"`
		MaxCompletionTokens int    `json:"max_completion_tokens"`
		Tokenizer          string  `json:"tokenizer"`
	}{
		ID:                  model.ID,
		ContextLength:       model.ContextLength,
		PricingPrompt:       model.PricingPrompt,
		PricingCompletion:   model.PricingCompletion,
		MaxCompletionTokens: model.MaxCompletionTokens,
		Tokenizer:           model.Tokenizer,
	}

	data, err := json.Marshal(hashData)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// ProcessModels compares incoming models against stored models and returns a changeset
func (p *Processor) ProcessModels(ctx context.Context, apiModels []models.APIModel) (*models.Changeset, error) {
	// Convert API models to domain models with calculated hashes
	incomingModels := make([]models.Model, 0, len(apiModels))
	for _, apiModel := range apiModels {
		model := apiModel.ToModel()
		model.DataHash = CalculateModelHash(model)
		incomingModels = append(incomingModels, *model)
	}

	// Get stored models from database
	storedModels, err := p.db.GetAllModels()
	if err != nil {
		return nil, fmt.Errorf("failed to get stored models: %w", err)
	}

	// Build maps for efficient lookup
	storedByID := make(map[string]*models.Model)
	for i := range storedModels {
		storedByID[storedModels[i].ID] = &storedModels[i]
	}

	incomingByID := make(map[string]*models.Model)
	for i := range incomingModels {
		incomingByID[incomingModels[i].ID] = &incomingModels[i]
	}

	changeset := &models.Changeset{
		NewModels:     []models.Model{},
		UpdatedModels: []models.Model{},
		RemovedModels: []models.Model{},
	}

	// Detect new and updated models
	for id, incoming := range incomingByID {
		stored, exists := storedByID[id]
		if !exists {
			// New model
			changeset.NewModels = append(changeset.NewModels, *incoming)
		} else if incoming.DataHash != stored.DataHash {
			// Updated model - check what changed
			changeset.UpdatedModels = append(changeset.UpdatedModels, *incoming)

			// Save price history if pricing changed
			if incoming.PricingPrompt != stored.PricingPrompt || incoming.PricingCompletion != stored.PricingCompletion {
				if err := p.db.SavePriceHistory(id, stored.PricingPrompt, stored.PricingCompletion); err != nil {
					// Log but don't fail - price history is supplementary
					fmt.Printf("failed to save price history for %s: %v\n", id, err)
				}
			}
		}
	}

	// Detect removed models
	for id, stored := range storedByID {
		if _, exists := incomingByID[id]; !exists {
			changeset.RemovedModels = append(changeset.RemovedModels, *stored)
		}
	}

	// Sort for consistent ordering
	sort.Slice(changeset.NewModels, func(i, j int) bool {
		return changeset.NewModels[i].ID < changeset.NewModels[j].ID
	})
	sort.Slice(changeset.UpdatedModels, func(i, j int) bool {
		return changeset.UpdatedModels[i].ID < changeset.UpdatedModels[j].ID
	})
	sort.Slice(changeset.RemovedModels, func(i, j int) bool {
		return changeset.RemovedModels[i].ID < changeset.RemovedModels[j].ID
	})

	// Save all incoming models to database
	for _, model := range incomingModels {
		if err := p.db.SaveModel(&model); err != nil {
			return nil, fmt.Errorf("failed to save model %s: %w", model.ID, err)
		}
	}

	return changeset, nil
}

// CalculateCostPer1KTokens calculates the cost per 1K tokens
// According to spec: cost_per_1k_tokens = (pricing.prompt + pricing.completion) * 1000
func CalculateCostPer1KTokens(pricingPrompt, pricingCompletion float64) float64 {
	return (pricingPrompt + pricingCompletion) * 1000
}

// CalculateContextCostRatio calculates the context length to cost ratio
// According to spec: context_cost_ratio = context_length / cost_per_1k_tokens
func CalculateContextCostRatio(contextLength int, costPer1KTokens float64) float64 {
	if costPer1KTokens == 0 {
		return 0
	}
	return float64(contextLength) / costPer1KTokens
}

// FormatPriceChange formats a price change for display
func FormatPriceChange(oldPrompt, oldCompletion, newPrompt, newCompletion float64) string {
	return fmt.Sprintf("$%.6f → $%.6f per 1K tokens (prompt: $%.6f → $%.6f, completion: $%.6f → $%.6f)",
		(oldPrompt+oldCompletion)*1000, (newPrompt+newCompletion)*1000,
		oldPrompt, newPrompt, oldCompletion, newCompletion)
}

// GetModelDetails returns a formatted string with model details
func GetModelDetails(model *models.Model) string {
	costPer1K := CalculateCostPer1KTokens(model.PricingPrompt, model.PricingCompletion)
	ctxCostRatio := CalculateContextCostRatio(model.ContextLength, costPer1K)

	var details []string
	details = append(details, fmt.Sprintf("Provider: %s", model.Provider))
	details = append(details, fmt.Sprintf("Context: %d tokens", model.ContextLength))
	details = append(details, fmt.Sprintf("Cost: $%.6f per 1K tokens", costPer1K))
	details = append(details, fmt.Sprintf("Context/Cost Ratio: %.2f", ctxCostRatio))

	if model.MaxCompletionTokens > 0 {
		details = append(details, fmt.Sprintf("Max Output: %d tokens", model.MaxCompletionTokens))
	}

	if model.Tokenizer != "" {
		details = append(details, fmt.Sprintf("Tokenizer: %s", model.Tokenizer))
	}

	return strings.Join(details, " | ")
}

// TopByContextLength returns the top N models sorted by largest context length
func TopByContextLength(modelList []models.Model, n int) []models.Model {
	sorted := make([]models.Model, len(modelList))
	copy(sorted, modelList)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ContextLength > sorted[j].ContextLength
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// TopByNewest returns the top N models sorted by newest first_seen
func TopByNewest(modelList []models.Model, n int) []models.Model {
	sorted := make([]models.Model, len(modelList))
	copy(sorted, modelList)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FirstSeen.After(sorted[j].FirstSeen)
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// TopByCostPer1K returns the top N models sorted by lowest cost per 1K tokens
func TopByCostPer1K(modelList []models.Model, n int) []models.Model {
	sorted := make([]models.Model, len(modelList))
	copy(sorted, modelList)

	sort.Slice(sorted, func(i, j int) bool {
		costI := CalculateCostPer1KTokens(sorted[i].PricingPrompt, sorted[i].PricingCompletion)
		costJ := CalculateCostPer1KTokens(sorted[j].PricingPrompt, sorted[j].PricingCompletion)
		return costI < costJ
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// TopByContextCostRatio returns the top N models sorted by highest context/cost ratio
func TopByContextCostRatio(modelList []models.Model, n int) []models.Model {
	sorted := make([]models.Model, len(modelList))
	copy(sorted, modelList)

	sort.Slice(sorted, func(i, j int) bool {
		costI := CalculateCostPer1KTokens(sorted[i].PricingPrompt, sorted[i].PricingCompletion)
		costJ := CalculateCostPer1KTokens(sorted[j].PricingPrompt, sorted[j].PricingCompletion)
		ratioI := CalculateContextCostRatio(sorted[i].ContextLength, costI)
		ratioJ := CalculateContextCostRatio(sorted[j].ContextLength, costJ)
		return ratioI > ratioJ
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}
