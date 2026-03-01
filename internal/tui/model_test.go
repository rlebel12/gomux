package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rlebel12/gomux/internal/gomux"
)

// key creates a tea.KeyMsg for a key type.
func key(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// keyRune creates a tea.KeyMsg for a rune.
func keyRune(r rune) tea.KeyMsg {
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

// testProfiles returns 2 sample profiles for testing.
func testProfiles() []gomux.Profile {
	return []gomux.Profile{
		{
			Name: "profile0",
			Windows: []gomux.Window{
				{Name: "editor"},
				{Name: "terminal"},
			},
		},
		{
			Name: "profile1",
			Windows: []gomux.Window{
				{Name: "server"},
			},
		},
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
			model := newModel(sessions, nil)
			model.cursor = tt.initialCursor

			// Apply key sequence
			for _, msg := range tt.keySequence {
				updated, _ := model.Update(msg)

				var ok bool

				model, ok = updated.(*Model)
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
		expected gomux.Selection
		notes    string
	}{
		{
			name:     "enter_selects",
			key:      key(tea.KeyEnter),
			expected: gomux.Selection{Name: "dev", IsProfile: false},
			notes:    "Attaches to selected session",
		},
		{
			name:     "q_quits",
			key:      keyRune('q'),
			expected: gomux.Selection{},
			notes:    "Clean exit",
		},
		{
			name:     "ctrl_c_quits",
			key:      key(tea.KeyCtrlC),
			expected: gomux.Selection{},
			notes:    "Alternative quit",
		},
		{
			name:     "n_switches_to_input",
			key:      keyRune('n'),
			expected: gomux.Selection{},
			notes:    "Switches to name entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessions := testSessions()
			model := newModel(sessions, nil)
			model.cursor = 0

			var cmd tea.Cmd

			updated, cmd := model.Update(tt.key)

			var ok bool

			model, ok = updated.(*Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			switch tt.name {
			case "enter_selects":
				if model.Result() != tt.expected {
					t.Errorf("expected result %v, got %v", tt.expected, model.Result())
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
			model := newModel(sessions, nil)

			// Switch to input mode
			updated, _ := model.Update(keyRune('n'))

			var ok bool

			model, ok = updated.(*Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			if !model.inputMode {
				t.Fatal("expected to be in input mode")
			}

			// Type the text
			for _, r := range tt.typedText {
				u, _ := model.Update(keyRune(r))

				model, ok = u.(*Model)
				if !ok {
					t.Fatal("Update did not return Model type")
				}
			}

			// Send the key
			var cmd tea.Cmd

			updated, cmd = model.Update(tt.key)

			model, ok = updated.(*Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			switch tt.name {
			case "submit_name":
				if model.Result().Name != tt.expected {
					t.Errorf("expected result name %q, got %q", tt.expected, model.Result().Name)
				}

				if model.Result().IsProfile {
					t.Error("expected IsProfile to be false for typed name")
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

			model := newModel(tt.existingSessions, nil)

			// Switch to input mode
			u, _ := model.Update(keyRune('n'))

			var ok bool

			model, ok = u.(*Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			// Type the text
			for _, r := range tt.typedText {
				upd, _ := model.Update(keyRune(r))

				model, ok = upd.(*Model)
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

			model, ok = updated.(*Model)
			if !ok {
				t.Fatal("Update did not return Model type")
			}

			verifySubmitResult(t, model, cmd, tt.typedText, tt.canSubmit)
		})
	}
}

// verifySubmitResult checks whether a submit action produced the expected outcome.
func verifySubmitResult(t *testing.T, model *Model, cmd tea.Cmd, typedText string, canSubmit bool) {
	t.Helper()

	if canSubmit {
		if model.Result().Name != typedText {
			t.Errorf("expected result name %q, got %q", typedText, model.Result().Name)
		}

		if cmd == nil {
			t.Error("expected tea.Quit command")
		}

		return
	}

	if !model.inputMode {
		t.Error("expected to still be in input mode (submission blocked)")
	}

	if model.Result().Name != "" {
		t.Error("expected no result (submission blocked)")
	}
}

func TestListViewNavigationWithProfiles(t *testing.T) {
	t.Parallel()

	profiles := testProfiles()

	tests := []struct {
		name           string
		sessions       []gomux.Session
		profiles       []gomux.Profile
		initialCursor  int
		keySequence    []tea.Msg
		expectedCursor int
		notes          string
	}{
		{
			name:           "sessions_only_navigates_normally",
			sessions:       makeSessions(3),
			profiles:       nil,
			initialCursor:  0,
			keySequence:    []tea.Msg{keyRune('j')},
			expectedCursor: 1, // sessions[1]
			notes:          "No profile section when empty",
		},
		{
			name:           "profiles_navigate_down_into_profiles",
			sessions:       makeSessions(2),
			profiles:       profiles,
			initialCursor:  1, // sessions[1]
			keySequence:    []tea.Msg{keyRune('j'), keyRune('j')},
			expectedCursor: 3, // profiles[1]
			notes:          "Cursor enters profile section at index 2, moves to 3",
		},
		{
			name:           "profiles_navigate_up_from_profiles",
			sessions:       makeSessions(2),
			profiles:       profiles,
			initialCursor:  1, // sessions[1]
			keySequence:    []tea.Msg{keyRune('j'), keyRune('j'), keyRune('k')},
			expectedCursor: 2, // profiles[0]
			notes:          "Cursor moves back within profiles",
		},
		{
			name:           "profiles_navigate_up_to_sessions",
			sessions:       makeSessions(2),
			profiles:       profiles,
			initialCursor:  1, // sessions[1]
			keySequence:    []tea.Msg{keyRune('j'), keyRune('j'), keyRune('k'), keyRune('k')},
			expectedCursor: 1, // sessions[1]
			notes:          "Cursor crosses back to sessions",
		},
		{
			name:           "profiles_clamp_at_bottom",
			sessions:       makeSessions(1),
			profiles:       profiles[:1],
			initialCursor:  0,
			keySequence:    []tea.Msg{keyRune('j'), keyRune('j'), keyRune('j')},
			expectedCursor: 1, // profiles[0]; max cursor = len(sessions)+len(profiles)-1 = 1
			notes:          "max cursor = len(sessions)+len(profiles)-1 = 1",
		},
		{
			name:           "only_profiles_no_sessions",
			sessions:       nil,
			profiles:       profiles,
			initialCursor:  0,
			keySequence:    []tea.Msg{keyRune('j')},
			expectedCursor: 1, // profiles[1]
			notes:          "Navigate within profiles-only list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newModel(tt.sessions, tt.profiles)
			model.cursor = tt.initialCursor

			for _, msg := range tt.keySequence {
				updated, _ := model.Update(msg)

				var ok bool

				model, ok = updated.(*Model)
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

func TestListViewSelectionWithProfiles(t *testing.T) {
	t.Parallel()

	profiles := testProfiles()

	tests := []struct {
		name     string
		sessions []gomux.Session
		profiles []gomux.Profile
		keys     []tea.Msg
		expected gomux.Selection
		notes    string
	}{
		{
			name:     "select_session",
			sessions: makeSessions(2),
			profiles: profiles,
			keys:     []tea.Msg{key(tea.KeyEnter)},
			expected: gomux.Selection{Name: "session0", IsProfile: false},
			notes:    "Select first session",
		},
		{
			name:     "select_profile",
			sessions: makeSessions(2),
			profiles: profiles,
			keys:     []tea.Msg{keyRune('j'), keyRune('j'), key(tea.KeyEnter)},
			expected: gomux.Selection{Name: "profile0", IsProfile: true},
			notes:    "Navigate to first profile (cursor=2), select",
		},
		{
			name:     "select_only_profile",
			sessions: nil,
			profiles: profiles[:1],
			keys:     []tea.Msg{key(tea.KeyEnter)},
			expected: gomux.Selection{Name: "profile0", IsProfile: true},
			notes:    "No sessions, select profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newModel(tt.sessions, tt.profiles)

			var (
				cmd tea.Cmd
				ok  bool
			)

			for _, msg := range tt.keys {
				updated, c := model.Update(msg)
				cmd = c

				model, ok = updated.(*Model)
				if !ok {
					t.Fatal("Update did not return Model type")
				}
			}

			if model.Result() != tt.expected {
				t.Errorf("expected selection %v, got %v", tt.expected, model.Result())
			}

			if cmd == nil {
				t.Error("expected tea.Quit command after Enter")
			}
		})
	}
}

func TestListViewRenderWithProfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sessions         []gomux.Session
		profiles         []gomux.Profile
		expectedContains string
		shouldNotContain string
		notes            string
	}{
		{
			name:             "shows_profile_section",
			sessions:         makeSessions(1),
			profiles:         testProfiles()[:1],
			expectedContains: "Profiles",
			notes:            "Section header rendered",
		},
		{
			name:             "no_profile_section_when_empty",
			sessions:         makeSessions(1),
			profiles:         nil,
			shouldNotContain: "Profiles",
			notes:            "Section hidden",
		},
		{
			name:     "profile_shows_window_count",
			sessions: nil,
			profiles: []gomux.Profile{
				{
					Name: "profile0",
					Windows: []gomux.Window{
						{Name: "editor"},
						{Name: "terminal"},
					},
				},
			},
			expectedContains: "2 windows",
			notes:            "Metadata from profile definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newModel(tt.sessions, tt.profiles)
			view := model.View()

			if tt.expectedContains != "" {
				if !containsString(view, tt.expectedContains) {
					t.Errorf("expected view to contain %q, got:\n%s", tt.expectedContains, view)
				}
			}

			if tt.shouldNotContain != "" {
				if containsString(view, tt.shouldNotContain) {
					t.Errorf("expected view NOT to contain %q, got:\n%s", tt.shouldNotContain, view)
				}
			}
		})
	}
}

func TestInputModeReturnsSelection(t *testing.T) {
	t.Parallel()

	sessions := makeSessions(1)
	profiles := testProfiles()
	model := newModel(sessions, profiles)

	// Switch to input mode
	updated, _ := model.Update(keyRune('n'))

	var ok bool

	model, ok = updated.(*Model)
	if !ok {
		t.Fatal("Update did not return Model type")
	}

	// Type a session name
	inputName := "my-new-session"
	for _, r := range inputName {
		u, _ := model.Update(keyRune(r))

		model, ok = u.(*Model)
		if !ok {
			t.Fatal("Update did not return Model type")
		}
	}

	// Submit
	updated, cmd := model.Update(key(tea.KeyEnter))

	model, ok = updated.(*Model)
	if !ok {
		t.Fatal("Update did not return Model type")
	}

	result := model.Result()
	if result.Name != inputName {
		t.Errorf("expected result name %q, got %q", inputName, result.Name)
	}

	if result.IsProfile {
		t.Error("expected IsProfile to be false for typed session name")
	}

	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

// makeSessions creates N sessions with predictable names (session0, session1, ...).
func makeSessions(n int) []gomux.Session {
	sessions := make([]gomux.Session, n)
	for i := range sessions {
		sessions[i] = gomux.Session{
			Name:    fmt.Sprintf("session%d", i),
			Created: time.Now(),
			Windows: 1,
		}
	}

	return sessions
}

// containsString returns true if s contains substr.
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
