# Implementation Tasks: OpenRouter Discord Monitor

## Phase 1: Project Setup

### Task 1.1: Initialize Go Module
- [x] Run `go mod init github.com/user/orpdic`
- [x] Create directory structure (cmd/, internal/)
- [x] Create .gitignore for Go projects
- **Acceptance**: `go mod tidy` completes without errors

### Task 1.2: Add Dependencies
- [x] Add `github.com/mattn/go-sqlite3 v1.14.22`
- [x] Add `github.com/caarlos0/env/v10 v10.0.0`
- [x] Add `github.com/cenkalti/backoff/v4 v4.3.0`
- [x] Run `go mod tidy`
- **Acceptance**: go.sum created, imports resolve

### Task 1.3: Create Configuration Module
- [x] Create `internal/config/config.go`
- [x] Define Config struct with env tags
- [x] Implement Load() function
- [x] Add validation (webhook URL required)
- **Acceptance**: Unit tests pass for config loading

## Phase 2: Database Layer

### Task 2.1: Create Database Schema
- [x] Create `internal/database/db.go`
- [x] Define schema for models table
- [x] Define schema for price_history table
- [x] Define schema for notifications table
- [x] Create indexes for performance
- **Acceptance**: Schema matches spec, migrations run

### Task 2.2: Implement Database Operations
- [x] Open() - connection with WAL mode
- [x] Migrate() - create tables if not exist
- [x] GetModel(id) - fetch single model
- [x] GetAllModels() - fetch all models
- [x] SaveModel(model) - insert or update
- [x] SavePriceHistory(model) - track price changes
- [x] LogNotification(notification) - audit log
- **Acceptance**: All CRUD operations tested

### Task 2.3: Add Model Types
- [x] Create `internal/models/types.go`
- [x] Define Model struct
- [x] Define Changeset struct
- [x] Define Notification struct
- [x] Add JSON tags for serialization
- **Acceptance**: Types compile, JSON marshal/unmarshal works

## Phase 3: OpenRouter API Client

### Task 3.1: Create HTTP Client
- [x] Create `internal/openrouter/client.go`
- [x] Implement Client struct with http.Client
- [x] Add timeout configuration
- [x] Add user-agent header
- **Acceptance**: Client can make HTTP requests

### Task 3.2: Implement Model Fetching
- [x] FetchModels(ctx) method
- [x] Parse JSON response into []Model
- [x] Handle API errors (4xx, 5xx)
- [x] Add retry logic with exponential backoff
- **Acceptance**: Successfully fetches and parses real API

### Task 3.3: Add Response Validation
- [x] Validate required fields present
- [x] Parse pricing strings to float64
- [x] Handle missing optional fields
- [x] Log parsing errors
- **Acceptance**: Invalid responses don't crash service

## Phase 4: Change Detection Processor

### Task 4.1: Implement Hash Calculation
- [x] Create `internal/processor/processor.go`
- [x] CalculateModelHash(model) function
- [x] Normalize model before hashing
- [x] Return consistent hash for same data
- **Acceptance**: Same model data = same hash

### Task 4.2: Implement Change Detection
- [x] ProcessModels(ctx, models) method
- [x] Compare incoming vs stored models
- [x] Identify new, updated, removed models
- [x] Return Changeset with categorized changes
- **Acceptance**: Unit tests for all change scenarios

### Task 4.3: Calculate Derived Metrics
- [x] Cost per 1K tokens calculation
- [x] Context/cost ratio calculation
- [x] Store metrics in model struct
- **Acceptance**: Calculations match expected values

## Phase 5: Discord Webhook Integration

### Task 5.1: Create Webhook Client
- [x] Create `internal/discord/webhook.go`
- [x] Implement WebhookClient struct
- [x] Validate webhook URL format
- [x] Add HTTP client configuration
- **Acceptance**: Client initializes correctly

### Task 5.2: Implement Embed Builder
- [x] BuildEmbedForChange(change) function
- [x] Color coding: green (new), yellow (update), red (removed)
- [x] Format model details (name, provider, pricing)
- [x] Add OpenRouter links
- **Acceptance**: Embeds match Discord API format

### Task 5.3: Send Notifications
- [x] SendNotification(ctx, changeset) method
- [x] Batch up to 10 embeds per request
- [x] Handle Discord rate limits (429)
- [x] Retry with exponential backoff
- [x] Log success/failure
- **Acceptance**: Test webhook receives messages

## Phase 6: Service Orchestration

### Task 6.1: Create Main Service
- [x] Create `internal/service/service.go`
- [x] Implement Service struct with all dependencies
- [x] Add Start() method with ticker
- [x] Add Stop() method for graceful shutdown
- **Acceptance**: Service starts and stops cleanly

### Task 6.2: Implement Polling Loop
- [x] Ticker triggers every PollInterval
- [x] Fetch -> Process -> Notify flow
- [x] Error handling at each step
- [x] Continue on non-fatal errors
- **Acceptance**: Loop runs continuously

### Task 6.3: Add Signal Handling
- [x] Handle SIGTERM, SIGINT
- [x] Graceful shutdown with timeout
- [x] Close database connection
- [x] Cancel in-flight operations
- **Acceptance**: Ctrl+C stops service cleanly

### Task 6.4: Add Structured Logging
- [x] Create logger with levels
- [x] Log all major events
- [x] Redact sensitive data (webhook token)
- [x] Add correlation IDs
- **Acceptance**: Logs are structured and readable

## Phase 7: Circuit Breaker & Resilience

### Task 7.1: Implement Circuit Breaker
- [x] Track consecutive failures
- [x] Open circuit after 5 failures
- [x] Pause for 1 hour when open
- [x] Reset on successful request
- **Acceptance**: Circuit opens/closes correctly

### Task 7.2: Add Health Checks
- [x] Optional HTTP server on :8080
- [x] /health endpoint
- [x] Check database connectivity
- [x] Return 200/503 status
- **Acceptance**: Health endpoint responds

## Phase 8: Deployment & Documentation

### Task 8.1: Create Dockerfile
- [x] Multi-stage build (builder + runtime)
- [x] Alpine base image
- [x] Non-root user
- [x] Expose health check port
- **Acceptance**: Image builds and runs

### Task 8.2: Create systemd Service File
- [x] Write openrouter-monitor.service
- [x] Auto-restart configuration
- [x] Environment file support
- [x] Logging to journald
- **Acceptance**: Service installs and runs

### Task 8.3: Write README
- [x] Project description
- [x] Installation instructions
- [x] Configuration options
- [x] Usage examples
- [x] Troubleshooting guide
- **Acceptance**: README is complete and accurate

## Phase 9: Testing & Verification

### Task 9.1: Unit Tests
- [x] Test config package
- [x] Test database operations
- [x] Test processor logic
- [x] Test webhook client
- [x] Aim for >70% coverage
- **Acceptance**: `go test ./...` passes (note: database tests require CGO_ENABLED=1 due to go-sqlite3)

### Task 9.2: Integration Test
- [x] Test against real OpenRouter API
- [x] Verify Discord webhook receives message
- [x] Test error scenarios
- [x] Verify database persistence
- **Acceptance**: Manual test confirms end-to-end flow

### Task 9.3: Load Testing
- [x] Test with large API response
- [x] Verify memory usage stays low
- [x] Test concurrent operations
- **Acceptance**: No memory leaks, handles load

## Task Checklist Summary

| Phase | Tasks | Total |
|-------|-------|-------|
| 1. Setup | 1.1, 1.2, 1.3 | 3 |
| 2. Database | 2.1, 2.2, 2.3 | 3 |
| 3. API Client | 3.1, 3.2, 3.3 | 3 |
| 4. Processor | 4.1, 4.2, 4.3 | 3 |
| 5. Discord | 5.1, 5.2, 5.3 | 3 |
| 6. Service | 6.1, 6.2, 6.3, 6.4 | 4 |
| 7. Resilience | 7.1, 7.2 | 2 |
| 8. Deployment | 8.1, 8.2, 8.3 | 3 |
| 9. Testing | 9.1, 9.2, 9.3 | 3 |
| **Total** | | **27** |

## Implementation Order

Recommended batching for implementation:

**Batch 1**: 1.1 → 1.2 → 1.3 → 2.1 → 2.2 → 2.3 (Foundation)
**Batch 2**: 3.1 → 3.2 → 3.3 → 4.1 → 4.2 → 4.3 (Core Logic)
**Batch 3**: 5.1 → 5.2 → 5.3 → 6.1 → 6.2 → 6.3 → 6.4 (Integration)
**Batch 4**: 7.1 → 7.2 → 8.1 → 8.2 → 8.3 → 9.1 → 9.2 → 9.3 (Polish & Deploy)
