# Delta for openrouter-discord-monitor-spec

## MODIFIED Requirements

### Requirement: REQ-4: Discord Notifications
(Previously: used general rich embeds with limited color-coding and included emojis in field titles)

- **MUST** send webhook to: https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOhy0elTv2ZcwNe4uHKqNJj8nd
- **MUST** use rich embeds with specific color coding by type:
  - New Models: GREEN (5763719)
  - Price Changes: YELLOW (16776960)
  - Removed Models: RED (15548997)
- **MUST NOT** include any emojis in titles, descriptions, or field names.
- **MUST** batch multiple changes into single webhook call
- **SHOULD** include direct links to models on OpenRouter

#### Scenario: Notification without emojis
- **GIVEN** a new model is detected
- **WHEN** the service prepares the webhook payload
- **THEN** the embed title "OpenRouter Model Update" contains no emojis
- **AND** all field titles (e.g., "New Models") contain no emojis

### Requirement: REQ-5: Periodic Digest
(Previously: sent digest with top 5 models)

- **MUST** send daily digest with the top 1 best model by: lowest cost, highest context/cost ratio.
- **SHOULD** include model recommendations based on use case.

#### Scenario: Digest notification with only 1 best model
- **GIVEN** multiple models are available
- **WHEN** the daily digest is generated
- **THEN** only the top 1 model for lowest cost is included
- **AND** only the top 1 model for highest context/cost ratio is included

### Requirement: REQ-6: Configuration
(Previously: defined defaults and variables)

- **MUST** support env vars: DISCORD_WEBHOOK_URL, POLL_INTERVAL_MINUTES (default 240, which is 4h), DB_PATH (default ./data.db), LOG_LEVEL
- **MUST** default `POLL_INTERVAL_MINUTES` to `240` (4 hours).

## ADDED Requirements

### Requirement: REQ-7: Repository and Deployment

- **MUST** synchronize local repository with GitHub remote `https://github.com/Yokonad/orpdisc.git`
- **MUST** use `main` as the default branch.

#### Scenario: Push to GitHub
- **GIVEN** the codebase is refactored
- **WHEN** the user initiates a push
- **THEN** the code is pushed to the `main` branch of `https://github.com/Yokonad/orpdisc.git`
