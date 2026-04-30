# Skill Registry

**Project**: orpdic
**Last Updated**: 2026-04-28
**Persistence**: engram (primary)

---

## User-Level Skills (Global)

Located at: `~/.config/opencode/skills/`

### SDD (Spec-Driven Development)
| Skill | Purpose | Trigger |
|-------|---------|---------|
| `sdd-init` | Initialize SDD context | Project initialization |
| `sdd-explore` | Investigate ideas, clarify requirements | Before committing to a change |
| `sdd-propose` | Create change proposal with intent, scope, approach | When creating a change |
| `sdd-spec` | Write delta specifications | During spec phase |
| `sdd-design` | Create technical design document | During design phase |
| `sdd-tasks` | Break down change into task checklist | During task phase |
| `sdd-apply` | Implement tasks from change | During apply phase |
| `sdd-verify` | Validate implementation against specs | During verify phase |
| `sdd-archive` | Sync delta specs to main and archive | After completion |
| `sdd-onboard` | End-to-end SDD workflow walkthrough | Onboarding |
| `sdd-verify/strict-tdd` | TDD enforcement for apply phase | During sdd-apply |
| `sdd-apply/strict-tdd` | TDD patterns for apply phase | During sdd-apply |

### Development Workflow
| Skill | Purpose | Trigger |
|-------|---------|---------|
| `issue-creation` | GitHub issue workflow | Creating issues |
| `branch-pr` | PR creation workflow | Creating PRs |
| `judgment-day` | Adversarial dual-review protocol | When review is requested |

### Special Purpose
| Skill | Purpose | Trigger |
|-------|---------|---------|
| `skill-creator` | Create new AI agent skills | When creating skills |
| `skill-registry` | Update skill registry | After installing/removing skills |
| `go-testing` | Go testing patterns + Bubbletea TUI | Writing Go tests |
| `find-skills` | Discover and install skills | Looking for functionality |
| `caveman` | Ultra-compressed communication | "caveman mode", token efficiency |

### Internal
| Skill | Purpose |
|-------|---------|
| `_shared` | Internal shared references |

---

## Project Conventions

**None detected** — project is empty.

When project conventions are established (e.g., `AGENTS.md`, `.cursorrules`, coding guidelines), they will be documented here.

---

## Notes

- SDD skills auto-load based on detected context (e.g., `sdd-*` skills during SDD phases)
- `go-testing` auto-loads when writing Go tests or using Bubbletea TUI testing
- `caveman` skill triggers on "caveman mode" or token efficiency requests
