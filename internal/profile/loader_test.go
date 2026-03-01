package profile_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rlebel12/gomux/internal/gomux"
	"github.com/rlebel12/gomux/internal/profile"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "profiles-*.yml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

func TestLoaderLoadProfiles(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home dir: %v", err)
	}

	tests := []struct {
		name     string
		yaml     string
		noFile   bool
		expected []gomux.Profile
	}{
		{
			name: "single_profile_minimal",
			yaml: "profiles:\n  - name: dev\n    windows:\n      - name: main\n",
			expected: []gomux.Profile{
				{
					Name: "dev",
					Windows: []gomux.Window{
						{Name: "main"},
					},
				},
			},
		},
		{
			name: "single_profile_full",
			yaml: `profiles:
  - name: work
    windows:
      - name: editor
        dir: /home/user/project
        command: nvim
        panes:
          - dir: /home/user/project
            command: git log
            split: horizontal
          - dir: /tmp
            command: htop
            split: vertical
`,
			expected: []gomux.Profile{
				{
					Name: "work",
					Windows: []gomux.Window{
						{
							Name:    "editor",
							Dir:     "/home/user/project",
							Command: "nvim",
							Panes: []gomux.Pane{
								{Dir: "/home/user/project", Command: "git log", Split: "horizontal"},
								{Dir: "/tmp", Command: "htop", Split: "vertical"},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple_profiles",
			yaml: `profiles:
  - name: dev
    windows:
      - name: main
  - name: ops
    windows:
      - name: logs
`,
			expected: []gomux.Profile{
				{Name: "dev", Windows: []gomux.Window{{Name: "main"}}},
				{Name: "ops", Windows: []gomux.Window{{Name: "logs"}}},
			},
		},
		{
			name: "pane_splits",
			yaml: `profiles:
  - name: dev
    windows:
      - name: main
        panes:
          - split: horizontal
          - split: vertical
`,
			expected: []gomux.Profile{
				{
					Name: "dev",
					Windows: []gomux.Window{
						{
							Name: "main",
							Panes: []gomux.Pane{
								{Split: "horizontal"},
								{Split: "vertical"},
							},
						},
					},
				},
			},
		},
		{
			name:     "empty_profiles_list",
			yaml:     "profiles: []\n",
			expected: []gomux.Profile{},
		},
		{
			name:     "file_not_found",
			noFile:   true,
			expected: []gomux.Profile{},
		},
		{
			name: "tilde_in_dir",
			yaml: `profiles:
  - name: dev
    windows:
      - name: main
        dir: ~/code/project
        panes:
          - dir: ~/code/project
`,
			expected: []gomux.Profile{
				{
					Name: "dev",
					Windows: []gomux.Window{
						{
							Name: "main",
							Dir:  filepath.Join(home, "code", "project"),
							Panes: []gomux.Pane{
								{Dir: filepath.Join(home, "code", "project")},
							},
						},
					},
				},
			},
		},
		{
			name: "no_tilde",
			yaml: `profiles:
  - name: dev
    windows:
      - name: main
        dir: /absolute/path
`,
			expected: []gomux.Profile{
				{
					Name: "dev",
					Windows: []gomux.Window{
						{Name: "main", Dir: "/absolute/path"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var path string
			if tt.noFile {
				path = filepath.Join(t.TempDir(), "nonexistent.yml")
			} else {
				path = writeTempYAML(t, tt.yaml)
			}

			loader := profile.NewLoader(path)
			got, err := loader.LoadProfiles(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("profile count: got %d, want %d", len(got), len(tt.expected))
			}

			for i, want := range tt.expected {
				g := got[i]
				if g.Name != want.Name {
					t.Errorf("profile[%d].Name: got %q, want %q", i, g.Name, want.Name)
				}
				if len(g.Windows) != len(want.Windows) {
					t.Fatalf("profile[%d].Windows count: got %d, want %d", i, len(g.Windows), len(want.Windows))
				}
				for j, ww := range want.Windows {
					gw := g.Windows[j]
					if gw.Name != ww.Name {
						t.Errorf("profile[%d].Windows[%d].Name: got %q, want %q", i, j, gw.Name, ww.Name)
					}
					if gw.Dir != ww.Dir {
						t.Errorf("profile[%d].Windows[%d].Dir: got %q, want %q", i, j, gw.Dir, ww.Dir)
					}
					if gw.Command != ww.Command {
						t.Errorf("profile[%d].Windows[%d].Command: got %q, want %q", i, j, gw.Command, ww.Command)
					}
					if len(gw.Panes) != len(ww.Panes) {
						t.Fatalf("profile[%d].Windows[%d].Panes count: got %d, want %d", i, j, len(gw.Panes), len(ww.Panes))
					}
					for k, wp := range ww.Panes {
						gp := gw.Panes[k]
						if gp.Dir != wp.Dir {
							t.Errorf("profile[%d].Windows[%d].Panes[%d].Dir: got %q, want %q", i, j, k, gp.Dir, wp.Dir)
						}
						if gp.Command != wp.Command {
							t.Errorf("profile[%d].Windows[%d].Panes[%d].Command: got %q, want %q", i, j, k, gp.Command, wp.Command)
						}
						if gp.Split != wp.Split {
							t.Errorf("profile[%d].Windows[%d].Panes[%d].Split: got %q, want %q", i, j, k, gp.Split, wp.Split)
						}
					}
				}
			}
		})
	}
}

func TestLoaderLoadProfilesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		errContains string
	}{
		{
			name:        "invalid_yaml",
			yaml:        "profiles:\n  - name: [broken\n",
			errContains: "parse profiles",
		},
		{
			name:        "missing_name",
			yaml:        "profiles:\n  - windows:\n      - name: main\n",
			errContains: "profile 0: missing name",
		},
		{
			name:        "missing_windows",
			yaml:        "profiles:\n  - name: dev\n",
			errContains: `profile "dev": missing windows`,
		},
		{
			name:        "empty_window_list",
			yaml:        "profiles:\n  - name: dev\n    windows: []\n",
			errContains: `profile "dev": missing windows`,
		},
		{
			name:        "window_missing_name",
			yaml:        "profiles:\n  - name: dev\n    windows:\n      - dir: /tmp\n",
			errContains: `profile "dev", window 0: missing name`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := writeTempYAML(t, tt.yaml)
			loader := profile.NewLoader(path)
			_, err := loader.LoadProfiles(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := err.Error(); !strings.Contains(got, tt.errContains) {
				t.Errorf("error %q does not contain %q", got, tt.errContains)
			}
		})
	}
}
