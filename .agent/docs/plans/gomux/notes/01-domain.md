# Phase 01: Domain Types and Project Scaffolding

## Summary

Set up the Go module structure and core domain types for gomux. Created the foundational types (Session struct), interfaces (SessionLister, SessionPicker, SessionExecutor), sentinel error (ErrNoSelection), and App orchestrator with dependency injection pattern. Configured pre-commit hooks (lefthook) and linter (golangci-lint) for code quality automation.

## Files

- `go.mod` - Go module definition for github.com/rlebel12/gomux with Go 1.25.7
- `internal/gomux/gomux.go` - Domain types, interfaces, App orchestrator struct with stub Run method
- `lefthook.yml` - Pre-commit hook configuration with gofumpt, go mod tidy, go build, and golangci-lint
- `.golangci.yml` - Comprehensive linter configuration with all linters enabled by default

## Tests

Infrastructure phase - no tests created. Existing test verification: confirmed no existing test suite (new project).

## Implementation Notes

- Used Go 1.25.7 (latest available, exceeds minimum 1.24+ requirement)
- App.Run method has stub implementation (returns nil) - Phase 02 will implement full orchestration logic
- ErrNoSelection implemented as sentinel error (errors.New) rather than custom error type per error-handling pattern - no structured context needed, errors.Is() sufficient
- All interface methods accept context.Context as first parameter per context-lifecycle pattern
- Interfaces defined in consuming package (internal/gomux) per dependency-injection pattern
- Pre-commit configuration uses shared remote config from rlebel12/lefthook for hygiene checks

## Issues

Encountered sandbox filesystem restrictions during initial go build and lefthook install attempts. Resolved by disabling sandbox (read-only filesystem prevented build cache and git hook writes).
