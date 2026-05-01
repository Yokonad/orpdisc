package openrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/Yokonad/orpdisc/internal/config"
	"github.com/Yokonad/orpdisc/internal/models"
)

const (
	// DefaultUserAgent is the user-agent header sent with API requests
	DefaultUserAgent = "OpenRouter-Discord-Monitor/1.0"
	// BaseURL is the OpenRouter API base URL
	BaseURL = "https://openrouter.ai/api/v1"
)

// Client wraps the HTTP client for OpenRouter API communication
type Client struct {
	httpClient       *http.Client
	baseURL          string
	userAgent        string
	maxRetries       int
	failureCount     int
	circuitOpen      bool
	circuitOpenTime  time.Time
	circuitTimeout   time.Duration
	circuitThreshold int
	mu               sync.Mutex
}

// NewClient creates a new OpenRouter API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		baseURL:          cfg.OpenRouterBaseURL,
		userAgent:        DefaultUserAgent,
		maxRetries:       cfg.MaxRetries,
		circuitTimeout:   cfg.CircuitBreakerTimeout,
		circuitThreshold: cfg.CircuitBreakerThreshold,
	}
}

// FetchModels fetches all available models from OpenRouter API
func (c *Client) FetchModels(ctx context.Context) ([]models.APIModel, error) {
	// Check circuit breaker
	if c.isCircuitOpen() {
		return nil, fmt.Errorf("circuit breaker is open, skipping request for %s", c.circuitTimeout-c.timeSinceCircuitOpen())
	}

	url := c.baseURL + "/models"

	var apiResp models.APIResponse
	var lastErr error

	operation := func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			c.recordFailure()
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				lastErr = fmt.Errorf("rate limited, retry after: %s", retryAfter)
			} else {
				lastErr = fmt.Errorf("rate limited (429)")
			}
			c.recordFailure()
			return fmt.Errorf("rate limited")
		}

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
			c.recordFailure()
			return fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			c.recordFailure()
			return err
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			lastErr = fmt.Errorf("failed to parse JSON response: %w", err)
			c.recordFailure()
			return err
		}

		return nil
	}

	backoffCfg := backoff.NewExponentialBackOff()
	backoffCfg.InitialInterval = 2 * time.Second
	backoffCfg.RandomizationFactor = 0.1
	backoffCfg.Multiplier = 2.0
	backoffCfg.MaxInterval = 16 * time.Second
	backoffCfg.MaxElapsedTime = 5 * time.Minute

	if err := backoff.Retry(operation, backoff.WithMaxRetries(backoffCfg, uint64(c.maxRetries))); err != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
	}

	// Success - reset circuit breaker
	c.recordSuccess()

	return apiResp.Models, nil
}

// isCircuitOpen checks if the circuit breaker is currently open
func (c *Client) isCircuitOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.circuitOpen {
		return false
	}

	// Check if timeout has elapsed (inline calculation to avoid deadlock)
	if time.Since(c.circuitOpenTime) >= c.circuitTimeout {
		c.circuitOpen = false
		c.failureCount = 0
		return false
	}

	return true
}

// timeSinceCircuitOpen returns the time since the circuit opened
func (c *Client) timeSinceCircuitOpen() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Since(c.circuitOpenTime)
}

// recordFailure increments the failure counter and opens the circuit if threshold is reached
func (c *Client) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failureCount++
	if c.failureCount >= c.circuitThreshold {
		c.circuitOpen = true
		c.circuitOpenTime = time.Now()
	}
}

// recordSuccess resets the failure counter and closes the circuit
func (c *Client) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failureCount = 0
	c.circuitOpen = false
}

// IsCircuitOpen returns the current state of the circuit breaker (for testing)
func (c *Client) IsCircuitOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.circuitOpen
}

// GetFailureCount returns the current failure count (for testing)
func (c *Client) GetFailureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failureCount
}

// ValidateResponse validates the API response and returns parsing errors
func (c *Client) ValidateResponse(models []models.APIModel) []error {
	var errors []error

	for i, model := range models {
		if model.ID == "" {
			errors = append(errors, fmt.Errorf("model at index %d: missing ID", i))
			continue
		}

		if model.Name == "" {
			errors = append(errors, fmt.Errorf("model %s: missing name", model.ID))
		}

		// Validate pricing if present
		if model.Pricing.Prompt != "" {
			if _, err := strconv.ParseFloat(model.Pricing.Prompt, 64); err != nil {
				errors = append(errors, fmt.Errorf("model %s: invalid prompt pricing: %w", model.ID, err))
			}
		}

		if model.Pricing.Completion != "" {
			if _, err := strconv.ParseFloat(model.Pricing.Completion, 64); err != nil {
				errors = append(errors, fmt.Errorf("model %s: invalid completion pricing: %w", model.ID, err))
			}
		}
	}

	return errors
}

// ParsePrice parses a price string to float64
func ParsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	price, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return price
}

// ExtractProvider extracts the provider from a model ID (e.g., "google" from "google/gemini-2.5-pro")
func ExtractProvider(modelID string) string {
	parts := strings.Split(modelID, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}