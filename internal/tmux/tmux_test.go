package tmux

import (
	"testing"
	"time"

	"github.com/rlebel12/gomux/internal/gomux"
)

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
