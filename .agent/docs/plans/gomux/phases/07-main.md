# Phase 07: Main Entrypoint

**Depends on:** Phase 02, Phase 04, Phase 06
**Phase Type:** Infrastructure
**Status:** Complete

<!-- Infrastructure phase: GREEN-REFACTOR only -->
<!-- main.go is pure wiring — no testable logic -->
<!-- Status progression: Pending -> GREEN -> Complete -->

---

## Existing Test Verification

**Run existing tests:** `go test ./... -v`

**Expected result:** All tests across all packages pass. This is the integration gate -- every package compiles together.

### Gate: EXISTING TESTS

- [x] Existing test suite executed (`go test ./...`)
- [x] All tests PASS before proceeding

---

## GREEN: Implement

**Objective:** Wire all concrete dependencies in `cmd/gomux/main.go` and verify the binary builds.

**Files:**

- `cmd/gomux/main.go`

**Implementation Guidance:**

```go
// cmd/gomux/main.go
package main

// Implementation approach:
// 1. Create root context with signal.NotifyContext(ctx, os.Interrupt)
//    - Defer cancel()
// 2. Create tmux.NewClient()
//    - If error → fmt.Fprintf(os.Stderr, "gomux: %v\n", err), os.Exit(1)
// 3. Extract sessionName from os.Args
//    - If len(os.Args) > 1 → sessionName = os.Args[1]
//    - Else → sessionName = "" (TUI mode)
// 4. Create gomux.New(client, tui.NewPicker(), client)
//    - Note: client satisfies both SessionLister and SessionExecutor
// 5. Call app.Run(ctx, sessionName)
//    - If error → fmt.Fprintf(os.Stderr, "gomux: %v\n", err), os.Exit(1)
//
// Error handling:
// - tmux not found → clear message: "gomux: tmux not found on PATH"
// - Run errors → prefixed with "gomux: "
//
// Context pattern:
// - signal.NotifyContext handles SIGINT for clean TUI shutdown
// - Context flows through to ListSessions (for CommandContext cancellation)
```

**Pattern Application:**

- **Dependency Injection**: All wiring in main.go. Constructor injection: concrete implementations passed to gomux.New.
- **Context Lifecycle**: Root context from `signal.NotifyContext` with `os.Interrupt`. Deferred cancel for cleanup.
- **Error Handling**: Fail fast on tmux not found. Print errors to stderr with "gomux:" prefix.

### Gate: GREEN

- [x] All existing tests still PASS (`go test ./...`)
- [x] Binary builds: `go build -o /tmp/claude/gomux ./cmd/gomux/`
- [x] `go vet ./...` passes
- [x] Implementation matches the wiring pattern from README.md

---

## REFACTOR: Quality

**Focus:** Entrypoint clarity — main.go should be immediately readable.

**Review Areas:**

- **Simplification**: main.go should be ~25 lines, no more
- **Error messages**: User-facing error messages are clear and actionable
- **Naming**: Variable names match their purpose (client, picker, app)
- **Documentation**: Brief package comment

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
3. All phases complete -- implementation is done

---

**Previous:** [Phase 02](02-orchestrator.md), [Phase 04](04-tmux-client.md), [Phase 06](06-tui-picker.md)
**Next:** Final phase
