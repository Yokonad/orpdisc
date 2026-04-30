package discord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yokonad/orpdisc/internal/models"
)

func TestNewWebhookClient(t *testing.T) {
	validURL := "https://discord.com/api/webhooks/123456/abcdef"

	client, err := NewWebhookClient(validURL, 30*time.Second, 5)
	if err != nil {
		t.Fatalf("NewWebhookClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewWebhookClient() returned nil client")
	}

	if client.webhookURL != validURL {
		t.Errorf("webhookURL = %s, want %s", client.webhookURL, validURL)
	}
}

func TestNewWebhookClientInvalidURL(t *testing.T) {
	// Empty URL will fail parsing
	invalidURL := ""

	_, err := NewWebhookClient(invalidURL, 30*time.Second, 5)
	if err == nil {
		t.Error("NewWebhookClient() expected error for empty URL, got nil")
	}
}

func TestSendNotificationNoChanges(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{}
	err := client.SendNotification(context.Background(), changeset)
	if err != nil {
		t.Errorf("SendNotification() error = %v", err)
	}
}

func TestSendNotificationNilChangeset(t *testing.T) {
	client := &WebhookClient{}

	err := client.SendNotification(context.Background(), nil)
	if err != nil {
		t.Errorf("SendNotification() error = %v", err)
	}
}

func TestSendNotificationSuccess(t *testing.T) {
	// Create a test server that records requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &WebhookClient{
		webhookURL: server.URL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
	}

	changeset := &models.Changeset{
		NewModels: []models.Model{
			{ID: "test/model", Name: "Test Model"},
		},
	}

	err := client.SendNotification(context.Background(), changeset)
	if err != nil {
		t.Errorf("SendNotification() error = %v", err)
	}
}

func TestBuildEmbedsForChangeset(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{
		NewModels: []models.Model{
			{ID: "google/gemini", Name: "Gemini"},
			{ID: "anthropic/claude", Name: "Claude"},
		},
		UpdatedModels: []models.Model{
			{ID: "openai/gpt", Name: "GPT"},
		},
		RemovedModels: []models.Model{
			{ID: "old/model", Name: "Old Model"},
		},
	}

	embeds := client.BuildEmbedsForChangeset(changeset)

	// Should have 3 embeds: new models, updates, removed
	if len(embeds) != 3 {
		t.Errorf("BuildEmbedsForChangeset() returned %d embeds, want 3", len(embeds))
	}

	// Check colors
	if embeds[0].Color != ColorGreen {
		t.Errorf("New models embed color = %d, want %d (green)", embeds[0].Color, ColorGreen)
	}
	if embeds[1].Color != ColorYellow {
		t.Errorf("Updated models embed color = %d, want %d (yellow)", embeds[1].Color, ColorYellow)
	}
	if embeds[2].Color != ColorRed {
		t.Errorf("Removed models embed color = %d, want %d (red)", embeds[2].Color, ColorRed)
	}
}

func TestBuildEmbedsForChangesetEmpty(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{}
	embeds := client.BuildEmbedsForChangeset(changeset)

	if len(embeds) != 0 {
		t.Errorf("BuildEmbedsForChangeset() returned %d embeds, want 0", len(embeds))
	}
}

func TestBuildEmbedsForChangesetNewOnly(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{
		NewModels: []models.Model{
			{ID: "new/model", Name: "New Model"},
		},
	}

	embeds := client.BuildEmbedsForChangeset(changeset)

	if len(embeds) != 1 {
		t.Errorf("BuildEmbedsForChangeset() returned %d embeds, want 1", len(embeds))
	}
	if embeds[0].Color != ColorGreen {
		t.Errorf("Embed color = %d, want %d (green)", embeds[0].Color, ColorGreen)
	}
}

func TestBuildEmbedsForChangesetUpdatedOnly(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{
		UpdatedModels: []models.Model{
			{ID: "updated/model", Name: "Updated Model"},
		},
	}

	embeds := client.BuildEmbedsForChangeset(changeset)

	if len(embeds) != 1 {
		t.Errorf("BuildEmbedsForChangeset() returned %d embeds, want 1", len(embeds))
	}
	if embeds[0].Color != ColorYellow {
		t.Errorf("Embed color = %d, want %d (yellow)", embeds[0].Color, ColorYellow)
	}
}

func TestBuildEmbedsForChangesetRemovedOnly(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{
		RemovedModels: []models.Model{
			{ID: "removed/model", Name: "Removed Model"},
		},
	}

	embeds := client.BuildEmbedsForChangeset(changeset)

	if len(embeds) != 1 {
		t.Errorf("BuildEmbedsForChangeset() returned %d embeds, want 1", len(embeds))
	}
	if embeds[0].Color != ColorRed {
		t.Errorf("Embed color = %d, want %d (red)", embeds[0].Color, ColorRed)
	}
}

func TestBuildEmbedsForChangesetDigest(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{
		NewModels: []models.Model{
			{ID: "google/gemini", Name: "Gemini"},
		},
		UpdatedModels: []models.Model{
			{ID: "anthropic/claude", Name: "Claude"},
		},
		IsDigest: true,
	}

	embeds := client.BuildEmbedsForChangeset(changeset)

	if len(embeds) != 1 {
		t.Errorf("BuildEmbedsForChangeset() (digest) returned %d embeds, want 1", len(embeds))
	}
	if embeds[0].Color != ColorBlue {
		t.Errorf("Digest embed color = %d, want %d (blue)", embeds[0].Color, ColorBlue)
	}
	if embeds[0].Title != "Daily Digest" {
		t.Errorf("Digest embed title = %s, want Daily Digest", embeds[0].Title)
	}
}

func TestBuildEmbedsForChangesetNoEmojis(t *testing.T) {
	client := &WebhookClient{}

	changeset := &models.Changeset{
		NewModels: []models.Model{
			{ID: "google/gemini", Name: "Gemini"},
		},
		UpdatedModels: []models.Model{
			{ID: "openai/gpt", Name: "GPT"},
		},
		RemovedModels: []models.Model{
			{ID: "old/model", Name: "Old Model"},
		},
	}

	embeds := client.BuildEmbedsForChangeset(changeset)

	// Check that no embed contains emoji characters (U+1F000 to U+1FAFF and other emoji ranges)
	emojiRanges := []struct {
		start rune
		end   rune
	}{
		{0x1F300, 0x1F9FF},  // Miscellaneous Symbols and Pictographs, Emoticons, Transport and Map Symbols, Activity and Game Icons
		{0x2600, 0x26FF},    // Miscellaneous Symbols (some emoji-like)
		{0x2700, 0x27BF},    // Dingbats (some emoji-like)
	}

	for i, embed := range embeds {
		for j, field := range embed.Fields {
			for _, r := range field.Name {
				for _, rng := range emojiRanges {
					if r >= rng.start && r <= rng.end {
						t.Errorf("Embed %d field %d contains emoji character: U+%04X", i, j, r)
					}
				}
			}
			for _, r := range field.Value {
				for _, rng := range emojiRanges {
					if r >= rng.start && r <= rng.end {
						t.Errorf("Embed %d field %d value contains emoji character: U+%04X", i, j, r)
					}
				}
			}
		}
	}
}

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{RetryAfter: 30 * time.Second}

	expected := "rate limited, retry after: 30s"
	if err.Error() != expected {
		t.Errorf("RateLimitError.Error() = %s, want %s", err.Error(), expected)
	}
}

func TestJoinLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{
			name:     "empty",
			lines:    []string{},
			expected: "",
		},
		{
			name:     "single",
			lines:    []string{"one"},
			expected: "one\n",
		},
		{
			name:     "multiple",
			lines:    []string{"one", "two", "three"},
			expected: "one\ntwo\nthree\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinLines(tt.lines)
			if result != tt.expected {
				t.Errorf("joinLines() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestJoinLinesWithBreaks(t *testing.T) {
	// Create more than 10 lines to test the break logic
	lines := make([]string, 15)
	for i := range lines {
		lines[i] = string(rune('0' + i%10))
	}

	result := joinLines(lines)

	// Should contain newlines
	if len(result) < 10 {
		t.Errorf("joinLines() seems too short: %s", result)
	}
}
