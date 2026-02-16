# Phase 02: App Orchestrator

**Depends on:** Phase 01
**Phase Type:** Standard
**Status:** Complete

<!-- Standard phase: Full RED-GREEN-REFACTOR -->
<!-- Status progression: Pending -> RED -> GREEN -> Complete -->

---

## RED: Write Tests

**Objective:** Test all App.Run routing logic using fake implementations of SessionLister, SessionPicker, and SessionExecutor.

**Files:**

- `internal/gomux/gomux_test.go`

**Test Fakes** (defined in test file, `package gomux_test`):

```go
// fakeLister — returns configured sessions/error
// fakePicker — returns configured name/error, captures received sessions
// fakeExecutor — records whether called, captures session name, returns configured error

// See README.md "Test fakes" section for exact shapes.
```

**Parameterized Tests** (table format -- MUST use `t.Run`):

| Case | sessionName | Lister Returns | Picker Returns | Expected Behavior | Notes |
|------|-------------|----------------|----------------|-------------------|-------|
| `direct_mode` | `"work"` | N/A (not called) | N/A (not called) | executor called with `"work"` | CLI bypass, no listing or picking |
| `empty_state` | `""` | `[]Session{}, nil` | N/A (not called) | executor called with `"default"` | Auto-create default session |
| `tui_mode_single` | `""` | `[{Name:"dev"}], nil` | `"dev", nil` | executor called with `"dev"` | Single session, picker invoked |
| `tui_mode_multi` | `""` | `[{Name:"dev"},{Name:"api"}], nil` | `"api", nil` | executor called with `"api"` | Multiple sessions, picker returns second |
| `user_quit` | `""` | `[{Name:"dev"}], nil` | `"", ErrNoSelection` | Run returns nil | Clean exit, no executor call |

**Discrete Tests** (bullet format -- separate test functions):

- **Test lister error propagation**: Lister returns error, Run returns that error (no picker or executor called)
- **Test executor error propagation**: Picker returns a name, executor returns error, Run returns that error
- **Test picker receives sessions**: Verify the sessions passed to Pick match what the lister returned (use fakePicker.received)

**Assertions:**

- In direct mode, lister and picker must NOT be called
- In empty state, picker must NOT be called
- In TUI mode, all three are called in sequence: list -> pick -> exec
- ErrNoSelection from picker causes Run to return nil (not an error)
- Errors from lister and executor propagate through Run

**Edge Cases:**

- Direct mode: sessionName is non-empty, lister should never be invoked
- Empty state: zero sessions, picker should never be invoked

### Gate: RED

- [x] Test file created with all enumerated test cases
- [x] All tests FAIL (App.Run is a stub returning nil from Phase 01)
- [x] Test coverage includes happy path and all edge cases

---

## GREEN: Implement

**Objective:** Implement App.Run with mode routing: direct attach, empty state auto-create, TUI selection, and user quit handling.

**Files:**

- `internal/gomux/gomux.go` (update Run method)

**Implementation Guidance:**

```go
func (a *App) Run(ctx context.Context, sessionName string) error {
    // Implementation approach:
    // 1. If sessionName is non-empty → bypass lister and picker, exec directly
    // 2. Call lister to get existing sessions
    // 3. If lister returns error → propagate (fmt.Errorf wrapping)
    // 4. If sessions is empty → exec into DefaultSessionName (skip picker)
    // 5. Call picker with sessions
    // 6. If picker returns ErrNoSelection → return nil (clean exit)
    // 7. If picker returns other error → propagate
    // 8. Call executor with chosen name
    // 9. If executor returns error → propagate
    // 10. Return nil (executor replaces process on success, but fakes return nil)
    //
    // Error handling:
    // - Lister errors → fmt.Errorf("list sessions: %w", err)
    // - ErrNoSelection → return nil (not an error condition)
    // - Picker errors → fmt.Errorf("pick session: %w", err)
    // - Executor errors → fmt.Errorf("exec session: %w", err)
}
```

**Pattern Application:**

- **Error Handling**: Wrap non-domain errors with `fmt.Errorf` and `%w` for context. Use `errors.Is(err, ErrNoSelection)` to detect user quit.
- **Context Lifecycle**: Pass ctx through to all interface method calls.
- **Dependency Injection**: All behavior delegated to injected interfaces -- no concrete tmux or TUI logic here.

### Gate: GREEN

- [x] All tests from RED phase now PASS
- [x] Test command: `go test ./internal/gomux/ -v` (8 tests)
- [x] Implementation follows pseudocode logic flow

---

## REFACTOR: Quality

**Focus:** Orchestration code clarity.

**Review Areas:**

- **Error wrapping**: Consistent context strings in fmt.Errorf
- **Naming**: Variable names clearly convey the routing logic
- **Simplification**: Reduce nesting if possible (early returns for direct mode and empty state)
- **Readability**: The three modes (direct, empty, TUI) should be visually distinct in the code

### Gate: REFACTOR

- [x] Reviewed for code duplication and extracted common patterns
- [x] Variable and function names are clear and descriptive
- [x] Complex logic simplified where possible
- [x] Error messages are helpful and actionable

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Phase 07 depends on this phase

---

**Previous:** [Phase 01](01-domain.md)
**Next:** [Phase 07](07-main.md) (after Phases 04 and 06 also complete)
