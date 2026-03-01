# Phase 01: Domain Types

## Summary

Added Profile, Window, Pane, and Selection domain types to the gomux package. Added ProfileLoader and SessionCreator interfaces. Updated SessionPicker to accept profiles and return Selection. Updated App struct, New constructor, and Run method to support profile-aware routing: loading profiles, filtering profiles that have running sessions, calling the picker with both sessions and profiles, and conditionally calling the creator when a profile is selected.

## Files

- `internal/gomux/gomux.go` — added Profile/Window/Pane/Selection types, ProfileLoader/SessionCreator interfaces, updated SessionPicker signature, updated App struct/New/Run
- `internal/gomux/gomux_test.go` — updated fakes and all test cases for new types/interfaces, added 5 new tests

## Tests

Test file: `internal/gomux/gomux_test.go`

Added/updated test cases: 16 total

- `TestAppRun` parameterized (10 rows): direct mode, empty state, TUI single/multi, user quit, profile selected creates session, session selected skips creation, no profiles empty list, direct mode skips profiles, no sessions but profiles shows picker
- `TestAppRun_ListerErrorPropagation`: lister error propagates, picker/executor not called
- `TestAppRun_ExecutorErrorPropagation`: executor error propagates
- `TestAppRun_PickerReceivesSessions`: picker receives correct sessions and profiles
- `TestAppRun_ProfileLoaderErrorPropagation`: loader error propagates, picker/creator not called
- `TestAppRun_SessionCreatorErrorPropagation`: creator error propagates, executor not called
- `TestAppRun_PickerReceivesFilteredProfiles`: profiles matching existing session names are excluded

## Implementation Notes

- Used `reflect.DeepEqual` for creator.received comparison in tests since `Profile` contains `[]Window` slices (not comparable with `==`).
- `filterProfiles` and `findProfile` are unexported helpers placed after the public App methods, consistent with the public-before-private convention.
- The `direct_mode_skips_profiles` test sets `listerSessions` but the direct mode returns before the lister is called; this field is irrelevant in that row but included for completeness.

## Issues

- **main.go build failure**: Updating `gomux.New` signature broke `cmd/gomux/main.go` which still passed the old 3-argument call. Additionally, `tui.Picker.Pick` still had the old signature `(ctx, sessions) (string, error)` and no longer satisfied `SessionPicker`. Resolved by adding temporary stubs in `main.go`: `legacyPicker` adapter wrapping the old `tui.Picker`, `noopProfileLoader`, and `noopSessionCreator`. Both are marked with `TODO(phase-04)` and `TODO(phase-06)` comments for cleanup in the respective phases.
