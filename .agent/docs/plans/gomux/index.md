# Implementation Plan: gomux

**Design Summary:** [README.md](README.md)
**Referenced Patterns:** Inherited from architecture doc
**Related Spec:** [gomux-spec.md](../../specs/gomux-spec.md)

<!-- Design summary from README.md — approved by user -->
<!-- Patterns: Dependency Injection, Error Handling, Context Lifecycle -->

## Phase Overview

<!-- Each phase is a complete RED-GREEN-REFACTOR cycle in its own file -->
<!-- TDD is MANDATORY: tests must be written first, fail, then pass after implementation -->

| Phase | Description | Depends On | Status |
|-------|-------------|------------|--------|
| [01-domain](phases/01-domain.md) | Domain types, interfaces, sentinel errors, and project scaffolding | None | Complete |
| [02-orchestrator](phases/02-orchestrator.md) | App.Run orchestration logic with fake-driven tests | Phase 01 | Pending |
| [03-tmux-parsing](phases/03-tmux-parsing.md) | parseSessions function for tmux format string output | Phase 01 | Complete |
| [04-tmux-client](phases/04-tmux-client.md) | Tmux Client wiring: NewClient, ListSessions, ExecIntoSession | Phase 03 | Pending |
| [05-tui-model](phases/05-tui-model.md) | Bubbletea Model with list view, input view, and name validation | Phase 01 | Pending |
| [06-tui-picker](phases/06-tui-picker.md) | Picker.Pick public API, lipgloss styling, view rendering | Phase 05 | Pending |
| [07-main](phases/07-main.md) | cmd/gomux/main.go entrypoint wiring | Phase 02, 04, 06 | Pending |

<!-- Status values: Pending | RED | GREEN | Complete -->

## Implementation Notes

<!-- Phase implementers fill in notes/*.md during execution -->
<!-- These capture decisions, issues, and deviations for post-implementation review -->

| Phase | Notes |
|-------|-------|
| Phase 01 | [notes/01-domain.md](notes/01-domain.md) |
| Phase 02 | [notes/02-orchestrator.md](notes/02-orchestrator.md) |
| Phase 03 | [notes/03-tmux-parsing.md](notes/03-tmux-parsing.md) |
| Phase 04 | [notes/04-tmux-client.md](notes/04-tmux-client.md) |
| Phase 05 | [notes/05-tui-model.md](notes/05-tui-model.md) |
| Phase 06 | [notes/06-tui-picker.md](notes/06-tui-picker.md) |
| Phase 07 | [notes/07-main.md](notes/07-main.md) |

## Dependencies

**Dependency Types:**

- **None**: Phase can start immediately
- **Phase NN**: Must wait for specified phase to complete (all gates passed)
- **Parallel**: Multiple phases can execute concurrently

**Dependency Graph:**

```
Phase 01 (domain)
  ├── Phase 02 (orchestrator)
  ├── Phase 03 (tmux parsing) ── Phase 04 (tmux client)
  └── Phase 05 (tui model) ──── Phase 06 (tui picker)
                                         │
Phase 07 (main) ◄────────── Phase 02, 04, 06
```

**Parallel Execution:** Phases 02, 03, and 05 can execute in parallel after Phase 01 completes. They share no code dependencies beyond the domain types.

**Parallel Execution with Arborist:**

When phases are marked for parallel execution, they execute in separate git worktrees:

1. Main agent creates a worktree per parallel phase: `arborist create <phase-id>`
2. Main agent spawns subagents concurrently, each with its worktree path
3. Subagents work in isolation, commit changes, report completion
4. Main agent merges changes sequentially using `git -C`:
   ```bash
   git -C /path/to/main/repo merge <branch> --ff-only
   ```
5. Main agent removes worktrees: `arborist rm --force <phase-id>`

## Success Criteria

- All seven phases complete with gates passing
- `gomux` binary builds with `go build ./cmd/gomux/`
- `go test ./...` passes with all tests green
- App.Run orchestration covers: direct mode, empty state, TUI mode, user quit, and error propagation
- parseSessions correctly handles multi-session output, empty output, and malformed lines
- TUI model handles all keybindings: j/k navigation, Enter select, n new, q quit, Esc cancel
- Name validation blocks dots/colons and warns on duplicate names

## Status

**Progress:** 2/7 phases complete
**Current Phase:** Phase 03 complete
**Blocked:** None

---

_When implementation completes: Delete this entire plan directory (including README.md)._
