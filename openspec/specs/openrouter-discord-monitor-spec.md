# Specification: OpenRouter Discord Monitor

## Overview
Automated service that polls OpenRouter API for model pricing/capabilities and sends Discord webhook notifications for significant changes and best-value models.

## Functional Requirements

### REQ-1: API Polling
- **MUST** poll GET https://openrouter.ai/api/v1/models every 30 minutes
- **MUST** handle ~400KB JSON response with 300+ models
- **MUST** filter for text models only (output_modalities=text)
- **SHOULD** implement exponential backoff on errors: 2s, 4s, 8s, 16s, max 5 retries

### REQ-2: Data Parsing
- **MUST** extract: id, name, description, context_length, pricing.prompt, pricing.completion, top_provider.max_completion_tokens, architecture.tokenizer
- **MUST** calculate: cost_per_1k_tokens = (pricing.prompt + pricing.completion) * 1000
- **MUST** calculate: context_cost_ratio = context_length / cost_per_1k_tokens

### REQ-3: Change Detection
- **MUST** store model state in SQLite with hash-based change detection
- **MUST** detect: new models, price changes, context length changes, removed models
- **MUST** persist notification history to prevent duplicate alerts

### REQ-4: Discord Notifications
- **MUST** send webhook to: https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOhy0elTv2ZcwNe4uHKqNJj8nd
- **MUST** use rich embeds with color coding: green (new), yellow (change), red (removed)
- **MUST** batch multiple changes into single webhook call
- **SHOULD** include direct links to models on OpenRouter

### REQ-5: Periodic Digest
- **SHOULD** send daily digest with top 5 models by: lowest cost, highest context/cost ratio
- **SHOULD** include model recommendations based on use case

## Non-Functional Requirements

### NFR-1: Performance
- **MUST** parse 400KB response in < 1 second
- **MUST** handle API response streaming for memory efficiency
- **SHOULD** use connection pooling for HTTP client

### NFR-2: Reliability
- **MUST** implement circuit breaker: pause for 1h after 5 consecutive failures
- **MUST** log all errors with structured logging
- **SHOULD** expose health check endpoint

### NFR-3: Security
- **MUST** read webhook URL from environment variable
- **MUST NOT** log webhook token or API keys
- **SHOULD** support webhook URL rotation without restart

### NFR-4: Configurability
- **MUST** support env vars: DISCORD_WEBHOOK_URL, POLL_INTERVAL_MINUTES (default 30), DB_PATH (default ./data.db), LOG_LEVEL

## Data Schema

```sql
CREATE TABLE models (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    provider TEXT NOT NULL,
    description TEXT,
    context_length INTEGER,
    max_completion_tokens INTEGER,
    pricing_prompt REAL,
    pricing_completion REAL,
    tokenizer TEXT,
    data_hash TEXT NOT NULL,
    first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id TEXT NOT NULL,
    pricing_prompt REAL,
    pricing_completion REAL,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES models(id)
);

CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL, -- 'new', 'price_change', 'removed', 'digest'
    model_id TEXT,
    old_value TEXT,
    new_value TEXT,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_models_provider ON models(provider);
CREATE INDEX idx_price_history_model ON price_history(model_id);
CREATE INDEX idx_notifications_sent ON notifications(sent_at);
```

## Scenarios

### SC-1: Initial Poll
**Given** the database is empty
**When** the service polls OpenRouter API
**Then** it stores all models in SQLite
**And** sends Discord notification for "Nuevos modelos detectados: N modelos"

### SC-2: Price Decrease
**Given** model X costs $0.01/1K tokens
**When** OpenRouter reduces price to $0.005/1K tokens
**Then** service detects hash change
**And** creates price_history entry
**And** sends Discord notification with old/new prices

### SC-3: New Model Added
**Given** the service has existing models in DB
**When** OpenRouter adds a new model Y
**Then** service detects new model ID
**And** inserts into models table
**And** sends Discord notification with model details

### SC-4: API Failure
**Given** OpenRouter API returns 500
**When** service attempts to poll
**Then** it retries with exponential backoff (max 5 attempts)
**And** logs error
**And** pauses for 1h after 5 consecutive failures

### SC-5: Discord Rate Limit
**Given** many changes detected
**When** sending notifications
**And** Discord returns 429
**Then** service respects Retry-After header
**And** queues remaining notifications

## Discord Embed Format

```json
{
  "embeds": [
    {
      "title": "Actualización de Modelos de OpenRouter",
      "description": "3 cambios detectados",
      "color": 3447003,
      "timestamp": "2026-04-28T16:00:00.000Z",
      "fields": [
        {
          "name": "⬆️ Nuevos Modelos",
          "value": "• google/gemini-2.5-pro\n• anthropic/claude-opus-4.7",
          "inline": false
        },
        {
          "name": "💰 Cambios de Precios",
          "value": "• openai/gpt-5.4: $0.01 → $0.008",
          "inline": false
        }
      ],
      "footer": {
        "text": "Monitor de OpenRouter"
      }
    }
  ]
}
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| DISCORD_WEBHOOK_URL | required | Discord webhook URL |
| POLL_INTERVAL_MINUTES | 30 | Minutes between polls |
| DB_PATH | ./data.db | SQLite database file |
| LOG_LEVEL | info | debug, info, warn, error |
| HTTP_TIMEOUT_SECONDS | 30 | API request timeout |

## API Endpoint

```
GET https://openrouter.ai/api/v1/models
```

Response structure per model:
- `id`: "anthropic/claude-opus-4.7"
- `name`: "Claude Opus 4.7"
- `context_length`: 200000
- `pricing.prompt`: "0.000015"
- `pricing.completion`: "0.000075"
- `top_provider.max_completion_tokens`: 4096
- `architecture.tokenizer": "claude"
