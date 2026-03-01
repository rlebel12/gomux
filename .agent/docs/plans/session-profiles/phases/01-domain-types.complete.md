# Phase 01: Domain Types and Interfaces

**Depends on:** None
**Phase Type:** Standard

---

## RED: Write Tests

**Objective:** Verify that the new domain types exist and the updated App constructor and Run method handle profile-aware routing correctly.

**Files:**

- `internal/gomux/gomux_test.go`

**Context:** The existing `gomux_test.go` already has fakes for `SessionLister`, `SessionPicker`, and `SessionExecutor`. This phase adds fakes for `ProfileLoader` and `SessionCreator`, updates the existing `fakePicker` to return `Selection` instead of `string`, and adds test cases for profile-aware routing in `App.Run`.

**Testing note:** All `context.Context` arguments in tests must use `t.Context()`, never `context.Background()`. The existing `gomux_test.go` uses `context.Background()` on lines 141, 177, 208, and 231 (in `TestAppRun`, `TestAppRun_ListerErrorPropagation`, `TestAppRun_ExecutorErrorPropagation`, `TestAppRun_PickerReceivesSessions`). All four must be replaced in this phase.

**Update to existing `empty_state` test row:** The existing `empty_state` row covers "no sessions → exec DefaultSessionName". This row must be updated to explicitly set `ProfileLoaderResult: []` (empty, no error) so the behavior is unambiguous: no sessions AND no profiles → exec DefaultSessionName. The new `no_sessions_but_profiles_shows_picker` row covers the other branch.

**Migration of existing test cases:**

The existing `fakePicker` changes signature:
- `Pick(ctx, sessions)` → `Pick(ctx, sessions, profiles)`
- Return type: `(string, error)` → `(Selection, error)`
- `received []gomux.Session` field → add `receivedProfiles []gomux.Profile` field to track profile arguments
- Existing `pickerName string` field → `pickerSelection Selection` field

All existing `TestAppRun` table rows update accordingly:
- `pickerName: "dev"` → `pickerSelection: Selection{Name: "dev", IsProfile: false}`
- Same pattern for all existing rows

The discrete test `TestAppRun_PickerReceivesSessions` gains a `wantProfiles` assertion alongside the existing `wantSessions` check.

New fakes:

- `fakeProfileLoader` — fields: `profiles []gomux.Profile`, `err error`, `called bool`
- `fakeSessionCreator` — fields: `received gomux.Profile`, `err error`, `called bool`

**Parameterized Tests** — extend existing `TestAppRun` table:

| Case | SessionName | Sessions | ProfileLoaderResult | PickerSelection | Expected Exec | Expected Create | Notes |
|------|-------------|----------|---------------------|-----------------|---------------|-----------------|-------|
| `profile_selected_creates_session` | `""` | `["existing"]` | `[{Name:"dev", Windows:[{Name:"editor"}]}]` | `Selection{Name:"dev", IsProfile:true}` | `"dev"` | `Profile{Name:"dev",...}` called | Profile selected → create → exec |
| `session_selected_skips_creation` | `""` | `["existing"]` | `[{Name:"dev",...}]` | `Selection{Name:"existing", IsProfile:false}` | `"existing"` | creator not called | Session selected → exec only |
| `no_profiles_file_empty_list` | `""` | `["existing"]` | `[]` (no error) | `Selection{Name:"existing", IsProfile:false}` | `"existing"` | creator not called | Missing file returns empty, picker works normally |
| `direct_mode_skips_profiles` | `"mysession"` | `[]` | `[{Name:"dev",...}]` | n/a (not called) | `"mysession"` | creator not called | Direct `gomux mysession` bypasses everything |
| `no_sessions_but_profiles_shows_picker` | `""` | `[]` | `[{Name:"dev", Windows:[{Name:"editor"}]}]` | `Selection{Name:"dev", IsProfile:true}` | `"dev"` | `Profile{Name:"dev",...}` called | No running sessions, profiles shown in picker |

**Discrete Tests:**

- **Test ProfileLoader error propagation**: ProfileLoader returns error → App.Run returns wrapped error, picker and creator not called
- **Test SessionCreator error propagation**: Profile selected, creator returns error → App.Run returns wrapped error, executor not called
- **Test picker receives filtered profiles**: Profiles whose names match existing session names are excluded from the profiles slice passed to Pick

**Assertions:**

- Creator called/not called based on `IsProfile` flag
- Creator receives correct Profile struct
- Executor receives correct session name
- Profile filtering removes profiles matching existing session names
- Error messages contain context wrapping

### Gate: RED

- [x] Test file updated with new fakes and test cases
- [x] All `context.Background()` calls in `gomux_test.go` replaced with `t.Context()`
- [x] All new tests FAIL (new types/interfaces don't exist yet)
- [x] Existing tests updated for Selection return type (also fail)

---

## GREEN: Implement

**Objective:** Add domain types, interfaces, update App struct/constructor/Run to support profile-aware routing.

**Files:**

- `internal/gomux/gomux.go`

**Implementation Guidance:**

```go
// New types — add after existing Session type

type Profile struct {
    Name    string
    Windows []Window
}

type Window struct {
    Name    string
    Dir     string
    Command string
    Panes   []Pane
}

type Pane struct {
    Dir     string
    Command string
    Split   string // "horizontal" or "vertical"
}

type Selection struct {
    Name      string
    IsProfile bool
}
```

```go
// New interfaces — add after existing interfaces

type ProfileLoader interface {
    LoadProfiles(ctx context.Context) ([]Profile, error)
}

type SessionCreator interface {
    CreateSessionFromProfile(ctx context.Context, profile Profile) error
}
```

```go
// Updated SessionPicker — signature changes
type SessionPicker interface {
    Pick(ctx context.Context, sessions []Session, profiles []Profile) (Selection, error)
}
```

```go
// Updated App struct — add profiles and creator fields
// Updated New() — accept ProfileLoader and SessionCreator parameters
// Constructor signature:
func New(
    lister SessionLister,
    picker SessionPicker,
    executor SessionExecutor,
    profiles ProfileLoader,
    creator SessionCreator,
) *App
```

```go
// Updated Run() method — routing logic:
//
// 1. If sessionName provided → exec directly (unchanged)
// 2. Load profiles via ProfileLoader
// 3. If no sessions AND no profiles → exec DefaultSessionName (unchanged)
// 4. Filter profiles: exclude any whose Name matches an existing session
// 5. Pick(ctx, sessions, filteredProfiles) → Selection
// 6. If Selection.IsProfile:
//    a. Find profile by name from loaded profiles
//    b. CreateSessionFromProfile(ctx, profile)
//    c. ExecIntoSession(ctx, selection.Name)
// 7. If !Selection.IsProfile:
//    a. ExecIntoSession(ctx, selection.Name) (unchanged)
```

### Gate: GREEN

- [x] All tests from RED phase now PASS
- [x] All existing tests pass with updated signatures
- [x] Test command: `go test ./internal/gomux/ -v`
- [x] Implementation follows routing logic above

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Profile filtering**: Ensure the filtering logic is clean — consider a helper function if it's more than a few lines
- **Error wrapping**: Use concise verb+object format (`"load profiles"`, `"create session"`)
- **Type placement**: New types defined before interfaces, interfaces before App struct, following the public-before-private convention

### Gate: REFACTOR

- [x] Reviewed for code duplication and extracted common patterns
- [x] Variable and function names are clear and descriptive
- [x] Error messages use concise verb+object format
- [x] Types and interfaces ordered logically in the file

---

## Phase Complete

When all gates pass:

1. Rename this file to replace the current status suffix with `.complete.md` using `git mv`
2. Proceed to Phases 02, 03, 04 (parallel)

---

**Previous:** First phase
**Next:** Phase 02 - Profile Loader, Phase 03 - Session Creator, Phase 04 - Picker Profiles (parallel)
