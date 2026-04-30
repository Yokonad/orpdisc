# Technical Design: OpenRouter Discord Monitor

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenRouter Discord Monitor                 │
│                        (Go Service)                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Poller    │───▶│   Processor  │───▶│   Notifier   │  │
│  │  (Ticker)   │    │  (Change     │    │  (Discord)   │  │
│  │             │    │   Detection) │    │              │  │
│  └─────────────┘    └──────────────┘    └──────────────┘  │
│         │                   │                    │           │
│         ▼                   ▼                    ▼           │
│  ┌────────────────────────────────────────────────────┐   │
│  │                    SQLite Storage                   │   │
│  │  ┌─────────┐  ┌─────────────┐  ┌──────────────┐   │   │
│  │  │ models  │  │price_history│  │notifications │   │   │
│  │  └─────────┘  └─────────────┘  └──────────────┘   │   │
│  └────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         │                                           │
         ▼                                           ▼
┌─────────────────┐                      ┌─────────────────────┐
│ OpenRouter API  │                      │ Discord Webhook     │
│ /api/v1/models  │                      │ /api/webhooks/...   │
└─────────────────┘                      └─────────────────────┘
```

## Module Structure

```
orpdic/
├── cmd/
│   └── monitor/
│       └── main.go              # Entry point, config, service lifecycle
├── internal/
│   ├── config/
│   │   └── config.go            # Environment-based configuration
│   ├── database/
│   │   └── db.go                # SQLite connection, migrations
│   ├── models/
│   │   └── types.go             # Domain models (Model, PriceChange, etc.)
│   ├── openrouter/
│   │   └── client.go            # HTTP client for OpenRouter API
│   ├── processor/
│   │   └── processor.go         # Change detection logic
│   ├── discord/
│   │   └── webhook.go           # Discord webhook client
│   └── service/
│       └── service.go           # Main service orchestration
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## Key Components

### 1. Configuration (internal/config/config.go)
```go
type Config struct {
    DiscordWebhookURL    string        `env:"DISCORD_WEBHOOK_URL,required"`
    PollInterval         time.Duration `env:"POLL_INTERVAL_MINUTES" envDefault:"30m"`
    DatabasePath          string        `env:"DB_PATH" envDefault:"./data.db"`
    LogLevel             string        `env:"LOG_LEVEL" envDefault:"info"`
    HTTPTimeout          time.Duration `env:"HTTP_TIMEOUT_SECONDS" envDefault:"30s"`
}
```

### 2. Database Layer (internal/database/db.go)
- Uses `github.com/mattn/go-sqlite3` driver
- Automatic migrations on startup
- Connection pooling with WAL mode for better concurrency
- Prepared statements for frequent queries

### 3. OpenRouter Client (internal/openrouter/client.go)
```go
type Client struct {
    httpClient *http.Client
    baseURL    string
}

func (c *Client) FetchModels(ctx context.Context) ([]Model, error)
```
- Implements exponential backoff with `github.com/cenkalti/backoff/v4`
- Circuit breaker pattern: tracks consecutive failures, pauses after threshold
- Streaming JSON decoder for memory efficiency

### 4. Processor (internal/processor/processor.go)
```go
type Processor struct {
    db *database.DB
}

func (p *Processor) ProcessModels(ctx context.Context, models []Model) (*Changeset, error)
```
- Compares incoming models against stored state
- Generates hash for each model (MD5 of normalized JSON)
- Returns Changeset: new models, updated models, removed models
- Calculates derived metrics (cost per 1K, context/cost ratio)

### 5. Discord Notifier (internal/discord/webhook.go)
```go
type WebhookClient struct {
    url        string
    httpClient *http.Client
}

func (c *WebhookClient) SendNotification(ctx context.Context, changes *Changeset) error
```
- Batches up to 10 embeds per webhook call
- Respects rate limits (429 handling)
- Color-coded embeds based on change type
- Retry with exponential backoff

### 6. Service Orchestration (internal/service/service.go)
```go
type Service struct {
    config    *config.Config
    db        *database.DB
    client    *openrouter.Client
    processor *processor.Processor
    notifier  *discord.WebhookClient
    ticker    *time.Ticker
}

func (s *Service) Start(ctx context.Context) error
func (s *Service) Stop() error
```
- Manages lifecycle: init, run, graceful shutdown
- Signal handling (SIGTERM, SIGINT)
- Health check endpoint (optional HTTP server)

## Data Flow

1. **Initialization**
   - Load configuration from environment
   - Initialize SQLite with migrations
   - Create HTTP client with timeouts
   - Validate Discord webhook URL

2. **Polling Loop**
   - Ticker triggers every POLL_INTERVAL_MINUTES
   - Fetch models from OpenRouter API
   - Parse and validate response
   - Calculate hashes and metrics

3. **Change Detection**
   - Query existing models from database
   - Compare by ID and hash
   - Identify: new, updated, removed
   - Store new state in database

4. **Notification**
   - If changes detected, build Discord embeds
   - Send webhook request
   - Log notification to database
   - Handle rate limits and retries

5. **Error Handling**
   - API errors: exponential backoff
   - 5 consecutive failures: circuit breaker (1h pause)
   - Discord errors: retry with backoff
   - Database errors: log and continue

## Error Recovery

| Error Scenario | Recovery Strategy |
|---------------|-------------------|
| OpenRouter 5xx | Retry with exponential backoff, max 5 attempts |
| OpenRouter 429 | Wait for Retry-After header |
| 5 consecutive failures | Circuit breaker: pause 1 hour |
| Discord 429 | Queue for retry, respect Retry-After |
| Database locked | Retry with short delay (SQLite busy) |
| Webhook URL invalid | Fatal error, service exits |

## Deployment

### Option 1: systemd (Recommended for VPS)
```ini
# /etc/systemd/system/openrouter-monitor.service
[Unit]
Description=OpenRouter Discord Monitor
After=network.target

[Service]
Type=simple
User=monitor
WorkingDirectory=/opt/monitor
ExecStart=/opt/monitor/monitor
Restart=always
RestartSec=10
Environment="DISCORD_WEBHOOK_URL=..."
Environment="DB_PATH=/var/lib/monitor/data.db"

[Install]
WantedBy=multi-user.target
```

### Option 2: Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o monitor ./cmd/monitor

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/monitor .
ENV DB_PATH=/data/data.db
VOLUME ["/data"]
CMD ["./monitor"]
```

## Security Considerations

1. **Webhook Token**: Store in env var, never commit to repo
2. **API Keys**: Not required for public models endpoint
3. **Database**: SQLite file should have restricted permissions (0600)
4. **Logging**: Redact webhook URL in logs
5. **Network**: HTTPS only, certificate validation

## Monitoring

- Structured logging with levels (debug, info, warn, error)
- Log entries: poll_start, poll_complete, change_detected, notification_sent, error
- Optional: HTTP health check endpoint at :8080/health
- Metrics: polls_total, changes_total, notifications_total, errors_total

## Testing Strategy

1. **Unit Tests**: Mock HTTP clients, in-memory SQLite
2. **Integration Tests**: Test against OpenRouter sandbox (if available)
3. **E2E Tests**: Manual webhook verification

## Dependencies

```go
require (
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/caarlos0/env/v10 v10.0.0      // Config parsing
    github.com/cenkalti/backoff/v4 v4.3.0    // Retry logic
    github.com/stretchr/testify v1.9.0       // Testing
)
```
