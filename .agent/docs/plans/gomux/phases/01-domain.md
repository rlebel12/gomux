# Phase 01: Domain Types and Project Scaffolding

**Depends on:** None
**Phase Type:** Infrastructure
**Status:** Pending

<!-- Infrastructure phase: GREEN-REFACTOR only (no testable logic) -->
<!-- Status progression: Pending -> GREEN -> Complete -->

---

## Existing Test Verification

**Run existing tests:** N/A (new project, no existing tests)

**Expected result:** No tests exist yet. This phase creates the foundational types and interfaces.

### Gate: EXISTING TESTS

- [ ] Confirmed: no existing test suite to break

---

## GREEN: Implement

**Objective:** Set up Go module, domain types, interfaces, sentinel errors, and pre-commit configuration.

**Files:**

- `go.mod` — module definition (`module github.com/rlebel12/gomux`, Go 1.24+)
- `internal/gomux/gomux.go` — domain types and interfaces
- `lefthook.yml` — pre-commit hooks (from `~/.agent/docs/go/references/pre-commit-configuration.md`)
- `.golangci.yml` — linter configuration (from `~/.agent/docs/go/references/golangci-lint-configuration.md`)

**Implementation Guidance:**

```go
// go.mod
// Module: github.com/rlebel12/gomux
// Go version: 1.24 (or latest stable)
// No external dependencies yet — only stdlib

// internal/gomux/gomux.go
package gomux

// Session type:
//   - Name string, Windows int, Attached bool, Created time.Time
//   - Pure data struct, no methods needed

// DefaultSessionName constant:
//   - Value: "default"
//   - Used when no sessions exist and TUI is skipped

// ErrNoSelection sentinel error:
//   - errors.New("no session selected")
//   - Returned by SessionPicker when user quits without selecting

// Three interfaces (defined here, implemented elsewhere):
//   1. SessionLister — ListSessions(ctx) → ([]Session, error)
//   2. SessionPicker — Pick(ctx, []Session) → (string, error)
//   3. SessionExecutor — ExecIntoSession(ctx, string) → error
//   All accept context.Context as first parameter per context-lifecycle pattern

// App struct:
//   - Holds lister, picker, executor as unexported fields
//   - New(lister, picker, executor) → *App constructor
//   - Run(ctx, sessionName) → error method signature (body implemented in Phase 02)
//
// For this phase, Run can return nil as a stub — Phase 02 adds the real logic.
```

**Pre-commit Setup:**

Copy the reference configurations from the Go pattern docs:

- `lefthook.yml` — from `~/.agent/docs/go/references/pre-commit-configuration.md`. Use the standard (non-monorepo) configuration.
- `.golangci.yml` — from `~/.agent/docs/go/references/golangci-lint-configuration.md`. Use the primary pattern as-is. Note: `forbidigo` bans `fmt.Print*` — this is fine since gomux uses `fmt.Fprintf(os.Stderr, ...)` only in `cmd/gomux/main.go` (excluded by the `cmd/.*/main\.go` exclusion rule).
- Run `lefthook install` to activate the hooks.

**Pattern Application:**

- **Dependency Injection**: Interfaces defined in consuming package (`internal/gomux`), not in implementing packages. Constructor accepts interfaces, returns concrete `*App`.
- **Context Lifecycle**: All interface methods accept `context.Context` as first parameter.
- **Error Handling**: `ErrNoSelection` is a sentinel value (not a custom type) since it carries no structured context — `errors.Is()` is sufficient.
- **Pre-commit Configuration**: Lefthook with gofumpt, go mod tidy, go build, and golangci-lint per Go reference docs.

### Gate: GREEN

- [ ] `go mod init` succeeds
- [ ] `go build ./internal/gomux/` compiles without errors
- [ ] Session struct, all three interfaces, App struct, and New constructor are defined
- [ ] ErrNoSelection and DefaultSessionName are exported
- [ ] `lefthook.yml` and `.golangci.yml` are in place
- [ ] `lefthook install` succeeds

---

## REFACTOR: Quality

**Focus:** Code quality of the domain package.

**Review Areas:**

- **Naming**: Interface names follow Go conventions (no `I` prefix, verb-based: Lister, Picker, Executor)
- **Documentation**: Exported types and methods have godoc comments
- **Organization**: Public exports before private fields, types grouped logically

### Gate: REFACTOR

- [ ] All exported types and functions have godoc comments
- [ ] Interface method signatures match the design summary exactly
- [ ] No unnecessary dependencies in go.mod

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Proceed to Phases 02, 03, and 05 (parallel)

---

**Previous:** First phase
**Next:** [Phase 02](02-orchestrator.md), [Phase 03](03-tmux-parsing.md), [Phase 05](05-tui-model.md) (parallel)
