# Change Proposal: fix-4h-notifications

## Intent

The **orpdic** (OpenRouter Discord Monitor) service is not running in production and has critical gaps in its notification logic:

1. **Service NOT running**: No systemd service installed, no `/etc/openrouter-monitor.env` file exists
2. **`SendDigest()` never called**: The service has a method to send periodic digests but it's never invoked from `Start()`. Currently, notifications are ONLY sent when model changes are detected. If no models change (common), NO notification is sent at all.
3. **Uncommitted changes**: Leftover changes from `translate-to-spanish` change need to be committed

This change ensures the service:
- Runs as a systemd service with proper configuration
- Sends a notification EVERY poll cycle (every 4 hours), regardless of whether models changed
- Has all code changes properly committed

---

## Scope

### Files to Modify

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/service/service.go` | Modify `Start()` method | Call `SendDigest()` after each `poll()` cycle to ensure notifications are always sent |
| `/etc/openrouter-monitor.env` | Create | Environment file with `DISCORD_WEBHOOK_URL` and other required vars |
| `openrouter-monitor.service` | Commit existing changes | Already modified, needs to be committed |
| `go.mod`, `go.sum` | Commit existing changes | Dependency updates from previous change |
| `openspec/` files | Commit existing changes | Spec files from `translate-to-spanish` change |

### In-Scope Changes

#### 1. Modify `internal/service/service.go`

**Current behavior:**
```go
func (s *Service) Start() error {
    ticker := time.NewTicker(s.cfg.PollInterval)
    defer ticker.Stop()
    
    s.poll() // Only runs once at startup
    
    for {
        select {
        case <-ticker.C:
            s.poll() // Only sends notification if changes detected
        }
    }
}
```

**Proposed behavior:**
```go
func (s *Service) Start() error {
    ticker := time.NewTicker(s.cfg.PollInterval)
    defer ticker.Stop()
    
    // Initial run
    s.poll()
    s.SendDigest(s.ctx) // Always send digest on startup
    
    for {
        select {
        case <-ticker.C:
            s.poll()              // Detects and notifies on changes
            s.SendDigest(s.ctx)   // Always sends scheduled digest
        }
    }
}
```

**Rationale:**
- `poll()` handles change detection + notification for actual model changes
- `SendDigest()` sends the regular scheduled notification (Top 1 by cost, Top 1 by context/cost ratio)
- Together, they ensure users ALWAYS receive a notification every 4 hours

#### 2. Create `/etc/openrouter-monitor.env`

**Content:**
```bash
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/<REDACTED>
POLL_INTERVAL_MINUTES=4h
DB_PATH=/var/lib/openrouter-monitor/data.db
LOG_LEVEL=info
HTTP_TIMEOUT_SECONDS=30
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
MAX_RETRIES=5
CIRCUIT_BREAKER_THRESHOLD=5
CIRCUIT_BREAKER_TIMEOUT_MINUTES=60
```

**Note:** The actual `DISCORD_WEBHOOK_URL` value needs to be obtained from the orchestrator (it was visible in the spec file: `https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOelTv2ZcwNe4uHKqNJj8nd`)

#### 3. Install systemd service

**Steps:**
```bash
# Copy service file
sudo cp openrouter-monitor.service /etc/systemd/system/

# Create environment file
sudo nano /etc/openrouter-monitor.env

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable openrouter-monitor
sudo systemctl start openrouter-monitor

# Verify
sudo systemctl status openrouter-monitor
```

#### 4. Commit pending changes

**Files to commit:**
- `go.mod`, `go.sum` - dependency updates
- `openrouter-monitor.service` - systemd unit modifications
- `README.md` - documentation updates
- `openspec/` files - spec updates from translate-to-spanish

**Commit message:**
```
chore: commit translate-to-spanish change artifacts

- go.mod/go.sum: Added indirect dependencies for modernc.org/sqlite
- openrouter-monitor.service: Updated paths for development environment
- README.md: Removed obsolete setup steps
- openspec/: Updated specs for Spanish translation change

Related: translate-to-spanish
```

---

## Approach

### Phase 1: Code Changes

1. **Modify `service.go`**:
   - Add `SendDigest(s.ctx)` call after `s.poll()` in the ticker loop
   - Handle potential errors from `SendDigest()` gracefully (log but don't stop the service)
   - Ensure context is properly passed through

### Phase 2: Configuration

2. **Create `/etc/openrouter-monitor.env`**:
   - Ask orchestrator for the Discord webhook URL
   - Create file with proper permissions (600, root:root)

### Phase 3: Deployment

3. **Install systemd service**:
   - Copy service file to `/etc/systemd/system/`
   - Enable and start the service
   - Verify it's running

### Phase 4: Cleanup

4. **Commit pending changes**:
   - Stage all unstaged changes
   - Create commit with descriptive message

---

## What's OUT of Scope

- **Changing the poll interval**: Currently set to 4h, this is working as intended
- **Modifying `SendDigest()` logic**: The method already correctly computes top models by cost and ratio
- **Adding new notification types**: No new Discord message formats needed
- **Database schema changes**: Existing schema supports all required functionality
- **API changes**: OpenRouter API interaction remains unchanged
- **Translation changes**: Spanish translation already completed in previous change

---

## Success Criteria

1. ✅ Service runs as systemd service with `systemctl status openrouter-monitor` showing active
2. ✅ `/etc/openrouter-monitor.env` exists with required environment variables
3. ✅ `Start()` method calls both `poll()` AND `SendDigest()` on each cycle
4. ✅ Discord webhook receives notification every 4 hours (even when no model changes)
5. ✅ All unstaged changes committed to git
6. ✅ Service logs show both poll execution and digest sending

---

## Risks

| Risk | Mitigation |
|------|------------|
| Duplicate notifications if `SendDigest()` also sends change notifications | `SendDigest()` only sends digest (top models), `poll()` sends change notifications - they serve different purposes |
| Service fails to start due to missing env vars | Validate env file exists before starting service |
| Webhook URL exposed in logs | Use `RedactedWebhookURL()` for all logging |
| Digest sends on every poll (not just periodic) | Current design sends digest every cycle - this is the intended behavior per REQ-5 |

---

## Dependencies

- Discord webhook URL (from orchestrator)
- Binary already built at `/home/yokonad/yndhome/yndthings/orpdic/monitor`
- Systemd available on target system

---

## Related Changes

- `translate-to-spanish`: Previous change that added Spanish translations (needs to be archived)
- `openrouter-discord-monitor`: Initial implementation change
