package tmux

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rlebel12/gomux/internal/gomux"
)

// errSimulatedRunnerFailure is returned by recordingRunner when a call matches errOnCall.
var errSimulatedRunnerFailure = errors.New("simulated runner failure")

// recordingRunner captures each call's args for assertion.
// errOnCall specifies the 0-indexed call number on which to return an error (-1 for never).
type recordingRunner struct {
	calls     [][]string
	errOnCall int
}

func (r *recordingRunner) run(_ context.Context, _ string, args ...string) error {
	callIdx := len(r.calls)
	r.calls = append(r.calls, args)

	if r.errOnCall >= 0 && callIdx == r.errOnCall {
		return fmt.Errorf("call %d: %w", callIdx, errSimulatedRunnerFailure)
	}

	return nil
}

func TestParseSessions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []gomux.Session
	}{
		{
			name:  "single_detached",
			input: "dev\t3\t0\t1700000000\n",
			expected: []gomux.Session{
				{
					Name:     "dev",
					Windows:  3,
					Attached: false,
					Created:  time.Unix(1700000000, 0),
				},
			},
		},
		{
			name:  "single_attached",
			input: "work\t1\t1\t1700001000\n",
			expected: []gomux.Session{
				{
					Name:     "work",
					Windows:  1,
					Attached: true,
					Created:  time.Unix(1700001000, 0),
				},
			},
		},
		{
			name:  "multiple_sessions",
			input: "dev\t3\t0\t1700000000\napi\t1\t1\t1700001000\nscratch\t2\t0\t1700002000\n",
			expected: []gomux.Session{
				{
					Name:     "dev",
					Windows:  3,
					Attached: false,
					Created:  time.Unix(1700000000, 0),
				},
				{
					Name:     "api",
					Windows:  1,
					Attached: true,
					Created:  time.Unix(1700001000, 0),
				},
				{
					Name:     "scratch",
					Windows:  2,
					Attached: false,
					Created:  time.Unix(1700002000, 0),
				},
			},
		},
		{
			name:     "empty_output",
			input:    "",
			expected: []gomux.Session{},
		},
		{
			name:     "whitespace_only",
			input:    "\n",
			expected: []gomux.Session{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessions, err := parseSessions(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(sessions) != len(tt.expected) {
				t.Fatalf("expected %d sessions, got %d", len(tt.expected), len(sessions))
			}

			for i, session := range sessions {
				expected := tt.expected[i]

				if session.Name != expected.Name {
					t.Errorf("session[%d].Name: expected %q, got %q", i, expected.Name, session.Name)
				}

				if session.Windows != expected.Windows {
					t.Errorf("session[%d].Windows: expected %d, got %d", i, expected.Windows, session.Windows)
				}

				if session.Attached != expected.Attached {
					t.Errorf("session[%d].Attached: expected %v, got %v", i, expected.Attached, session.Attached)
				}

				if !session.Created.Equal(expected.Created) {
					t.Errorf("session[%d].Created: expected %v, got %v", i, expected.Created, session.Created)
				}
			}
		})
	}
}

func TestParseSessionsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "missing_fields",
			input:       "dev\t3\n",
			errContains: "invalid field count",
		},
		{
			name:        "non_numeric_windows",
			input:       "dev\tabc\t0\t1700000000\n",
			errContains: "parse window count",
		},
		{
			name:        "non_numeric_timestamp",
			input:       "dev\t3\t0\tnot-a-number\n",
			errContains: "parse created timestamp",
		},
		{
			name:        "invalid_attached",
			input:       "dev\t3\t2\t1700000000\n",
			errContains: "invalid attached flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseSessions(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestCreateSessionFromProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		profile      gomux.Profile
		expectedCmds [][]string
	}{
		{
			name: "single_window_no_command",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{Name: "main", Dir: "/home/rob"},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "main", "-c", "/home/rob"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "single_window_with_command",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{Name: "editor", Dir: "/home/rob", Command: "nvim"},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "editor", "-c", "/home/rob"},
				{"send-keys", "-t", "dev:0", "nvim", "Enter"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "multiple_windows",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{Name: "editor", Dir: "/home/rob"},
					{Name: "server", Dir: "/srv"},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "editor", "-c", "/home/rob"},
				{"new-window", "-t", "dev", "-n", "server", "-c", "/srv"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "window_with_panes",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{
						Name: "main",
						Dir:  "/home/rob",
						Panes: []gomux.Pane{
							{Dir: "/home/rob/a", Split: "horizontal"},
							{Dir: "/home/rob/b", Split: "vertical"},
						},
					},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "main", "-c", "/home/rob"},
				{"split-window", "-h", "-t", "dev:0", "-c", "/home/rob/a"},
				{"split-window", "-v", "-t", "dev:0", "-c", "/home/rob/b"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "horizontal_split",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{
						Name: "main",
						Dir:  "/home/rob",
						Panes: []gomux.Pane{
							{Dir: "/dir", Split: "horizontal"},
						},
					},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "main", "-c", "/home/rob"},
				{"split-window", "-h", "-t", "dev:0", "-c", "/dir"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "vertical_split",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{
						Name: "main",
						Dir:  "/home/rob",
						Panes: []gomux.Pane{
							{Dir: "/dir", Split: "vertical"},
						},
					},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "main", "-c", "/home/rob"},
				{"split-window", "-v", "-t", "dev:0", "-c", "/dir"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "default_split_is_vertical",
			profile: gomux.Profile{
				Name: "dev",
				Windows: []gomux.Window{
					{
						Name: "main",
						Dir:  "/home/rob",
						Panes: []gomux.Pane{
							{Dir: "/dir", Split: ""},
						},
					},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "dev", "-n", "main", "-c", "/home/rob"},
				{"split-window", "-v", "-t", "dev:0", "-c", "/dir"},
				{"select-window", "-t", "dev:0"},
			},
		},
		{
			name: "full_layout",
			profile: gomux.Profile{
				Name: "myapp",
				Windows: []gomux.Window{
					{
						Name:    "editor",
						Dir:     "/home/rob/code",
						Command: "nvim",
						Panes: []gomux.Pane{
							{Dir: "/home/rob/code/api", Split: "horizontal", Command: "go run ."},
						},
					},
					{
						Name:    "logs",
						Dir:     "/var/log",
						Command: "tail -f app.log",
					},
				},
			},
			expectedCmds: [][]string{
				{"new-session", "-d", "-s", "myapp", "-n", "editor", "-c", "/home/rob/code"},
				{"send-keys", "-t", "myapp:0", "nvim", "Enter"},
				{"split-window", "-h", "-t", "myapp:0", "-c", "/home/rob/code/api"},
				{"send-keys", "-t", "myapp:0", "go run .", "Enter"},
				{"new-window", "-t", "myapp", "-n", "logs", "-c", "/var/log"},
				{"send-keys", "-t", "myapp:1", "tail -f app.log", "Enter"},
				{"select-window", "-t", "myapp:0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &recordingRunner{errOnCall: -1}
			client := &Client{runCmd: runner.run}

			err := client.CreateSessionFromProfile(t.Context(), tt.profile)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(runner.calls) != len(tt.expectedCmds) {
				t.Fatalf("expected %d commands, got %d: %v", len(tt.expectedCmds), len(runner.calls), runner.calls)
			}

			for i, expectedArgs := range tt.expectedCmds {
				gotArgs := runner.calls[i]
				if len(gotArgs) != len(expectedArgs) {
					t.Errorf("command[%d]: expected args %v, got %v", i, expectedArgs, gotArgs)
					continue
				}

				for j, arg := range expectedArgs {
					if gotArgs[j] != arg {
						t.Errorf("command[%d][%d]: expected %q, got %q (full: expected=%v got=%v)",
							i, j, arg, gotArgs[j], expectedArgs, gotArgs)
					}
				}
			}
		})
	}
}

func TestCreateSessionFromProfileKillsOnError(t *testing.T) {
	t.Parallel()

	// Profile with enough commands to have a 3rd call (index 2)
	profile := gomux.Profile{
		Name: "dev",
		Windows: []gomux.Window{
			{Name: "editor", Dir: "/home/rob", Command: "nvim"},
			{Name: "server", Dir: "/srv"},
		},
	}
	// Commands: new-session(0), send-keys(1), new-window(2) <- fails, then kill-session attempted
	runner := &recordingRunner{errOnCall: 2}
	client := &Client{runCmd: runner.run}

	err := client.CreateSessionFromProfile(t.Context(), profile)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !contains(err.Error(), "dev") {
		t.Errorf("expected error to contain profile name %q, got %q", "dev", err.Error())
	}

	// Verify kill-session was attempted after failure
	foundKill := false
	for _, call := range runner.calls {
		if len(call) >= 3 && call[0] == "kill-session" && call[1] == "-t" && call[2] == "dev" {
			foundKill = true
			break
		}
	}

	if !foundKill {
		t.Errorf("expected kill-session -t dev to be called for cleanup, got calls: %v", runner.calls)
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
