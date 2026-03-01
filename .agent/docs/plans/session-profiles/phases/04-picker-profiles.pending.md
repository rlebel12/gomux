# Phase 04: Picker Profile Section

**Depends on:** Phase 01
**Phase Type:** Standard

---

## RED: Write Tests

**Objective:** Verify the TUI picker displays profiles in a separate section, handles cross-section navigation, and returns the correct Selection type.

**Files:**

- `internal/tui/model_test.go`

**Context:** The existing picker tests use `key()` and `keyRune()` helpers and `testSessions()` for sample data. This phase updates the `Picker` to accept profiles alongside sessions and adds a second section with a divider. The `Result()` method changes to return a `Selection` instead of `string`.

**Cursor semantics:** The model uses a single flat integer cursor spanning both sections. Index `0..len(sessions)-1` = sessions, `len(sessions)..len(sessions)+len(profiles)-1` = profiles. Test assertions use the raw integer value with a comment indicating which section item it maps to.

**New test helper:**

- `testProfiles() []gomux.Profile` — returns 2 sample profiles for testing

**Parameterized Tests — navigation with profiles (`TestListViewNavigationWithProfiles`):**

| Case | Sessions | Profiles | Keys | Expected Cursor | Notes |
|------|----------|----------|------|-----------------|-------|
| `sessions_only_navigates_normally` | 3 | 0 | `j` | `1` // sessions[1] | No profile section when empty |
| `profiles_navigate_down_into_profiles` | 2 | 2 | `j,j` | `3` // profiles[1] | Cursor enters profile section at index 2, moves to 3 |
| `profiles_navigate_up_from_profiles` | 2 | 2 | `j,j,k` | `2` // profiles[0] | Cursor moves back within profiles |
| `profiles_navigate_up_to_sessions` | 2 | 2 | `j,j,k,k` | `1` // sessions[1] | Cursor crosses back to sessions |
| `profiles_clamp_at_bottom` | 1 | 1 | `j,j,j` | `1` // profiles[0] | max cursor = len(sessions)+len(profiles)-1 = 1 |
| `only_profiles_no_sessions` | 0 | 2 | `j` | `1` // profiles[1] | Navigate within profiles-only list |

**Parameterized Tests — selection with profiles (`TestListViewSelectionWithProfiles`):**

| Case | Sessions | Profiles | Keys | Expected Selection | Notes |
|------|----------|----------|------|--------------------|-------|
| `select_session` | 2 | 2 | `Enter` | `Selection{Name:"session0", IsProfile:false}` | Select first session |
| `select_profile` | 2 | 2 | `j,j,Enter` | `Selection{Name:"profile0", IsProfile:true}` | Navigate to first profile (cursor=2), select |
| `select_only_profile` | 0 | 1 | `Enter` | `Selection{Name:"profile0", IsProfile:true}` | No sessions, select profile |

**Parameterized Tests — view rendering (`TestListViewRenderWithProfiles`):**

| Case | Sessions | Profiles | Expected Contains | Notes |
|------|----------|----------|-------------------|-------|
| `shows_profile_section` | 1 | 1 | `"Profiles"` divider text | Section header rendered |
| `no_profile_section_when_empty` | 1 | 0 | No `"Profiles"` text | Section hidden |
| `profile_shows_window_count` | 0 | 1 (2 windows) | `"2 windows"` | Metadata from profile definition |

**Discrete Tests:**

- **Test input mode returns Selection**: New session name entered via input mode returns `Selection{Name: inputName, IsProfile: false}` — a typed name is always treated as a session, never a profile

Note: The existing quit test (`q` sets quitting flag) already covers quit behavior. No new quit test needed — `q` handling is unchanged by profile integration.

**Assertions:**

- Selection type correctly identifies session vs profile
- Divider is visually present only when profiles exist
- Cursor can traverse both sections smoothly
- Profile metadata shows window count from definition
- Existing session selection behavior preserved

### Gate: RED

- [ ] Test file updated with profile section tests
- [ ] All new tests FAIL (profile section not implemented)
- [ ] Tests cover navigation, selection, rendering, and edge cases

---

## GREEN: Implement

**Objective:** Update the Picker model to support a two-section list with profiles.

**Files:**

- `internal/tui/model.go`
- `internal/tui/styles.go`

**Implementation Guidance — model.go:**

```go
// Updated Model struct:
// Add: profiles []gomux.Profile
// Change: result string → result gomux.Selection
// The cursor spans both sections: 0..len(sessions)-1 = sessions,
// len(sessions)..len(sessions)+len(profiles)-1 = profiles

// Updated newModel(sessions []gomux.Session, profiles []gomux.Profile) *Model
// Accept profiles parameter

// Updated Picker struct and Pick method:
// Pick(ctx context.Context, sessions []gomux.Session, profiles []gomux.Profile) (gomux.Selection, error)
// Pass profiles to newModel

// Result() → gomux.Selection instead of string
// Returns m.result (now a Selection struct)

// handleListKey updates:
// - Navigation clamping: max cursor = len(sessions) + len(profiles) - 1
// - Enter: determine section from cursor position:
//     cursor < len(sessions) → result = Selection{Name: sessions[cursor].Name, IsProfile: false}
//     cursor >= len(sessions) → result = Selection{Name: profiles[cursor-len(sessions)].Name, IsProfile: true}

// handleInputKey Enter branch:
// m.result = Selection{Name: value, IsProfile: false}  // typed name is always a session

// renderListView updates:
// 1. Render session items (existing logic)
// 2. If len(profiles) > 0:
//    a. Render divider line with "Profiles" label
//    b. Render profile items with window count metadata
// 3. Cursor glyph logic: for session items, cursor matches index directly;
//    for profile items, cursor matches len(sessions) + profile index

// formatProfileMetadata(profile gomux.Profile) string
// Returns "N windows" based on len(profile.Windows)
```

**Implementation Guidance — styles.go:**

```go
// Add divider style function
// dividerStyle() — foreground color 240 (gray), matches existing metadata style
```

### Gate: GREEN

- [ ] All tests from RED phase now PASS
- [ ] Existing picker tests pass with updated signatures
- [ ] Test command: `go test ./internal/tui/ -v`
- [ ] Profile section renders correctly with divider

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Cursor logic**: Is section detection clean? Consider a helper `isProfileSection() bool`
- **View rendering**: Profile section rendering doesn't duplicate session rendering logic excessively
- **Divider**: Simple, clear visual separation

### Gate: REFACTOR

- [ ] Cursor section detection is clean and not repeated
- [ ] View rendering for profiles doesn't duplicate session rendering
- [ ] Style functions follow existing naming conventions
- [ ] No unnecessary complexity in navigation logic

---

## Phase Complete

When all gates pass:

1. Rename this file to replace the current status suffix with `.complete.md` using `git mv`
2. Proceed to Phase 05 (after Phases 02, 03 also complete)

---

**Previous:** Phase 01 - Domain Types
**Next:** Phase 05 - App Integration (requires Phases 02, 03, 04)
