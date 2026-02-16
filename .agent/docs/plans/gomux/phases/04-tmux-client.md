# Phase 04: Tmux Client Wiring

**Depends on:** Phase 03
**Phase Type:** Infrastructure
**Status:** Complete

<!-- Infrastructure phase: GREEN-REFACTOR only (no testable logic) -->
<!-- The actual I/O operations (exec.CommandContext, syscall.Exec, exec.LookPath) -->
<!-- are thin wiring — not unit-testable without real tmux. -->
<!-- Status progression: Pending -> GREEN -> Complete -->

---

## Existing Test Verification

**Run existing tests:** `go test ./internal/tmux/ -v`

**Expected result:** All parseSessions tests from Phase 03 continue to pass after adding the Client type.

### Gate: EXISTING TESTS

- [x] Existing test suite executed
- [x] All tests PASS before proceeding

---

## GREEN: Implement

**Objective:** Implement the `Client` struct with `NewClient`, `ListSessions`, and `ExecIntoSession` methods.

**Files:**

- `internal/tmux/tmux.go` (add to existing file)

**Implementation Guidance:**

```go
// Client struct:
//   - tmuxPath string (absolute path to tmux binary)
//   - Satisfies gomux.SessionLister and gomux.SessionExecutor

// NewClient() (*Client, error):
//   Implementation approach:
//   1. Call exec.LookPath("tmux") to resolve absolute path
//   2. If not found → return nil, fmt.Errorf("tmux not found on PATH: %w", err)
//   3. Return &Client{tmuxPath: resolved}, nil
//
//   Rationale: Resolving path eagerly means ExecIntoSession
//   doesn't need to look it up at exec time. Fail-fast if tmux missing.

// ListSessions(ctx context.Context) ([]gomux.Session, error):
//   Implementation approach:
//   1. Build command: exec.CommandContext(ctx, c.tmuxPath, "list-sessions", "-F", formatString)
//      Format string: "#{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}"
//   2. Run command, capture output
//   3. If exit code non-zero and stderr contains "no server running" or "no sessions"
//      → return empty slice, nil (not an error)
//   4. If other error → return nil, fmt.Errorf("list sessions: %w", err)
//   5. Pass output to parseSessions()
//   6. Return parsed sessions
//
//   Error handling:
//   - tmux not running → empty slice (no error)
//   - tmux command fails → wrapped error
//   - Parse error → propagated from parseSessions

// ExecIntoSession(ctx context.Context, sessionName string) error:
//   Implementation approach:
//   1. Build args: []string{"tmux", "new-session", "-A", "-s", sessionName}
//   2. Get current environment: os.Environ()
//   3. Call syscall.Exec(c.tmuxPath, args, environ)
//   4. If syscall.Exec returns (which means it failed) → return the error
//
//   Note: On success, syscall.Exec never returns (process is replaced).
//   The error return path only executes on failure.
```

**Pattern Application:**

- **Error Handling**: Distinguish between "tmux not running" (expected, return empty) and actual command failures. Use `fmt.Errorf` wrapping for unexpected errors.
- **Context Lifecycle**: `exec.CommandContext` uses the context for cancellation of the list-sessions command. `syscall.Exec` does not use context (process replacement is instant).

### Gate: GREEN

- [x] All existing tests still PASS
- [x] Test command: `go test ./internal/tmux/ -v`
- [x] `go build ./internal/tmux/` compiles without errors
- [x] Client satisfies both gomux.SessionLister and gomux.SessionExecutor interfaces

---

## REFACTOR: Quality

**Focus:** Client code clarity and error handling.

**Review Areas:**

- **Error messages**: Clear distinction between "tmux not found" vs "tmux command failed"
- **Naming**: Format string defined as a package-level constant for readability
- **Simplification**: The "no server running" detection should be robust but simple
- **Documentation**: Godoc on exported methods explaining behavior

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

**Previous:** [Phase 03](03-tmux-parsing.md)
**Next:** [Phase 07](07-main.md) (after Phases 02 and 06 also complete)
