# Phase 05: App Integration

## Summary

Verification phase confirming all components from Phases 01-04 integrate correctly. The full test suite across gomux, profile, tmux, and tui packages passed without any gaps. The REFACTOR pass fixed lint issues across the codebase: a dynamic error in `App.Run`, test sentinel errors in `gomux_test.go`, line-length violations in `gomux_test.go` and `tmux.go`, an exhaustive switch linter configuration issue, nil-vs-empty slice consistency in `profile/loader.go`, and a nestif complexity violation in `tui/model_test.go`.

## Files

- `internal/gomux/gomux.go` — Added `ErrProfileNotFound` sentinel; replaced dynamic error with wrapped sentinel
- `internal/gomux/gomux_test.go` — Added package-level sentinel errors; wrapped `fakePicker.Pick` signature to fit line length
- `internal/profile/loader.go` — Changed `convertPanes` to return empty slice instead of nil for consistency
- `internal/tmux/tmux.go` — Extracted long `if` condition into a named boolean variable for readability
- `internal/tui/model_test.go` — Extracted `verifySubmitResult` helper to reduce nesting complexity
- `.golangci.yml` — Added `exhaustive.default-signifies-exhaustive: true` so `default:` cases satisfy exhaustive checks

## Tests

Existing test suites verified across all packages before any changes:
- `internal/gomux`: 15 test cases (10 parameterized + 5 discrete)
- `internal/profile`: 13 test cases (8 success + 5 error)
- `internal/tmux`: 13 test cases (5 parse + 8 session creator)
- `internal/tui`: 26 test cases across navigation, actions, input, profiles

All 67 tests pass before and after refactoring.

## Implementation Notes

- The `exhaustive` linter in golangci-lint v2 defaults `default-signifies-exhaustive` to `false`, which caused it to flag `switch msg.Type` statements even though they had `default:` cases. Fixed by adding the setting to `.golangci.yml`.
- `noopProfileLoader` in `main.go` returns `nil, nil` (nil slice). `App.Run` handles this correctly since `len(nil) == 0` in Go — no code change needed.
- `ErrProfileNotFound` is exported so callers can test for it with `errors.Is` if needed.

## Issues

None.
