# Design: Translate to Spanish

## Technical Approach

Apply a copy-only localization pass to the existing notification, service, and documentation layers. The implementation will keep current control flow, types, and interfaces unchanged; only user-facing strings in `internal/discord/webhook.go`, `internal/service/service.go`, and `README.md` will be rewritten to Spanish so the behavior defined by `sdd/translate-to-spanish/spec` changes language without changing semantics.

## Architecture Decisions

### Decision: Localize in place instead of introducing i18n infrastructure

| Option | Tradeoff | Decision |
|---|---|---|
| Replace literals directly in scoped files | Small diff, matches current architecture, but not extensible to multiple locales | Chosen |
| Add translation tables/constants package | Better reuse, but unnecessary abstraction for a single-language change | Rejected |
| Add runtime locale switching | Future-proof, but out of scope and behavior-changing | Rejected |

Rationale: The proposal scopes this change to Spanish-only user-facing copy. Adding localization infrastructure would increase surface area without solving a current problem.

### Decision: Preserve identifiers and only translate externally visible text

| Option | Tradeoff | Decision |
|---|---|---|
| Translate strings only | Safe, low-risk, no API/code churn | Chosen |
| Rename variables, functions, and exported symbols | More linguistic consistency, but higher risk and unnecessary churn | Rejected |

Rationale: The proposal explicitly excludes internal identifiers. Keeping code identifiers stable avoids refactors that provide ZERO product value.

### Decision: Update tests that assert copy exactly

| Option | Tradeoff | Decision |
|---|---|---|
| Keep tests unchanged | Fast, but guarantees failures or stale English expectations | Rejected |
| Adjust literal assertions to Spanish | Minimal maintenance, verifies localized contract | Chosen |
| Remove string assertions | Less brittle, but loses coverage of required copy | Rejected |

Rationale: `internal/discord/webhook_test.go` currently asserts English digest titles, so the test suite must follow the new language contract.

## Data Flow

The runtime flow does not change; only emitted text changes.

    Service poll/start ──→ logger.Info/Debug/Error ──→ Spanish log lines
           │
           ├──→ HealthCheckServer /health ──→ Spanish plain-text response
           │
           └──→ WebhookClient.BuildEmbedsForChangeset ──→ Spanish Discord embeds

README translation is static documentation work and does not affect runtime flow.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `openspec/changes/translate-to-spanish/design.md` | Create | Document implementation strategy for the localization change. |
| `internal/discord/webhook.go` | Modify | Replace embed titles, descriptions, field labels, digest labels, and footer text with Spanish equivalents; optionally centralize repeated literals as local constants in the same file. |
| `internal/discord/webhook_test.go` | Modify | Update assertions that currently expect English notification copy, especially digest title checks. |
| `internal/service/service.go` | Modify | Translate startup, polling, shutdown, digest, error, and `/health` response strings to Spanish without changing logging structure. |
| `README.md` | Modify | Rewrite repository documentation in Spanish while preserving commands, filenames, env vars, and operational steps. |

## Interfaces / Contracts

No public interfaces, method signatures, routes, or configuration keys change.

Contract updates:
- `GET /health` keeps the same status codes (`200`, `503`) but returns Spanish plain text.
- Discord embeds keep the same structure (`Title`, `Description`, `Fields`, `Footer`, `URL`, colors) with Spanish content.
- README keeps the same command examples and environment variable names.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Discord embed literals | Update `internal/discord/webhook_test.go` expected titles/labels to Spanish and keep existing color/shape assertions. |
| Unit | Health response text | Add or update handler tests for `/health` success/failure responses if present; otherwise verify manually during implementation because the string contract changes. |
| Review | README completeness | Compare section-by-section to ensure commands and variable names stay intact while prose becomes Spanish. |

## Migration / Rollout

No migration required. Rollout is immediate on deploy because this is a copy-only change.

## Open Questions

- [ ] Should the footer `OpenRouter Monitor` remain branded in English or be translated to `Monitor de OpenRouter` for consistency?
- [ ] Should `rate limited` error text in `RateLimitError.Error()` also be translated, or is it considered internal diagnostics outside the requested scope?
