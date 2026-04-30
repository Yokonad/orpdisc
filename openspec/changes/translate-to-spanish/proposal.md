# Proposal: Translate to Spanish

## Intent

Localize the OpenRouter Discord Monitor to Spanish so the user receives notifications, documentation, and operational feedback in their preferred language.

## Scope

### In Scope
- Translate Discord embed titles, descriptions, field names, and footers in `internal/discord`.
- Translate user-facing service messages, digest text, and health-check responses in `internal/service`.
- Translate `README.md` to Spanish while preserving setup steps, commands, and configuration names.

### Out of Scope
- Renaming variables, types, or internal identifiers.
- Translating API keys, environment variable names, or non-user-facing implementation comments.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `openrouter-discord-monitor-spec`: Localize user-facing notifications, logs, health responses, and repository documentation from English to Spanish.

## Approach

Apply a Spanish-first localization pass to current user-facing copy without changing behavior. Update the base spec examples so notification wording and documentation expectations match the new language contract.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/discord/webhook.go` | Modified | Translate embed copy and digest labels. |
| `internal/service/service.go` | Modified | Translate runtime logs, digest logs, and `/health` responses. |
| `README.md` | Modified | Translate repository documentation to Spanish. |
| `openspec/specs/openrouter-discord-monitor-spec.md` | Modified | Align spec examples and user-facing wording with Spanish localization. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Partial translation leaves mixed-language UX | Medium | Inventory all user-facing strings in scoped files and update tests/examples together. |
| Tests may assert English strings | Medium | Update string-based assertions alongside localization changes. |

## Rollback Plan

Revert the localized copy changes in `internal/discord`, `internal/service`, `README.md`, and the related spec delta if Spanish terminology causes confusion or breaks expectations.

## Dependencies

- Existing `openrouter-discord-monitor-spec` as the behavior contract to update.

## Success Criteria

- [ ] All Discord notification text is in Spanish.
- [ ] All scoped user-facing logs and health responses are in Spanish.
- [ ] `README.md` is fully translated to Spanish without changing commands or env var names.
