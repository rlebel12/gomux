# Phase 07: Main Entrypoint

## Summary

Created the main entrypoint at cmd/gomux/main.go that wires all components together: tmux client, TUI picker, and app orchestrator. The entrypoint follows standard Go patterns with signal-based context cancellation, fail-fast error handling for missing tmux, and consistent error message prefixing.

## Files

- cmd/gomux/main.go: Main entrypoint that wires dependencies and handles signals/errors

## Tests

Existing test suite verified before implementation:
- gomux package: 8 tests (all passing)
- tmux package: 10 tests (all passing)
- tui package: 17 tests (all passing)
Total: 35 tests covering orchestration, parsing, and TUI behavior

## Implementation Notes

- Binary size: 4.4MB (includes bubbletea and dependencies)
- Error messages naturally include context from underlying packages (e.g., tmux.NewClient returns "tmux not found on PATH")
- No additional argument parsing needed beyond extracting the first arg as session name
- Context cancellation from signal.NotifyContext flows through to all operations
- Extracted run() function to avoid os.Exit being called after defer (gocritic lint rule)

## Issues

<!-- Problems encountered and how they were resolved -->
