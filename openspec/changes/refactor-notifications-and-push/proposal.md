# Proposal: Refactor Notifications and Push

## Intent

Refactor Discord notifications for a cleaner, more readable presentation, reduce polling frequency to align with lower-noise monitoring, and establish the project in GitHub under the `orpdisc` module/repository identity.

## Scope

### In Scope
- Remove emojis from Discord embeds and field labels
- Assign fixed colors by notification type and limit cost/value rankings to the single best model
- Standardize polling defaults and module/import naming on `github.com/Yokonad/orpdisc`
- Initialize git, configure `origin`, and publish `main` to GitHub

### Out of Scope
- New Discord channels or routing strategies
- Advanced analytics, scoring models, or broader digest redesign

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `openrouter-discord-monitor-spec`: update notification presentation, ranking output, polling default, and repository/module expectations

## Approach

- Keep the existing monitor flow, but refactor `internal/discord` embed construction to remove emoji affordances and use explicit per-type color constants.
- Update ranking selection in `internal/processor` and digest formatting to surface only the best model for cost and value.
- Align `internal/config`, spec text, and git metadata with a 4-hour poll interval and the `github.com/Yokonad/orpdisc` module path.
- Perform git initialization/push as a separate execution step after proposal approval.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `openspec/specs/openrouter-discord-monitor-spec.md` | Modified | Update behavioral contract for polling, rankings, and notification UI |
| `go.mod` | Modified | Confirm canonical module path |
| `internal/config/config.go` | Modified | Set/verify 4h default polling interval |
| `internal/processor/processor.go` | Modified | Restrict ranking outputs to top 1 |
| `internal/discord/webhook.go` | Modified | Remove emojis and map distinct colors by notification type |
| `.git/`, remote config | New/Modified | Initialize repository and connect GitHub origin |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Module rename/import drift breaks builds | Medium | Update imports atomically and verify references before implementation |
| Notification color/type mapping becomes inconsistent | Low | Centralize color constants by notification type |
| Discord webhook rate limits during rollout | Low | Preserve existing batching/retry behavior |

## Rollback Plan

Revert the notification/config/import changes in one commit and restore the previous remote/repository configuration if the GitHub publish step is not accepted.

## Dependencies

- GitHub repository access for `https://github.com/Yokonad/orpdisc.git`

## Success Criteria

- [ ] Discord notifications contain no emojis in titles or fields
- [ ] Cost and value rankings surface only one best model each
- [ ] Default polling interval is 4 hours across code and spec
- [ ] Repository is initialized and published to `main` on GitHub
