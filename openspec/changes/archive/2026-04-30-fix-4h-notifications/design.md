# Design: fix-4h-notifications

## Technical Approach

Single-line code change to wire `SendDigest()` into the `Start()` loop, plus production deployment (env file + systemd). The service already has all the logic — the bug is purely that `SendDigest` is never called. No new types, no new packages, no schema changes.

## Architecture Decisions

### Decision: Error isolation between poll and digest

| Option | Tradeoff |
|--------|----------|
| **A) Sequential, log errors separately** | poll() error logged → SendDigest() still runs. Simple, resilient. |
| B) Abort on poll error | Skip digest if poll fails. Loses digest when OpenRouter is down. |
| C) Parallel goroutines | Adds concurrency for no benefit at 4h intervals. |

**Choice**: A — poll and digest are independent concerns. If `poll()` fails (OpenRouter down), `SendDigest()` still runs and sends the last-known top models from the DB. Both errors are logged independently; neither stops the service loop.

### Decision: SendDigest on startup AND every ticker cycle

| Option | Tradeoff |
|--------|----------|
| **A) Initial + every ticker** | Immediate feedback that service is alive. Two calls on first boot. |
| B) Ticker only | No notification until first 4h passes. Silent startup. |

**Choice**: A — per REQ-DIGEST-SCHEDULE spec scenario "Startup calls poll then digest". Confirms service health immediately.

### Decision: Error handling for SendDigest

| Option | Tradeoff |
|--------|----------|
| **A) Log error, continue loop** | Non-fatal. Service keeps polling even if Discord is down. |
| B) Return error from Start() | One failed digest kills the entire service. Unacceptable. |

**Choice**: A — `if err := s.SendDigest(s.ctx); err != nil { s.logger.Error(...) }` matches existing pattern in `poll()` (line 181-183).

### Decision: Env file path at `/etc/openrouter-monitor.env`

The systemd unit file already references `EnvironmentFile=/etc/openrouter-monitor.env` (line 10 of `openrouter-monitor.service`). This is a fixed contract — no decision needed, just implementation.

### Decision: DB_PATH uses development path

| Option | Tradeoff |
|--------|----------|
| **A) `/home/yokonad/yndhome/yndthings/orpdic/data.db`** | Matches existing binary location and WorkingDirectory in service unit. |
| B) `/var/lib/openrouter-monitor/data.db` | FHS-compliant but requires creating directory and changing ownership. |

**Choice**: A — the service already runs as root with WorkingDirectory set to the project dir. Keeps it simple for a single-user deployment.

## Data Flow

Updated `Start()` sequence (lines 122-147 of `service.go`):

```
Service.Start()
  │
  ├─ s.poll()                          # initial: fetch → process → notify changes
  ├─ s.SendDigest(s.ctx)               # NEW: send top-1 digest
  │   if err → s.logger.Error(...)     # log only, don't stop
  │
  └─ for { select {
       case <-ticker.C:
         s.poll()                      # fetch → process → notify changes
         s.SendDigest(s.ctx)           # NEW: send top-1 digest
           if err → s.logger.Error(...)
       }}
```

Discord receives up to 2 notifications per cycle:
1. **Change notification** (from `poll()`) — only if models changed (green/yellow/red embeds)
2. **Digest notification** (from `SendDigest()`) — always (blue embed with top-1 cost + top-1 ratio)

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/service/service.go` | Modify | Add `s.SendDigest(s.ctx)` + error logging after each `s.poll()` call in `Start()` (lines 132 and 141) |
| `/etc/openrouter-monitor.env` | Create | Env file with `DISCORD_WEBHOOK_URL`, `POLL_INTERVAL_MINUTES=4h`, `DB_PATH`, `LOG_LEVEL=info`. Permissions 0600. |
| `/etc/systemd/system/openrouter-monitor.service` | Copy | From existing `openrouter-monitor.service` in project root. Then daemon-reload + enable + start. |

Note: Commit of pending unstaged changes (go.mod, go.sum, README.md, service file, openspec files) is a separate pre-deployment step.

## Interfaces / Contracts

No new interfaces. The existing `SendDigest(ctx context.Context) error` signature is already defined at line 274. The change is purely call-site wiring.

The `SendDigest` → `SendNotification` → `BuildEmbedsForChangeset` path already handles `IsDigest=true` correctly (blue embed, "Resumen Diario" title, distinct format). No Discord message format changes needed.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `Start()` calls `SendDigest` after `poll()` on startup | Mock `Service` dependencies (db, client, proc, webhook). Verify `webhook.SendNotification` called twice — once for poll changeset, once for digest changeset (`IsDigest=true`). |
| Unit | `SendDigest` error does not stop service loop | Mock `SendDigest` to return error, verify `Start()` continues looping. |
| Unit | `poll()` failure does not block `SendDigest` | Mock `client.FetchModels` to fail, verify `SendDigest` still called. |
| Integration | Service starts and sends digest with real DB | Manual test: `go build && ./monitor` with real env vars. Check Discord channel for blue embed. |
| E2E | Systemd service runs and sends notifications | Deploy, wait one cycle, check `journalctl -u openrouter-monitor` for digest log lines. |

Note: No `service_test.go` exists today. The unit tests above would be NEW — the first service-layer tests.

## Migration / Rollout

1. **Pre-deploy**: Commit all pending changes (go.mod, go.sum, README.md, service file, openspec files)
2. **Build**: `go build -o monitor ./cmd/monitor`
3. **Deploy code**: The binary at `./monitor` is already referenced by the systemd unit
4. **Deploy config**: Create `/etc/openrouter-monitor.env` with 0600 permissions
5. **Deploy service**: Copy unit file, daemon-reload, enable, start
6. **Verify**: `systemctl status openrouter-monitor` + check Discord channel
7. **Rollback**: `sudo systemctl stop openrouter-monitor` → revert the `service.go` change → rebuild → restart

No database migration required. No feature flags needed.

## Open Questions

- None — all decisions resolved by spec and proposal.
