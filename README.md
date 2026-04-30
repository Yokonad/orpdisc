# orpdisc

Automated service that polls OpenRouter API for model pricing/capabilities and sends Discord webhook notifications for the best model.

## Features

- **Automated Polling**: Polls OpenRouter API every 4 hours (configurable)
- **Change Detection**: Hash-based detection for new models, price changes, and removed models
- **Discord Notifications**: Rich embed notifications with dynamic colors
- **Circuit Breaker**: Resilience pattern to prevent API exhaustion
- **Health Checks**: Optional HTTP health check endpoint

## Installation

### Prerequisites

- Go 1.21+
- SQLite3
- Discord webhook URL

### From Source

```bash
git clone https://github.com/Yokonad/orpdisc.git
cd orpdisc
go build -o monitor ./cmd/monitor
```

### From Docker

```bash
docker build -t openrouter-monitor .
docker run -d \
  --name openrouter-monitor \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -v /path/to/data:/data \
  openrouter-monitor
```

### From systemd (Recommended for VPS)

```bash
# Copy binary
sudo cp monitor /opt/monitor/

# Copy service file
sudo cp openrouter-monitor.service /etc/systemd/system/

# Create environment file
sudo nano /etc/openrouter-monitor.env
# Add: DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...

# Create user
sudo useradd -r -m -d /var/lib/monitor -s /bin/false monitor

# Set permissions
sudo chown -R monitor:monitor /var/lib/monitor

# Reload systemd and start
sudo systemctl daemon-reload
sudo systemctl enable openrouter-monitor
sudo systemctl start openrouter-monitor
```

## Configuration

The service is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DISCORD_WEBHOOK_URL` | **required** | Discord webhook URL |
| `POLL_INTERVAL_MINUTES` | `240` | Minutes between polls |
| `DB_PATH` | `./data.db` | SQLite database file path |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `HTTP_TIMEOUT_SECONDS` | `30` | HTTP request timeout |
| `OPENROUTER_BASE_URL` | `https://openrouter.ai/api/v1` | OpenRouter API base URL |
| `MAX_RETRIES` | `5` | Maximum retry attempts for failed requests |
| `CIRCUIT_BREAKER_THRESHOLD` | `5` | Number of consecutive failures before circuit opens |
| `CIRCUIT_BREAKER_TIMEOUT_MINUTES` | `60` | Minutes to wait before retry after circuit opens |
| `HEALTH_CHECK_PORT` | `:8080` | Port for health check HTTP server (empty to disable) |

### Example Environment File

```
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOhy0elTv2ZcwNe4uHKqNJj8nd
POLL_INTERVAL_MINUTES=30
DB_PATH=/var/lib/monitor/data.db
LOG_LEVEL=info
CIRCUIT_BREAKER_THRESHOLD=5
CIRCUIT_BREAKER_TIMEOUT_MINUTES=60
HEALTH_CHECK_PORT=:8080
```

## Usage

### Running from Command Line

```bash
# With environment variables
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... ./monitor

# With environment file
./monitor  # reads from .env file if present
```

### Docker

```bash
docker run -d \
  --name openrouter-monitor \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -e POLL_INTERVAL_MINUTES=30 \
  -v /path/to/data:/data \
  openrouter-monitor
```

### systemd

```bash
# Start the service
sudo systemctl start openrouter-monitor

# Check status
sudo systemctl status openrouter-monitor

# View logs
sudo journalctl -u openrouter-monitor -f

# Stop the service
sudo systemctl stop openrouter-monitor

# Restart the service
sudo systemctl restart openrouter-monitor
```

## Health Checks

The service optionally exposes an HTTP health check endpoint:

```bash
# Set HEALTH_CHECK_PORT to enable
HEALTH_CHECK_PORT=:8080 ./monitor

# Check health
curl http://localhost:8080/health

# Response:
# - 200 OK: "healthy" when database is reachable
# - 503 Service Unavailable: "unhealthy: <error>" when database is down
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenRouter Discord Monitor               │
│                        (Go Service)                         │
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
│  └────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Discord Notifications

The monitor sends rich embeds to Discord:

- **🆕 New Models** (green): When new models appear on OpenRouter
- **📝 Model Updates** (yellow): When pricing or context changes
- **🗑️ Models No Longer Available** (red): When models are removed

## Troubleshooting

### Service won't start

1. Check that `DISCORD_WEBHOOK_URL` is set correctly
2. Verify database path is writable
3. Check logs: `journalctl -u openrouter-monitor -n 50`

### No notifications being sent

1. Verify webhook URL is valid
2. Check that the Discord channel hasn't been deleted
3. Review logs for API errors
4. Ensure models have actually changed (first run populates DB without notifications)

### Circuit breaker keeps opening

1. Check OpenRouter API status
2. Increase `CIRCUIT_BREAKER_TIMEOUT_MINUTES`
3. Increase `MAX_RETRIES` for temporary issues

### Database locked errors

1. Ensure only one instance is running
2. Check disk space
3. For high concurrency, consider switching to PostgreSQL

### Health check returns 503

1. Check database connectivity: `sqlite3 /path/to/data.db "SELECT 1;"`
2. Verify disk space isn't full
3. Check file permissions

## License

MIT
