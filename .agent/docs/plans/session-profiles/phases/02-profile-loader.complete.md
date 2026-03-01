# Phase 02: YAML Profile Loader

**Depends on:** Phase 01
**Phase Type:** Standard

---

## RED: Write Tests

**Objective:** Verify YAML profile parsing, file handling, and tilde expansion in the new `internal/profile` package.

**Files:**

- `internal/profile/loader_test.go`

**Test Strategy:** The loader reads from a filesystem path. For testability, tests write YAML to temp files and pass the path. The `Loader` struct accepts a config path via constructor — tests provide temp file paths, production uses `~/.config/gomux/profiles.yml`.

**Parameterized Tests — YAML parsing (`TestLoaderLoadProfiles`):**

| Case | YAML Content | Expected Profiles | Notes |
|------|-------------|-------------------|-------|
| `single_profile_minimal` | `profiles:\n  - name: dev\n    windows:\n      - name: main` | `[{Name:"dev", Windows:[{Name:"main"}]}]` | Minimal valid profile |
| `single_profile_full` | Full YAML with dir, command, panes | Profile with all fields populated | All fields round-trip correctly |
| `multiple_profiles` | Two profiles in list | Two Profile structs | Multiple profiles parsed |
| `pane_splits` | Window with horizontal and vertical panes | Panes with correct Split values | Split direction preserved |
| `empty_profiles_list` | `profiles: []` | `[]` (empty slice) | Valid YAML, no profiles |
| `file_not_found` | n/a (no file created) | `[]` (empty slice, NO error) | Missing file is not an error |
| `tilde_in_dir` | `dir: ~/code/project` | Dir expanded to `$HOME/code/project` | Tilde expansion happens at load time |
| `no_tilde` | `dir: /absolute/path` | Dir unchanged | Absolute paths left alone |

**Parameterized Tests — YAML errors (`TestLoaderLoadProfilesErrors`):**

| Case | YAML Content | Expected Error Contains | Notes |
|------|-------------|------------------------|-------|
| `invalid_yaml` | `profiles:\n  - name: [broken` | `"parse profiles"` | Malformed YAML |
| `missing_name` | `profiles:\n  - windows:\n      - name: main` | `"profile 0: missing name"` | Profile without name field (indexed) |
| `missing_windows` | `profiles:\n  - name: dev` | `"profile \"dev\": missing windows"` | Profile without windows |
| `empty_window_list` | `profiles:\n  - name: dev\n    windows: []` | `"profile \"dev\": missing windows"` | Empty windows list same as missing |
| `window_missing_name` | `profiles:\n  - name: dev\n    windows:\n      - dir: /tmp` | `"profile \"dev\", window 0: missing name"` | Window without name (profile+window indexed) |

**Assertions:**

- Parsed profiles match expected structs (deep equality)
- Missing file returns empty slice, nil error
- Tilde expanded to actual `$HOME` value
- Validation errors contain descriptive messages with profile/window context
- Invalid YAML returns wrapped parse error

### Gate: RED

- [ ] Test file created with all enumerated test cases
- [ ] All tests FAIL (Loader type does not exist yet)
- [ ] Tests use temp files for YAML content

---

## GREEN: Implement

**Objective:** Implement the YAML file loader that satisfies the `ProfileLoader` interface.

**Files:**

- `internal/profile/loader.go`

**Implementation Guidance:**

```go
// Package profile provides YAML-based profile loading.
package profile

// yamlFile is the top-level YAML structure
// Fields: Profiles []yamlProfile `yaml:"profiles"`

// yamlProfile maps a single profile entry
// Fields: Name, Windows []yamlWindow

// yamlWindow maps a window entry
// Fields: Name, Dir, Command, Panes []yamlPane

// yamlPane maps a pane entry
// Fields: Dir, Command, Split

// Loader implements gomux.ProfileLoader
// Fields: path string (config file path)

// NewLoader constructor
// Accept config file path, return *Loader

// LoadProfiles(ctx context.Context) ([]gomux.Profile, error)
// Implementation approach:
// 1. Read file at path — if os.IsNotExist, return empty slice (not error)
// 2. Unmarshal YAML into yamlFile struct
// 3. Validate: each profile must have name and at least one window, each window must have name
// 4. Convert yamlFile → []gomux.Profile, expanding ~ in all Dir fields
// 5. Return profiles

// expandTilde(path string) string
// Replace leading "~/" with os.UserHomeDir() result
// If UserHomeDir fails or path doesn't start with "~/", return unchanged
```

**Validation logic (per-level):**
```
// For each profile (index i):
//   if profile.Name == "" → fmt.Errorf("profile %d: missing name", i)
//   if len(profile.Windows) == 0 → fmt.Errorf("profile %q: missing windows", profile.Name)
//   for each window (index j):
//     if window.Name == "" → fmt.Errorf("profile %q, window %d: missing name", profile.Name, j)
```

**Error wrapping style:**
- `fmt.Errorf("read profiles: %w", err)` for file read errors
- `fmt.Errorf("parse profiles: %w", err)` for YAML unmarshal errors
- Validation errors include profile name/index and window index as shown above

### Gate: GREEN

- [ ] All tests from RED phase now PASS
- [ ] Test command: `go test ./internal/profile/ -v`
- [ ] Missing file returns empty slice, not error
- [ ] Tilde expansion works correctly

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Validation**: Ensure validation errors include enough context (profile index/name, window index/name)
- **Conversion**: YAML types → domain types conversion is clean and doesn't leak YAML tags into domain
- **Tilde expansion**: Simple and correct — no over-engineering for edge cases

### Gate: REFACTOR

- [ ] YAML-specific types are unexported (package-private)
- [ ] Domain types (gomux.Profile etc.) have no YAML tags
- [ ] Error messages include sufficient context for debugging on a phone
- [ ] No unnecessary abstractions

---

## Phase Complete

When all gates pass:

1. Rename this file to replace the current status suffix with `.complete.md` using `git mv`
2. Proceed to Phase 05 (after Phases 03, 04 also complete)

---

**Previous:** Phase 01 - Domain Types
**Next:** Phase 05 - App Integration (requires Phases 02, 03, 04)
