# Phase 05: TUI Model

**Depends on:** Phase 01
**Phase Type:** Standard
**Status:** Pending

<!-- Standard phase: Full RED-GREEN-REFACTOR -->
<!-- Status progression: Pending -> RED -> GREEN -> Complete -->

---

## RED: Write Tests

**Objective:** Test the Bubbletea Model's `Update` function for key handling, state transitions, and name validation. Tests drive the Model through message sequences and assert on resulting state.

**Files:**

- `internal/tui/model_test.go`

**Testing Approach:**

The Bubbletea `Update(msg) (Model, Cmd)` function is pure — given a model state and a message, it returns a new model state. This makes it directly unit-testable without running a `tea.Program`. Tests construct a model, send key messages, and assert on the model's state fields.

```go
// Helper to create a tea.KeyMsg for testing:
// key(k tea.KeyType) tea.Msg → tea.KeyMsg{Type: k}
// rune(r rune) tea.Msg → tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
```

**List View Navigation Tests** (table format -- MUST use `t.Run`):

| Case | Initial Cursor | Key Sequence | Expected Cursor | Notes |
|------|---------------|--------------|-----------------|-------|
| `move_down` | 0 | `j` | 1 | j moves cursor down |
| `move_up` | 1 | `k` | 0 | k moves cursor up |
| `arrow_down` | 0 | `Down` | 1 | Arrow key alternative |
| `arrow_up` | 1 | `Up` | 0 | Arrow key alternative |
| `wrap_bottom` | 2 (last) | `j` | 2 | Does not wrap past end |
| `wrap_top` | 0 | `k` | 0 | Does not wrap past start |

**List View Action Tests** (table format -- MUST use `t.Run`):

| Case | Key | Expected | Notes |
|------|-----|----------|-------|
| `enter_selects` | `Enter` | Result is sessions[cursor].Name | Attaches to selected session |
| `q_quits` | `q` | Result signals quit (ErrNoSelection) | Clean exit |
| `ctrl_c_quits` | `Ctrl+C` | Result signals quit | Alternative quit |
| `n_switches_to_input` | `n` | View mode changes to input | Switches to name entry |

**Input View Tests** (table format -- MUST use `t.Run`):

| Case | Typed Text | Key | Expected | Notes |
|------|-----------|-----|----------|-------|
| `submit_name` | `"my-session"` | `Enter` | Result is `"my-session"` | Confirms new session name |
| `esc_returns_to_list` | `"partial"` | `Esc` | View mode back to list, input cleared | Cancels input |
| `ctrl_c_in_input` | `"partial"` | `Ctrl+C` | Result signals quit | Quits from input view |

**Name Validation Tests** (table format -- MUST use `t.Run`):

| Case | Typed Text | Existing Sessions | Expected State | Notes |
|------|-----------|-------------------|----------------|-------|
| `valid_name` | `"new-session"` | `["dev","api"]` | No error, no warning, Enter submits | Happy path |
| `name_with_dot` | `"my.session"` | `[]` | Error message shown, Enter blocked | Dot is invalid in tmux |
| `name_with_colon` | `"my:session"` | `[]` | Error message shown, Enter blocked | Colon is invalid in tmux |
| `duplicate_name` | `"dev"` | `["dev","api"]` | Warning shown, Enter still allowed | Will attach, not create |
| `empty_name` | `""` | `[]` | Enter blocked | Cannot submit empty name |

**Assertions:**

- Cursor position changes correctly for navigation keys
- Enter in list view returns the session name at the cursor position
- q and Ctrl+C in list view signal quit (model's quit flag set, tea.Quit returned)
- n in list view transitions to input view
- Enter in input view returns the typed name (when valid)
- Esc in input view returns to list view and clears input
- Validation: dot/colon names set error state, duplicate names set warning state
- Validation: error state blocks submission, warning state does not

### Gate: RED

- [ ] Test file created with all enumerated test cases
- [ ] All tests FAIL (Model type does not exist yet)
- [ ] Test coverage includes navigation, selection, input, and validation

---

## GREEN: Implement

**Objective:** Implement the Bubbletea Model with two-view state machine (list view and input view), key handling, and name validation.

**Files:**

- `internal/tui/model.go`

**Implementation Guidance:**

```go
// Model struct fields:
//   - sessions []gomux.Session    — the session list to display
//   - cursor int                  — selected index in list view
//   - inputMode bool              — true when in name-input view
//   - textInput textinput.Model   — bubbles text input component
//   - result string               — chosen session name (set on selection/submission)
//   - quitting bool               — true when user wants to quit
//   - err string                  — validation error message (dots/colons)
//   - warning string              — validation warning message (duplicate name)

// newModel(sessions []gomux.Session) Model:
//   Implementation approach:
//   1. Create textinput.Model with placeholder "session name"
//   2. Initialize cursor at 0
//   3. Store sessions for validation lookups
//   4. Return configured Model

// Init() tea.Cmd:
//   Return nil (no initial command needed — list view is static)

// Update(msg tea.Msg) (tea.Model, tea.Cmd):
//   Implementation approach:
//   1. Switch on msg type
//   2. For tea.KeyMsg:
//      a. If inputMode → handleInputKey(msg)
//      b. Else → handleListKey(msg)
//   3. Return updated model and any command
//
//   handleListKey(msg tea.KeyMsg):
//   - j/Down → cursor++ (clamped to len(sessions)-1)
//   - k/Up → cursor-- (clamped to 0)
//   - Enter → set result to sessions[cursor].Name, return tea.Quit
//   - n → set inputMode=true, focus text input
//   - q → set quitting=true, return tea.Quit
//   - Ctrl+C → set quitting=true, return tea.Quit
//
//   handleInputKey(msg tea.KeyMsg):
//   - Enter → if valid (no error, non-empty): set result to input value, return tea.Quit
//   - Esc → set inputMode=false, reset text input, clear err/warning
//   - Ctrl+C → set quitting=true, return tea.Quit
//   - Other → forward to textinput.Update, then validate:
//     * Check for dots/colons → set err
//     * Check for duplicate name → set warning
//     * Otherwise → clear err and warning

// View() string:
//   Stub for this phase — return "" or minimal output.
//   Full view rendering happens in Phase 06.

// Exported accessors for test assertions:
//   - Result() string — returns the chosen session name
//   - Quitting() bool — returns whether user quit without selecting
```

**Pattern Application:**

- **Dependency Injection**: The Model receives `[]gomux.Session` at construction, not a lister interface. The Picker.Pick method (Phase 06) calls the lister and passes results to the model.
- **Error Handling**: Validation errors are UI-local state (err/warning strings), not Go errors. The only Go error path is ErrNoSelection when quitting.

### Gate: GREEN

- [ ] All tests from RED phase now PASS
- [ ] Test command: `go test ./internal/tui/ -v` (18 tests)
- [ ] Implementation follows pseudocode logic flow

---

## REFACTOR: Quality

**Focus:** Model code organization and clarity.

**Review Areas:**

- **Duplication**: Extract shared quit logic between list and input views
- **Naming**: View mode flag and state fields clearly convey purpose
- **Simplification**: Key handling should be readable switch/case blocks
- **Validation**: Name validation logic should be a single, clear function

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
3. Proceed to Phase 06

---

**Previous:** [Phase 01](01-domain.md)
**Next:** [Phase 06](06-tui-picker.md)
