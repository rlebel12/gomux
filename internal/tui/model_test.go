package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rlebel12/gomux/internal/gomux"
)

// key creates a tea.KeyMsg for a key type.
//
//nolint:ireturn // Test helper must return tea.Msg interface
func key(k tea.KeyType) tea.Msg {
	return tea.KeyMsg{Type: k}
}

// keyRune creates a tea.KeyMsg for a rune.
//
//nolint:ireturn // Test helper must return tea.Msg interface
func keyRune(r rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// testSessions creates sample sessions for testing.
func testSessions() []gomux.Session {
	return []gomux.Session{
		{Name: "dev", Created: time.Now(), Windows: 3, Attached: true},
		{Name: "api", Created: time.Now(), Windows: 2, Attached: false},
		{Name: "test", Created: time.Now(), Windows: 1, Attached: false},
	}
}

func TestListViewNavigation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		notes          string
		keySequence    []tea.Msg
		initialCursor  int
		expectedCursor int
	}{
		{
			name:           "move_down",
			initialCursor:  0,
			keySequence:    []tea.Msg{keyRune('j')},
			expectedCursor: 1,
			notes:          "j moves cursor down",
		},
		{
			name:           "move_up",
			initialCursor:  1,
			keySequence:    []tea.Msg{keyRune('k')},
			expectedCursor: 0,
			notes:          "k moves cursor up",
		},
		{
			name:           "arrow_down",
			initialCursor:  0,
			keySequence:    []tea.Msg{key(tea.KeyDown)},
			expectedCursor: 1,
			notes:          "Arrow key alternative",
		},
		{
			name:           "arrow_up",
			initialCursor:  1,
			keySequence:    []tea.Msg{key(tea.KeyUp)},
			expectedCursor: 0,
			notes:          "Arrow key alternative",
		},
		{
			name:           "wrap_bottom",
			initialCursor:  2,
			keySequence:    []tea.Msg{keyRune('j')},
			expectedCursor: 2,
			notes:          "Does not wrap past end",
		},
		{
			name:           "wrap_top",
			initialCursor:  0,
			keySequence:    []tea.Msg{keyRune('k')},
			expectedCursor: 0,
			notes:          "Does not wrap past start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessions := testSessions()
			model := newModel(sessions)
			model.cursor = tt.initialCursor

			// Apply key sequence
			for _, msg := range tt.keySequence {
				updated, _ := model.Update(msg)

				var ok bool

				model, ok = updated.(Model)
				if !ok {
					t.Fatal("Update did not return Model type")
				}
			}

			if model.cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, model.cursor)
			}
		})
	}
}

func TestListViewActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      tea.Msg
		expected string
		notes    string
	}{
		{
			name:     "enter_selects",
			key:      key(tea.KeyEnter),
			expected: "dev",
			notes:    "Attaches to selected session",
		},
		{
			name:     "q_quits",
			key:      keyRune('q'),
			expected: "",
			notes:    "Clean exit",
		},
		{
			name:     "ctrl_c_quits",
			key:      key(tea.KeyCtrlC),
			expected: "",
			notes:    "Alternative quit",
		},
		{
			name:     "n_switches_to_input",
			key:      keyRune('n'),
			expected: "",
			notes:    "Switches to name entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessions := testSessions()
			model := newModel(sessions)
			model.cursor = 0

			var cmd tea.Cmd

			updated, cmd := model.Update(tt.key)

			var ok bool

			model, ok = updated.(Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			switch tt.name {
			case "enter_selects":
				if model.Result() != tt.expected {
					t.Errorf("expected result %q, got %q", tt.expected, model.Result())
				}

				if cmd == nil {
					t.Error("expected tea.Quit command")
				}
			case "q_quits", "ctrl_c_quits":
				if !model.Quitting() {
					t.Error("expected quitting to be true")
				}

				if cmd == nil {
					t.Error("expected tea.Quit command")
				}
			case "n_switches_to_input":
				if !model.inputMode {
					t.Error("expected inputMode to be true")
				}
			}
		})
	}
}

func TestInputView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		typedText  string
		key        tea.Msg
		expected   string
		notes      string
		checkInput bool
	}{
		{
			name:       "submit_name",
			typedText:  "my-session",
			key:        key(tea.KeyEnter),
			expected:   "my-session",
			notes:      "Confirms new session name",
			checkInput: false,
		},
		{
			name:       "esc_returns_to_list",
			typedText:  "partial",
			key:        key(tea.KeyEsc),
			expected:   "",
			notes:      "Cancels input",
			checkInput: true,
		},
		{
			name:       "ctrl_c_in_input",
			typedText:  "partial",
			key:        key(tea.KeyCtrlC),
			expected:   "",
			notes:      "Quits from input view",
			checkInput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessions := testSessions()
			model := newModel(sessions)

			// Switch to input mode
			updated, _ := model.Update(keyRune('n'))

			var ok bool

			model, ok = updated.(Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			if !model.inputMode {
				t.Fatal("expected to be in input mode")
			}

			// Type the text
			for _, r := range tt.typedText {
				u, _ := model.Update(keyRune(r))

				model, ok = u.(Model)
				if !ok {
					t.Fatal("Update did not return Model type")
				}
			}

			// Send the key
			var cmd tea.Cmd

			updated, cmd = model.Update(tt.key)

			model, ok = updated.(Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			switch tt.name {
			case "submit_name":
				if model.Result() != tt.expected {
					t.Errorf("expected result %q, got %q", tt.expected, model.Result())
				}

				if cmd == nil {
					t.Error("expected tea.Quit command")
				}
			case "esc_returns_to_list":
				if model.inputMode {
					t.Error("expected to be back in list mode")
				}

				if model.textInput.Value() != "" {
					t.Error("expected input to be cleared")
				}
			case "ctrl_c_in_input":
				if !model.Quitting() {
					t.Error("expected quitting to be true")
				}

				if cmd == nil {
					t.Error("expected tea.Quit command")
				}
			}
		})
	}
}

func TestNameValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		typedText        string
		notes            string
		existingSessions []gomux.Session
		expectError      bool
		expectWarning    bool
		canSubmit        bool
	}{
		{
			name:             "valid_name",
			typedText:        "new-session",
			existingSessions: []gomux.Session{{Name: "dev"}, {Name: "api"}},
			expectError:      false,
			expectWarning:    false,
			canSubmit:        true,
			notes:            "Happy path",
		},
		{
			name:             "name_with_dot",
			typedText:        "my.session",
			existingSessions: []gomux.Session{},
			expectError:      true,
			expectWarning:    false,
			canSubmit:        false,
			notes:            "Dot is invalid in tmux",
		},
		{
			name:             "name_with_colon",
			typedText:        "my:session",
			existingSessions: []gomux.Session{},
			expectError:      true,
			expectWarning:    false,
			canSubmit:        false,
			notes:            "Colon is invalid in tmux",
		},
		{
			name:             "duplicate_name",
			typedText:        "dev",
			existingSessions: []gomux.Session{{Name: "dev"}, {Name: "api"}},
			expectError:      false,
			expectWarning:    true,
			canSubmit:        true,
			notes:            "Will attach, not create",
		},
		{
			name:             "empty_name",
			typedText:        "",
			existingSessions: []gomux.Session{},
			expectError:      false,
			expectWarning:    false,
			canSubmit:        false,
			notes:            "Cannot submit empty name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newModel(tt.existingSessions)

			// Switch to input mode
			u, _ := model.Update(keyRune('n'))

			var ok bool

			model, ok = u.(Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			// Type the text
			for _, r := range tt.typedText {
				upd, _ := model.Update(keyRune(r))

				model, ok = upd.(Model)
				if !ok {
					t.Fatal("Update did not return Model type")
				}
			}

			// Check error state
			hasError := model.err != ""
			if hasError != tt.expectError {
				t.Errorf("expected error=%v, got error=%v (err=%q)", tt.expectError, hasError, model.err)
			}

			// Check warning state
			hasWarning := model.warning != ""
			if hasWarning != tt.expectWarning {
				t.Errorf("expected warning=%v, got warning=%v (warning=%q)", tt.expectWarning, hasWarning, model.warning)
			}

			// Try to submit
			var cmd tea.Cmd

			updated, cmd := model.Update(key(tea.KeyEnter))

			model, ok = updated.(Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			//nolint:nestif // Test validation logic naturally has nested conditions
			if tt.canSubmit {
				// Should have set result and returned quit command
				if model.Result() != tt.typedText {
					t.Errorf("expected result %q, got %q", tt.typedText, model.Result())
				}

				if cmd == nil {
					t.Error("expected tea.Quit command")
				}
			} else {
				// Should still be in input mode
				if !model.inputMode {
					t.Error("expected to still be in input mode (submission blocked)")
				}

				if model.Result() != "" {
					t.Error("expected no result (submission blocked)")
				}
			}
		})
	}
}
