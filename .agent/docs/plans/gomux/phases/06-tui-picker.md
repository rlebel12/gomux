# Phase 06: TUI Picker and Styling

**Depends on:** Phase 05
**Phase Type:** Infrastructure
**Status:** Pending

<!-- Infrastructure phase: GREEN-REFACTOR only -->
<!-- Picker.Pick is thin delegation to tea.Program. -->
<!-- View rendering and lipgloss styling are visual — tested manually. -->
<!-- Status progression: Pending -> GREEN -> Complete -->

---

## Existing Test Verification

**Run existing tests:** `go test ./internal/tui/ -v`

**Expected result:** All Model tests from Phase 05 continue to pass after adding Picker and styling.

### Gate: EXISTING TESTS

- [ ] Existing test suite executed
- [ ] All tests PASS before proceeding

---

## GREEN: Implement

**Objective:** Implement the `Picker` type, the `View()` method with lipgloss styling, and the styles constants.

**Files:**

- `internal/tui/picker.go` — Picker struct, NewPicker, Pick method
- `internal/tui/styles.go` — lipgloss style definitions
- `internal/tui/model.go` — complete the View() method

**Implementation Guidance:**

```go
// picker.go

// Picker struct:
//   - Empty struct (no state needed)
//   - Satisfies gomux.SessionPicker

// NewPicker() *Picker:
//   Return &Picker{}

// Pick(ctx context.Context, sessions []gomux.Session) (string, error):
//   Implementation approach:
//   1. Create model via newModel(sessions)
//   2. Create tea.Program with model and tea.WithAltScreen()
//   3. Run the program: finalModel, err := p.Run()
//   4. If err → return "", fmt.Errorf("run picker: %w", err)
//   5. Type-assert finalModel to Model
//   6. If model.Quitting() → return "", gomux.ErrNoSelection
//   7. Return model.Result(), nil
```

```go
// styles.go

// Define lipgloss styles as package-level variables:
//   - titleStyle: subtle, for "gomux" header
//   - selectedStyle: bold, for the selected session name
//   - normalStyle: default weight, for unselected session names
//   - metadataStyle: dim/muted, for the second line (windows, attached, time)
//   - attachedStyle: distinct color (green or yellow), for "attached" badge
//   - footerStyle: dim, for the help bar
//   - errorStyle: red foreground, for validation errors
//   - warningStyle: yellow foreground, for validation warnings
//   - cursorGlyph: "▸ " for selected, "  " for unselected
```

```go
// model.go — View() method

// View() string:
//   Implementation approach:
//   1. If inputMode → renderInputView()
//   2. Else → renderListView()
//
//   renderListView():
//   1. Render title "gomux" with titleStyle
//   2. For each session:
//      a. Line 1: cursor glyph + session name (selectedStyle or normalStyle)
//      b. Line 2: indented metadata line
//         - "{N} window" or "{N} windows" (singular/plural)
//         - " · attached" (only if session.Attached, use attachedStyle)
//         - " · created {relative time}" (e.g., "2h ago", "1d ago")
//      c. Blank line between sessions
//   3. Footer: "↑↓ navigate · enter attach · n new · q quit" with footerStyle
//
//   renderInputView():
//   1. Render title "gomux" with titleStyle
//   2. "New session name:" label
//   3. Text input component view
//   4. If err non-empty → render error with errorStyle
//   5. If warning non-empty → render warning with warningStyle
//   6. Footer: "enter confirm · esc cancel" with footerStyle
//
//   relativeTime(created time.Time) string:
//   Implementation approach:
//   1. Calculate duration since created
//   2. If < 1 minute → "just now"
//   3. If < 1 hour → "{N}m ago"
//   4. If < 1 day → "{N}h ago"
//   5. If < 1 week → "{N}d ago"
//   6. Else → "{N}w ago"
```

**Note on `relativeTime`:** This function has branching logic (minute/hour/day/week thresholds). If the implementer finds the logic non-trivial, consider adding parameterized tests for it -- either here or as a small addition to the Phase 05 test file. The function should accept a `now` parameter (or use a clock interface) to make it testable without real time.

**Pattern Application:**

- **Dependency Injection**: Picker receives sessions as a parameter, not via constructor. It creates the internal Model and delegates to Bubbletea.
- **Error Handling**: Bubbletea program errors are wrapped. ErrNoSelection is returned when the model's quit flag is set.

### Gate: GREEN

- [ ] All existing tests still PASS
- [ ] Test command: `go test ./internal/tui/ -v`
- [ ] `go build ./internal/tui/` compiles without errors
- [ ] Picker satisfies the gomux.SessionPicker interface
- [ ] View renders correctly for list view and input view (manual verification)

---

## REFACTOR: Quality

**Focus:** View rendering clarity and style consistency.

**Review Areas:**

- **Duplication**: Extract common rendering helpers (title, footer)
- **Naming**: Style variables clearly convey their purpose
- **Simplification**: relativeTime function should be clean and readable
- **Readability**: View methods should be easy to follow top-to-bottom

### Gate: REFACTOR

- [ ] Reviewed for code duplication and extracted common patterns
- [ ] Variable and function names are clear and descriptive
- [ ] Complex logic simplified where possible
- [ ] Error messages are helpful and actionable

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Phase 07 depends on this phase

---

**Previous:** [Phase 05](05-tui-model.md)
**Next:** [Phase 07](07-main.md) (after Phases 02 and 04 also complete)
