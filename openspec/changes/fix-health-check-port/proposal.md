# Change Proposal: fix-health-check-port

## Summary
Fix health check server port conflict by making the port configurable via environment variable instead of hardcoded `:8080`.

## Problem
The health check server in `cmd/monitor/main.go:65` is hardcoded to port 8080, which conflicts with another service on the system:
```
Health server error: listen tcp :8080: bind: address already in use
```

## Proposed Solution

### 1. Add Config Field (`internal/config/config.go`)
Add `HealthCheckPort` field with env tag `HEALTH_CHECK_PORT` and default value `:9090`:
```go
HealthCheckPort string `env:"HEALTH_CHECK_PORT" envDefault:":9090"`
```

### 2. Update Main (`cmd/monitor/main.go`)
Replace hardcoded `:8080` with `cfg.HealthCheckPort`:
```go
healthServer := svc.HealthCheckServer(cfg.HealthCheckPort)
```

### 3. Environment File (`/etc/openrouter-monitor.env`)
Add configuration line:
```
HEALTH_CHECK_PORT=:9090
```

### 4. Build & Deploy
- Rebuild binary: `go build -o monitor ./cmd/monitor`
- Restart systemd service: `systemctl restart openrouter-monitor`

### 5. Verification
- Service running: `systemctl status openrouter-monitor`
- Health check: `curl http://localhost:9090/health`
- Test notification: Send test payload to verify end-to-end

## Files to Change
- `internal/config/config.go` - Add HealthCheckPort field
- `cmd/monitor/main.go` - Use config instead of hardcoded port
- `/etc/openrouter-monitor.env` - Add HEALTH_CHECK_PORT env var

## Constraints
- Port 9090 verified as FREE (port 8080 is in use)
- Service currently running on PID 3411358
- Working tree has 2 unpushed commits
- All user-facing strings must be in Spanish (not applicable for this technical config)

## Testing
1. Verify port 9090 is free before deployment
2. After deployment, confirm health endpoint responds on :9090
3. Verify service continues normal operation (polling, notifications)
