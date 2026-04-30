# Tasks: fix-4h-notifications

## Phase 1: Code Implementation

- [ ] 1.1 Modify `internal/service/service.go`: In `Start()`, add `s.SendDigest(s.ctx)` after the initial `s.poll()` call (line 132), with error logging: `if err := s.SendDigest(s.ctx); err != nil { s.logger.Error("SendDigest failed: %v", err) }`
- [ ] 1.2 Modify `internal/service/service.go`: In `Start()` ticker loop (line 141), add `s.SendDigest(s.ctx)` after `s.poll()` with same error handling pattern

## Phase 2: Configuration & Deployment Setup

- [ ] 2.1 Create `/etc/openrouter-monitor.env` with content: `DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOhy0elTv2ZcwNe4uHKqNJj8nd`, `POLL_INTERVAL_MINUTES=4h`, `DB_PATH=/home/yokonad/yndhome/yndthings/orpdic/data.db`, `LOG_LEVEL=info`
- [ ] 2.2 Set permissions on `/etc/openrouter-monitor.env`: `sudo chmod 600 /etc/openrouter-monitor.env` and `sudo chown root:root /etc/openrouter-monitor.env`
- [ ] 2.3 Copy `openrouter-monitor.service` to `/etc/systemd/system/openrouter-monitor.service`: `sudo cp /home/yokonad/yndhome/yndthings/orpdic/openrouter-monitor.service /etc/systemd/system/`

## Phase 3: Build & Service Start

- [ ] 3.1 Build updated binary: `go build -o monitor ./cmd/monitor` in project directory
- [ ] 3.2 Reload systemd: `sudo systemctl daemon-reload`
- [ ] 3.3 Enable service: `sudo systemctl enable openrouter-monitor`
- [ ] 3.4 Start service: `sudo systemctl start openrouter-monitor`

## Phase 4: Verification

- [ ] 4.1 Check service status: `sudo systemctl status openrouter-monitor` — expect `active (running)`
- [ ] 4.2 Check journal logs: `sudo journalctl -u openrouter-monitor --no-pager -n 20` — expect startup messages and digest log lines

## Phase 5: Commit

- [ ] 5.1 Stage all changes: `git add go.mod go.sum README.md openrouter-monitor.service openspec/ internal/service/service.go`
- [ ] 5.2 Commit with message: `fix: call SendDigest on every poll cycle, install systemd service, create env file`

## Dependencies

| Task | Depends On |
|------|-----------|
| 1.1, 1.2 | None |
| 2.1, 2.2 | None |
| 2.3 | None |
| 3.1 | 1.1, 1.2 (code must be changed before build) |
| 3.2, 3.3, 3.4 | 2.1, 2.2, 2.3 (env file and service file must exist) |
| 4.1, 4.2 | 3.4 (service must be started) |
| 5.1, 5.2 | None (can commit anytime) |

## Implementation Order

1. Phase 1 (code changes) → 2. Phase 2 (config files) → 3. Phase 3 (build & deploy) → 4. Phase 4 (verify) → 5. Phase 5 (commit)
