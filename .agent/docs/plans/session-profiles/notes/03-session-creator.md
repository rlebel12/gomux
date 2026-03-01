# Phase 03: Session Creator

## Summary

Added `CreateSessionFromProfile` to `tmux.Client`. The method builds an ordered sequence of tmux subcommands from the profile structure (new-session, new-window, split-window, send-keys, select-window) and executes each via an injected `commandRunner` function type. On failure, a best-effort `kill-session` cleans up the partial session. The default runner wraps `exec.CommandContext`. Tests inject a recording runner that captures calls for assertion.

## Files

- `internal/tmux/tmux.go` - Added `commandRunner` type, `runCmd` field on `Client`, `defaultCommandRunner`, `CreateSessionFromProfile`, `buildSessionCommands`, and `splitFlagFor`
- `internal/tmux/tmux_test.go` - Added `recordingRunner` fake, `TestCreateSessionFromProfile` (8 parameterized cases), and `TestCreateSessionFromProfileKillsOnError`

## Tests

**Test file:** `internal/tmux/tmux_test.go`

**Test count:** 10 new test cases (8 parameterized in `TestCreateSessionFromProfile`, 1 discrete in `TestCreateSessionFromProfileKillsOnError`) plus 9 existing parse session tests

**Behaviors covered:**
- Single window with no command: verify minimal new-session + select-window sequence
- Single window with command: verify send-keys appended after new-session
- Multiple windows: verify new-window used for windows after the first
- Window with panes (horizontal + vertical): verify split-window commands generated per pane
- Horizontal split: verify `-h` flag used for horizontal splits
- Vertical split: verify `-v` flag used for vertical splits
- Default split is vertical: verify empty Split field defaults to `-v`
- Full layout: two windows, first with pane and command, second with command only
- Kill on error: runner fails on 3rd call; verify kill-session -t <name> attempted and error returned

## Implementation Notes

- `buildSessionCommands` is a pure function (no receiver) that constructs the command list from a profile; this makes the logic easy to read independently of the runner injection mechanism.
- `splitFlagFor` uses named constants (`splitHorizontal`, `splitVertical`) to avoid magic strings.
- `defaultCommandRunner` uses `CombinedOutput` to capture both stdout and stderr and includes them in the error for actionable diagnostics.
- Test uses direct struct literal field assignment (`&Client{runCmd: runner.run}`) since `tmux_test.go` is in package `tmux` (white-box access).

## Issues

<!-- Problems encountered and how they were resolved -->
