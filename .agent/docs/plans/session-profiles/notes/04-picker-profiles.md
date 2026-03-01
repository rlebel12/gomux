# Phase 04: Picker Profiles

## Summary

Updated the TUI picker to support a two-section list: existing tmux sessions above a divider, and profiles below. The `Model` struct gained a `profiles` field and changed its `result` from `string` to `gomux.Selection`. Navigation, selection, and rendering all span both sections using a flat cursor index. The `Picker.Pick` method signature was updated to match the `gomux.SessionPicker` interface.

## Files

- `internal/tui/model.go` — Updated Model struct, newModel signature, Result type, cursor logic, and renderListView with profile section
- `internal/tui/picker.go` — Updated Pick signature to accept profiles and return gomux.Selection
- `internal/tui/styles.go` — Added dividerStyle for the "── Profiles ──" section header
- `internal/tui/model_test.go` — Updated existing tests for new signatures; added navigation, selection, rendering, and input mode tests with profiles
- `cmd/gomux/main.go` — Removed legacyPicker shim; wired tui.NewPicker() directly

## Tests

**Test file:** `internal/tui/model_test.go`

**10 test functions covering:**
- `TestListViewNavigation` (6 cases) — cursor movement with sessions only (updated to new signature)
- `TestListViewActions` (4 cases) — Enter/quit/input mode with sessions (updated to Selection type)
- `TestInputView` (3 cases) — input submission, escape, ctrl-c (updated to Selection type)
- `TestNameValidation` (5 cases) — valid name, dot, colon, duplicate, empty (updated to Selection type)
- `TestListViewNavigationWithProfiles` (6 cases) — cross-section navigation with profiles
- `TestListViewSelectionWithProfiles` (3 cases) — Selection type for session vs profile
- `TestListViewRenderWithProfiles` (3 cases) — view contains/excludes profile section
- `TestInputModeReturnsSelection` (1 case) — typed name yields Selection{IsProfile:false}

## Implementation Notes

- Added `initialCursor` field to `TestListViewNavigationWithProfiles` test struct. The phase plan's key sequences for cross-section navigation tests (`j,j` → cursor=3, `j,j,k` → cursor=2, `j,j,k,k` → cursor=1) only hold when starting from `cursor=1`, not `cursor=0`. The plan's notes "Cursor enters profile section at index 2, moves to 3" describe the two-step journey: enter-profiles, then move-within-profiles. The `sessions_only`, `profiles_clamp_at_bottom`, and `only_profiles_no_sessions` cases start from cursor=0.
- `legacyPicker` shim in `cmd/gomux/main.go` removed and replaced with direct `tui.NewPicker()` usage, as the phase TODO instructed. The `tui.Picker` now satisfies the `gomux.SessionPicker` interface directly.
- `renderListItem` helper extracted from `renderListView` to share cursor/style logic between session and profile items, satisfying the refactor goal of no duplication.

## Issues

<!-- Problems encountered and how they were resolved -->
