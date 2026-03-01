# Phase 02: Profile Loader

## Summary

Implemented a YAML file-based `ProfileLoader` in the `internal/profile` package. The `Loader` struct reads a YAML file at a configured path, unmarshals it into unexported YAML-specific types, validates that all profiles have names and windows and all windows have names, then converts to domain types while expanding tilde in all `Dir` fields. Missing config files return an empty slice without error for backward compatibility.

## Files

- `internal/profile/loader.go` — `Loader` struct implementing `gomux.ProfileLoader` with YAML parsing, validation, tilde expansion, and domain type conversion
- `internal/profile/loader_test.go` — parameterized tests for parsing and validation
- `go.mod` / `go.sum` — added `gopkg.in/yaml.v3` dependency

## Tests

- Test file: `internal/profile/loader_test.go`
- 8 parameterized cases in `TestLoaderLoadProfiles`: minimal profile, full profile with all fields, multiple profiles, pane splits, empty list, file not found, tilde expansion, absolute path unchanged
- 5 parameterized cases in `TestLoaderLoadProfilesErrors`: invalid YAML, missing profile name, missing windows, empty window list, window missing name
- Total: 13 test cases

## Implementation Notes

- Added `gopkg.in/yaml.v3` dependency via `go get` since it was not in go.mod.
- The `err113` linter required validation errors to use sentinel errors wrapped with `%w` rather than plain `fmt.Errorf` with dynamic strings. Two sentinels were added: `errMissingName` and `errMissingWindows`. Error messages still contain the profile index/name and window index as context prefix (e.g., `profile "dev": missing windows`), satisfying both the plan's format spec and the linter.
- The `paralleltest` linter required `t.Parallel()` in both the top-level test functions and in each `t.Run` subtest closure.
- The `gocritic` linter flagged `filepath.Join(home, "code/project")` because the second argument contains a path separator. Fixed by splitting into `filepath.Join(home, "code", "project")`.
- `convertPanes` returns `nil` (not an empty slice) when panes is empty, to avoid unintended non-nil empty slice in domain types.

## Issues

<!-- Problems encountered and how they were resolved -->
