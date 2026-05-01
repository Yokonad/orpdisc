package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/Yokonad/orpdisc/internal/models"
)

const (
	// Discord API base URL
	discordAPIURL = "https://discord.com/api/webhooks"

	// MaxEmbedsPerRequest is the maximum number of embeds per webhook request
	MaxEmbedsPerRequest = 10

	// OpenRouterBaseURL is the base URL for model links
	OpenRouterBaseURL = "https://openrouter.ai/models/"

	// ColorGreen is used for new model notifications (#2ECC71)
	ColorGreen = 3066993

	// ColorYellow is used for price/context change notifications (#FFFF00)
	ColorYellow = 16776960

	// ColorRed is used for removed model notifications (#E74C3C)
	ColorRed = 15158332

	// ColorBlue is used for digest/ranking notifications (#3498DB)
	ColorBlue = 3447003
)

// WebhookClient handles sending notifications to Discord webhooks
type WebhookClient struct {
	httpClient *http.Client
	webhookURL string
	maxRetries int
}

// DiscordWebhookPayload represents the Discord webhook payload structure
type DiscordWebhookPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

// DiscordEmbed represents a Discord embed object
type DiscordEmbed struct {
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Color       int               `json:"color,omitempty"`
	Timestamp   string            `json:"timestamp,omitempty"`
	Fields      []DiscordField    `json:"fields,omitempty"`
	Footer      *DiscordFooter    `json:"footer,omitempty"`
	URL         string            `json:"url,omitempty"`
}

// DiscordField represents a field in a Discord embed
type DiscordField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordFooter represents the footer of a Discord embed
type DiscordFooter struct {
	Text string `json:"text,omitempty"`
}

// NewWebhookClient creates a new Discord webhook client
func NewWebhookClient(webhookURL string, timeout time.Duration, maxRetries int) (*WebhookClient, error) {
	// Validate webhook URL is not empty
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook URL cannot be empty")
	}
	// Validate webhook URL format
	if _, err := url.Parse(webhookURL); err != nil {
		return nil, fmt.Errorf("invalid webhook URL: %w", err)
	}

	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		webhookURL: webhookURL,
		maxRetries: maxRetries,
	}, nil
}

// SendNotification sends a changeset as a Discord notification
// It batches embeds up to MaxEmbedsPerRequest per request
func (c *WebhookClient) SendNotification(ctx context.Context, changeset *models.Changeset) error {
	if changeset == nil || !changeset.HasChanges() {
		return nil
	}

	// Build all embeds from the changeset
	embeds := c.BuildEmbedsForChangeset(changeset)

	if len(embeds) == 0 {
		return nil
	}

	// Send in batches
	for i := 0; i < len(embeds); i += MaxEmbedsPerRequest {
		end := i + MaxEmbedsPerRequest
		if end > len(embeds) {
			end = len(embeds)
		}

		batch := embeds[i:end]
		payload := DiscordWebhookPayload{Embeds: batch}

		if err := c.sendWithRetry(ctx, payload); err != nil {
			return fmt.Errorf("failed to send batch %d-%d: %w", i+1, end, err)
		}
	}

	return nil
}

// sendWithRetry sends the payload with exponential backoff retry
func (c *WebhookClient) sendWithRetry(ctx context.Context, payload DiscordWebhookPayload) error {
	var lastErr error

	operation := func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "OpenRouter-Discord-Monitor/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			return err
		}
		defer resp.Body.Close()

		// Handle rate limiting (429)
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			var retryDuration time.Duration

			if retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					retryDuration = time.Duration(seconds) * time.Second
				}
			}

			if retryDuration == 0 {
				retryDuration = 5 * time.Second // default fallback
			}

			lastErr = fmt.Errorf("rate limited by Discord, retry after: %s", retryDuration)
			return &RateLimitError{RetryAfter: retryDuration}
		}

		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("Discord API error %d: %s", resp.StatusCode, string(respBody))
			return fmt.Errorf("Discord returned status %d", resp.StatusCode)
		}

		return nil
	}

	backoffCfg := backoff.NewExponentialBackOff()
	backoffCfg.InitialInterval = 1 * time.Second
	backoffCfg.RandomizationFactor = 0.1
	backoffCfg.Multiplier = 2.0
	backoffCfg.MaxInterval = 30 * time.Second
	backoffCfg.MaxElapsedTime = 2 * time.Minute

	if err := backoff.Retry(operation, backoff.WithMaxRetries(backoffCfg, uint64(c.maxRetries))); err != nil {
		return fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
	}

	return nil
}

// BuildEmbedsForChangeset converts a Changeset into Discord embeds
func (c *WebhookClient) BuildEmbedsForChangeset(changeset *models.Changeset) []DiscordEmbed {
	var embeds []DiscordEmbed

	timestamp := time.Now().Format(time.RFC3339)

	// Handle digest separately - uses blue color and has special formatting
	if changeset.IsDigest {
		var fields []DiscordField

		if len(changeset.NewModels) > 0 {
			m := changeset.NewModels[0]
			cost := m.CostPer1KTokens()
			value := fmt.Sprintf("[%s](%s%s)\n$%.6f/1K tokens\n%d context\nRatio: %.2f",
				m.Name, OpenRouterBaseURL, m.ID, cost, m.ContextLength, m.ContextCostRatio())
			fields = append(fields, DiscordField{
				Name:   "Mejor por Costo",
				Value:  value,
				Inline: true,
			})
		}

		if len(changeset.UpdatedModels) > 0 {
			m := changeset.UpdatedModels[0]
			cost := m.CostPer1KTokens()
			value := fmt.Sprintf("[%s](%s%s)\n$%.6f/1K tokens\n%d context\nRatio: %.2f",
				m.Name, OpenRouterBaseURL, m.ID, cost, m.ContextLength, m.ContextCostRatio())
			fields = append(fields, DiscordField{
				Name:   "Mejor Relacion Contexto/Costo",
				Value:  value,
				Inline: true,
			})
		}

		if len(changeset.RemovedModels) > 0 {
			m := changeset.RemovedModels[0]
			cost := m.CostPer1KTokens()
			value := fmt.Sprintf("[%s](%s%s)\n$%.6f/1K tokens\n%d context\nMax salida: %d",
				m.Name, OpenRouterBaseURL, m.ID, cost, m.ContextLength, m.MaxCompletionTokens)
			fields = append(fields, DiscordField{
				Name:   "Mas Capaz (Mayor Contexto)",
				Value:  value,
				Inline: true,
			})
		}

		if len(changeset.DigestNewest) > 0 {
			m := changeset.DigestNewest[0]
			cost := m.CostPer1KTokens()
			value := fmt.Sprintf("[%s](%s%s)\n$%.6f/1K tokens\n%d context\nVisto: %s",
				m.Name, OpenRouterBaseURL, m.ID, cost, m.ContextLength,
				m.FirstSeen.Format("02/01/2006"))
			fields = append(fields, DiscordField{
				Name:   "Modelo Mas Nuevo",
				Value:  value,
				Inline: true,
			})
		}

		embed := DiscordEmbed{
			Title:       "Resumen de Modelos",
			Description: "Mejores modelos del momento",
			Color:       ColorBlue,
			Timestamp:   timestamp,
			Fields:      fields,
			Footer:      &DiscordFooter{Text: "Monitor de OpenRouter"},
		}
		embeds = append(embeds, embed)
		return embeds
	}

	// Build new models section
	if len(changeset.NewModels) > 0 {
		var modelLines []string
		maxDisplay := 10
		displayCount := len(changeset.NewModels)
		if displayCount > maxDisplay {
			displayCount = maxDisplay
		}
		for i := 0; i < displayCount; i++ {
			m := changeset.NewModels[i]
			modelLines = append(modelLines, fmt.Sprintf("• [%s](%s%s) — $%.6f/1K tokens, %d context", m.Name, OpenRouterBaseURL, m.ID, m.CostPer1KTokens(), m.ContextLength))
		}
		if len(changeset.NewModels) > maxDisplay {
			modelLines = append(modelLines, fmt.Sprintf("... y %d modelos mas", len(changeset.NewModels)-maxDisplay))
		}

		embed := DiscordEmbed{
			Title:       "Nuevos Modelos Descubiertos",
			Description: fmt.Sprintf("%d nuevo(s) modelo(s) detectado(s)", len(changeset.NewModels)),
			Color:       ColorGreen,
			Timestamp:  timestamp,
			Fields: []DiscordField{
				{
					Name:   "Nuevos Modelos",
					Value:  joinLines(modelLines),
					Inline: false,
				},
			},
			Footer: &DiscordFooter{Text: "Monitor de OpenRouter"},
		}
		embeds = append(embeds, embed)
	}

	// Build price/context changes section
	if len(changeset.UpdatedModels) > 0 {
		var modelLines []string
		maxDisplay := 10
		displayCount := len(changeset.UpdatedModels)
		if displayCount > maxDisplay {
			displayCount = maxDisplay
		}
		for i := 0; i < displayCount; i++ {
			m := changeset.UpdatedModels[i]
			modelLines = append(modelLines, fmt.Sprintf("• [%s](%s%s) — $%.6f/1K tokens, %d context", m.Name, OpenRouterBaseURL, m.ID, m.CostPer1KTokens(), m.ContextLength))
		}
		if len(changeset.UpdatedModels) > maxDisplay {
			modelLines = append(modelLines, fmt.Sprintf("... y %d modelos mas", len(changeset.UpdatedModels)-maxDisplay))
		}

		embed := DiscordEmbed{
			Title:       "Actualizaciones de Modelos",
			Description: fmt.Sprintf("%d modelo(s) con precio o contexto actualizado(s)", len(changeset.UpdatedModels)),
			Color:       ColorYellow,
			Timestamp:  timestamp,
			Fields: []DiscordField{
				{
					Name:   "Cambios de Precio/Contexto",
					Value:  joinLines(modelLines),
					Inline: false,
				},
			},
			Footer: &DiscordFooter{Text: "Monitor de OpenRouter"},
		}
		embeds = append(embeds, embed)
	}

	// Build removed models section
	if len(changeset.RemovedModels) > 0 {
		var modelLines []string
		maxDisplay := 10
		displayCount := len(changeset.RemovedModels)
		if displayCount > maxDisplay {
			displayCount = maxDisplay
		}
		for i := 0; i < displayCount; i++ {
			m := changeset.RemovedModels[i]
			modelLines = append(modelLines, fmt.Sprintf("• %s — $%.6f/1K tokens, %d context", m.Name, m.CostPer1KTokens(), m.ContextLength))
		}
		if len(changeset.RemovedModels) > maxDisplay {
			modelLines = append(modelLines, fmt.Sprintf("... y %d modelos mas", len(changeset.RemovedModels)-maxDisplay))
		}

		embed := DiscordEmbed{
			Title:       "Modelos Ya No Disponibles",
			Description: fmt.Sprintf("%d modelo(s) eliminado(s) de OpenRouter", len(changeset.RemovedModels)),
			Color:       ColorRed,
			Timestamp:  timestamp,
			Fields: []DiscordField{
				{
					Name:   "Modelos Eliminados",
					Value:  joinLines(modelLines),
					Inline: false,
				},
			},
			Footer: &DiscordFooter{Text: "Monitor de OpenRouter"},
		}
		embeds = append(embeds, embed)
	}

	return embeds
}

// RateLimitError represents a Discord rate limit error
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after: %s", e.RetryAfter)
}

// joinLines joins strings with newlines, limiting to avoid Discord embed limits
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 && i%10 == 0 {
			result += "\n" // Add break every 10 lines to avoid embed limits
		}
		result += line + "\n"
	}
	return result
}
