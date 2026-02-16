# Phase 02: App Orchestrator

## Summary

Implemented the App.Run orchestrator that routes between three modes: direct attach (CLI bypass), empty state auto-create, and TUI selection. Handles user quit gracefully and propagates errors from dependencies.

## Files

- internal/gomux/gomux.go: Implemented App.Run method with mode routing logic
- internal/gomux/gomux_test.go: Created comprehensive test suite with fakes for all dependencies

## Tests

8 tests created in internal/gomux/gomux_test.go:
- 5 parameterized tests for routing logic (direct mode, empty state, TUI single/multi session, user quit)
- 3 discrete tests for error propagation (lister error, executor error, session passing verification)

## Implementation Notes

Refactored to extract execSession helper method to eliminate duplication of executor calls with error wrapping. Updated Run documentation to match implementation. The three modes (direct, empty, TUI) are now visually distinct with clear comments.

## Issues

<!-- Problems encountered and how they were resolved -->
