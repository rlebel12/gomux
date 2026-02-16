# Phase 03: Tmux Session Parsing

## Summary

Implemented `parseSessions` function to parse tab-delimited tmux output into Session structs. The parser handles session name, window count, attached flag (0/1), and Unix timestamp conversion. Includes comprehensive error handling for malformed input with line number context.

## Files

- `internal/tmux/tmux.go` (created): Implements parseSessions function for tmux output parsing
- `internal/tmux/tmux_test.go` (created): Test suite for parseSessions with happy path and error cases

## Tests

- Test file: `internal/tmux/tmux_test.go`
- 9 test cases total (5 happy path, 4 error cases)
- Happy path coverage: single detached session, single attached session, multiple sessions with mixed states, empty output, whitespace-only output
- Error coverage: missing fields, non-numeric window count, non-numeric timestamp, invalid attached flag (not 0 or 1)

## Implementation Notes

- Extracted field indices and count into named constants during refactor phase for better maintainability
- Comments updated to say "Parse" instead of "Field N" for consistency

## Issues

<!-- Problems encountered and how they were resolved -->
