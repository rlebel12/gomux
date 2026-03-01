# Phase 03: Session Creator (tmux.Client)

**Depends on:** Phase 01
**Phase Type:** Standard

---

## RED: Write Tests

**Objective:** Verify that `CreateSessionFromProfile` generates and executes the correct sequence of tmux commands for various profile layouts.

**Files:**

- `internal/tmux/tmux_test.go`

**Test Strategy:** To test session creation without a real tmux server, introduce a `commandRunner` function type injected into `Client`. The default runner wraps `exec.CommandContext`. Tests inject a recording runner that captures the sequence of commands. This tests the public `CreateSessionFromProfile` method through its observable behavior (the commands it runs), not an internal builder function.

**Design change to Client:**

```go
// commandRunner executes a tmux subcommand and returns any error.
type commandRunner func(ctx context.Context, tmuxPath string, args ...string) error

// Client gains a runCmd field, defaulting to a wrapper around exec.CommandContext.
// NewClient sets the default. Tests inject via direct field assignment
// (e.g., client.runCmd = recordingRunner) since tmux_test.go is in package tmux.
```

**Recording fake for tests:**

```go
// recordingRunner captures each call's args for assertion
// Fields: calls [][]string, errOnCall int (return error on Nth call, -1 for never)
```

**Parameterized Tests — command generation via recording runner (`TestCreateSessionFromProfile`):**

| Case | Profile | Expected Command Sequence | Notes |
|------|---------|---------------------------|-------|
| `single_window_no_command` | `{Name:"dev", Windows:[{Name:"main", Dir:"/home/rob"}]}` | `[["new-session","-d","-s","dev","-n","main","-c","/home/rob"], ["select-window","-t","dev:0"]]` | Minimal: one window, no command |
| `single_window_with_command` | `{Name:"dev", Windows:[{Name:"editor", Dir:"/home/rob", Command:"nvim"}]}` | `[["new-session","-d","-s","dev","-n","editor","-c","/home/rob"], ["send-keys","-t","dev:0","nvim","Enter"], ["select-window","-t","dev:0"]]` | send-keys for startup command |
| `multiple_windows` | Two windows | new-session + new-window for second + select-window :0 | Second window uses new-window |
| `window_with_panes` | Window with 2 panes (horiz, vert) | new-session + split-window per pane + send-keys per pane command | Pane splits and commands |
| `horizontal_split` | Pane with `Split:"horizontal"` | includes `["split-window","-h","-t","dev:0","-c","/dir"]` | -h flag for horizontal |
| `vertical_split` | Pane with `Split:"vertical"` | includes `["split-window","-v","-t","dev:0","-c","/dir"]` | -v flag for vertical |
| `default_split_is_vertical` | Pane with `Split:""` | includes `["split-window","-v",...]` | Empty split defaults to vertical |
| `full_layout` | 2 windows, first with 2 panes, all with dirs/commands | Full command sequence | Integration: realistic profile |

**Pane targeting strategy:** After `split-window`, tmux activates the new pane. `send-keys` targets the window (`name:i`) which sends to whichever pane is active — this is the active pane from the most recent split. This matches tmuxinator convention and is correct for sequential splits within a window.

**Discrete Tests:**

- **Test CreateSessionFromProfile kills partial session on error**: Recording runner configured to fail on 3rd call. Verify that after failure, a `kill-session -t <name>` command is attempted, and the method returns a wrapped error.

**Assertions:**

- Command sequences recorded by the runner match expected argument lists exactly
- Window indexing is correct (`:0`, `:1`, etc.)
- `send-keys` targets the window (e.g., `dev:0`), relying on active pane after split
- `select-window` always targets `:0` at the end
- First window created via `new-session`, subsequent via `new-window`
- On mid-sequence failure: `kill-session` attempted, error returned with context

### Gate: RED

- [ ] Test file updated with recording runner and parameterized tests
- [ ] All new tests FAIL (CreateSessionFromProfile does not exist)
- [ ] Test cases cover single window, multi-window, panes, splits, commands, error cleanup

---

## GREEN: Implement

**Objective:** Add session creation capability to tmux.Client using the injected command runner.

**Files:**

- `internal/tmux/tmux.go`

**Implementation Guidance:**

```go
// commandRunner type and Client changes:
//
// type commandRunner func(ctx context.Context, tmuxPath string, args ...string) error
//
// Client struct gains: runCmd commandRunner
// NewClient sets default: runCmd wraps exec.CommandContext
//   default runner: create exec.Command, capture CombinedOutput, return error if exit != 0

// CreateSessionFromProfile(ctx context.Context, profile gomux.Profile) error
// Implementation approach:
// 1. Build command sequence (inline or via helper)
// 2. For each command args:
//    a. Call c.runCmd(ctx, c.tmuxPath, args...)
//    b. If error: attempt c.runCmd(ctx, c.tmuxPath, "kill-session", "-t", profile.Name)
//       (best-effort cleanup, ignore kill errors)
//    c. Return fmt.Errorf("create session %q: %w", profile.Name, err)
// 3. Return nil on success
//
// Command sequence building:
// 1. First window: ["new-session", "-d", "-s", name, "-n", windowName, "-c", dir]
//    - Omit -c if Dir is empty
//    - Omit -n if window Name is empty
// 2. If first window has command: ["send-keys", "-t", "name:0", command, "Enter"]
// 3. For each pane in first window:
//    a. splitFlag: "horizontal" → "-h", anything else → "-v"
//    b. ["split-window", splitFlag, "-t", "name:0"]  (add "-c", dir if Dir set)
//    c. If pane has command: ["send-keys", "-t", "name:0", command, "Enter"]
// 4. For each subsequent window (index i):
//    a. ["new-window", "-t", name, "-n", windowName, "-c", dir]
//    b. If command: ["send-keys", "-t", "name:i", command, "Enter"]
//    c. Panes same as step 3 but targeting "name:i"
// 5. Final: ["select-window", "-t", "name:0"]
```

### Gate: GREEN

- [ ] All tests from RED phase now PASS
- [ ] Existing tmux tests still pass
- [ ] Test command: `go test ./internal/tmux/ -v`
- [ ] Partial session cleanup on error

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Command building**: Consider if a helper for "add send-keys if command present" reduces repetition
- **Error messages**: Include profile name and which command failed
- **Flag mapping**: Split direction → flag mapping is clear and explicit
- **Runner default**: The default `commandRunner` is clean and minimal

### Gate: REFACTOR

- [ ] Command building logic is readable and well-structured
- [ ] No duplicated command construction patterns
- [ ] Error messages include sufficient context
- [ ] Split flag mapping is explicit (no magic strings)
- [ ] Default command runner is clean

---

## Phase Complete

When all gates pass:

1. Rename this file to replace the current status suffix with `.complete.md` using `git mv`
2. Proceed to Phase 05 (after Phases 02, 04 also complete)

---

**Previous:** Phase 01 - Domain Types
**Next:** Phase 05 - App Integration (requires Phases 02, 03, 04)
