# Delta for OpenRouter Discord Monitor

## MODIFIED Requirements

### REQ-5: Periodic Digest

- **MUST** send a digest notification after every poll cycle, regardless of whether model changes were detected
- **MUST** include top 1 model by lowest cost per 1K tokens
- **MUST** include top 1 model by highest context_length/cost ratio
- **MUST** mark the notification as digest type (`IsDigest=true`) for distinct Discord embed formatting
- **SHOULD** include model recommendations based on use case

(Previously: SHOULD send daily digest with top 5 models by lowest cost and highest context/cost ratio)

#### Scenario: Initial poll sends digest

- GIVEN the service starts for the first time
- WHEN the initial `poll()` completes
- THEN `SendDigest()` is called immediately after
- AND a digest notification with top 1 by cost and top 1 by ratio is sent to Discord

#### Scenario: Ticker cycle sends digest even with no model changes

- GIVEN the service is running and the ticker fires (every 4h)
- WHEN `poll()` runs and detects no model changes
- THEN `SendDigest()` is still called after `poll()`
- AND a digest notification is sent to Discord

#### Scenario: Ticker cycle sends both change notification and digest

- GIVEN the service is running and the ticker fires
- WHEN `poll()` runs and detects model changes
- THEN `poll()` sends a change notification via Discord webhook
- AND `SendDigest()` sends a separate digest notification
- AND the user receives two distinct notifications

#### Scenario: SendDigest error does not stop service

- GIVEN the service is running
- WHEN `SendDigest()` returns an error (e.g., webhook failure)
- THEN the error is logged
- AND the service continues its polling loop normally

## ADDED Requirements

### REQ-DIGEST-SCHEDULE: Digest Invocation in Start()

The `Start()` method MUST call `s.SendDigest(s.ctx)` after every `s.poll()` invocation — both the initial poll at startup and each ticker-triggered poll. If `SendDigest()` returns an error, the error MUST be logged but MUST NOT stop or interrupt the service loop.

#### Scenario: Startup calls poll then digest

- GIVEN the service is started
- WHEN `Start()` executes the initial `s.poll()`
- THEN `s.SendDigest(s.ctx)` is called immediately after
- AND both operations complete before entering the ticker loop

### REQ-ENV-FILE: Environment File for Production

The file `/etc/openrouter-monitor.env` MUST be created with at minimum `DISCORD_WEBHOOK_URL` and `POLL_INTERVAL_MINUTES` variables. The file MUST have permissions `0600` (owner read/write only) to protect the webhook URL. Additional config variables (`DB_PATH`, `LOG_LEVEL`, `HTTP_TIMEOUT_SECONDS`, etc.) SHOULD be included.

#### Scenario: Service reads configuration from env file

- GIVEN `/etc/openrouter-monitor.env` exists with `DISCORD_WEBHOOK_URL` set
- WHEN the openrouter-monitor systemd service starts
- THEN the service reads the webhook URL from the environment file
- AND successfully sends Discord notifications

#### Scenario: Env file has restricted permissions

- GIVEN `/etc/openrouter-monitor.env` contains a webhook URL
- WHEN file permissions are checked
- THEN the file mode is `0600` (readable only by owner)

### REQ-SYSTEMD-SERVICE: Systemd Service Deployment

The `openrouter-monitor.service` unit file MUST be installed to `/etc/systemd/system/`. The service MUST be enabled (auto-start on boot) and started after installation. The service MUST load environment variables from `/etc/openrouter-monitor.env` via `EnvironmentFile` directive.

#### Scenario: Service starts on system boot

- GIVEN the system boots
- WHEN systemd activates enabled units
- THEN the openrouter-monitor service starts automatically
- AND begins polling on the configured interval

#### Scenario: Service can be verified via systemctl

- GIVEN the service is installed and enabled
- WHEN `systemctl status openrouter-monitor` is executed
- THEN the status shows `active (running)`

### REQ-COMMIT-PENDING: Commit Uncommitted Changes

All uncommitted changes in the repository MUST be committed with a descriptive conventional commit message referencing the originating change.

#### Scenario: Clean working tree after commit

- GIVEN there are unstaged or untracked files from prior changes
- WHEN the commit is executed
- THEN `git status` shows a clean working tree
- AND the commit message follows conventional commit format
