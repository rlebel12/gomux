# Implementation Plan: Session Profiles

**Design Summary:** [README.md](README.md)
**Referenced Patterns:** Inherited from architecture doc (DI, Error Handling, Context Lifecycle)
**Related Spec:** None

## Phase Overview

| Phase | Description | Depends On |
|-------|-------------|------------|
| 01-domain-types | Add Profile, Window, Pane, Selection types and ProfileLoader/SessionCreator interfaces to gomux package | None |
| 02-profile-loader | Implement YAML profile loader in new `internal/profile` package | Phase 01 |
| 03-session-creator | Add `CreateSessionFromProfile` method to tmux.Client | Phase 01 |
| 04-picker-profiles | Update TUI picker to display profiles section and return Selection type | Phase 01 |
| 05-app-integration | Wire profile loading and creation into App.Run routing logic | Phases 02, 03, 04 |
| 06-main-wiring | Wire ProfileLoader and SessionCreator into main.go, add yaml.v3 dependency | Phase 05 |

## Implementation Notes

| Phase | Notes |
|-------|-------|
| Phase 01 | [notes/01-domain-types.md](notes/01-domain-types.md) |
| Phase 02 | [notes/02-profile-loader.md](notes/02-profile-loader.md) |
| Phase 03 | [notes/03-session-creator.md](notes/03-session-creator.md) |
| Phase 04 | [notes/04-picker-profiles.md](notes/04-picker-profiles.md) |
| Phase 05 | [notes/05-app-integration.md](notes/05-app-integration.md) |
| Phase 06 | [notes/06-main-wiring.md](notes/06-main-wiring.md) |

## Dependencies

**Dependency Graph:**

```
Phase 01 (domain types)
├── Phase 02 (profile loader)    ─┐
├── Phase 03 (session creator)   ─┼─→ Phase 05 (app integration) → Phase 06 (main wiring)
└── Phase 04 (picker profiles)   ─┘
```

Phases 02, 03, and 04 can execute in parallel after Phase 01 completes. Phase 05 requires all three. Phase 06 is the final wiring pass.

**Parallel Execution with Arborist:**

Phases 02, 03, and 04 are independent and can run in separate worktrees after Phase 01 merges to main.

## Success Criteria

- `gomux` with no profiles file behaves identically to current behavior (backward compatible)
- Profiles defined in `~/.config/gomux/profiles.yml` appear in picker below running sessions
- Selecting a profile creates the full tmux session layout (windows, panes, dirs, commands) and attaches
- Profiles whose sessions already exist appear only in the sessions section
- All existing tests continue to pass
- New tests cover YAML parsing, profile filtering, picker navigation, session creation, and app routing

---

_When implementation completes: Delete this entire plan directory (including README.md)._
