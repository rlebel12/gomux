// Package tmux provides integration with the tmux terminal multiplexer.
package tmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rlebel12/gomux/internal/gomux"
)

// commandRunner executes a tmux subcommand and returns any error.
type commandRunner func(ctx context.Context, tmuxPath string, args ...string) error

var (
	// ErrInvalidFieldCount is returned when a line doesn't have the expected number of fields.
	ErrInvalidFieldCount = errors.New("invalid field count")
	// ErrInvalidAttachedFlag is returned when the attached flag is not 0 or 1.
	ErrInvalidAttachedFlag = errors.New("invalid attached flag")
)

// formatString is the tmux list-sessions format string for parsing session data.
const formatString = "#{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}"

// parseSessions parses tmux list-sessions -F output into Session structs.
//
// Expected format per line (tab-delimited):
//
//	#{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}
//
// Returns an empty slice (not nil) for empty input.
func parseSessions(output string) ([]gomux.Session, error) {
	const (
		fieldName      = 0
		fieldWindows   = 1
		fieldAttached  = 2
		fieldCreated   = 3
		expectedFields = 4
	)

	lines := strings.Split(output, "\n")
	sessions := []gomux.Session{}

	lineNum := 0
	for _, line := range lines {
		lineNum++

		// Skip empty lines (handles trailing newline)
		if line == "" {
			continue
		}

		// Split by tab - expect exactly 4 fields
		fields := strings.Split(line, "\t")
		if len(fields) != expectedFields {
			return nil, fmt.Errorf(
				"line %d: %w: expected %d, got %d",
				lineNum, ErrInvalidFieldCount, expectedFields, len(fields),
			)
		}

		// Parse session name (use as-is)
		name := fields[fieldName]

		// Parse window count
		windows, err := strconv.Atoi(fields[fieldWindows])
		if err != nil {
			return nil, fmt.Errorf("line %d: parse window count: %w", lineNum, err)
		}

		// Parse attached flag (0 or 1)
		var attached bool

		switch fields[fieldAttached] {
		case "0":
			attached = false
		case "1":
			attached = true
		default:
			return nil, fmt.Errorf(
				"line %d: %w: expected 0 or 1, got %q",
				lineNum, ErrInvalidAttachedFlag, fields[fieldAttached],
			)
		}

		// Parse created timestamp
		timestamp, err := strconv.ParseInt(fields[fieldCreated], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: parse created timestamp: %w", lineNum, err)
		}

		created := time.Unix(timestamp, 0)

		sessions = append(sessions, gomux.Session{
			Name:     name,
			Windows:  windows,
			Attached: attached,
			Created:  created,
		})
	}

	return sessions, nil
}

// Client provides integration with the tmux command-line tool.
type Client struct {
	runCmd   commandRunner
	tmuxPath string
}

// NewClient creates a new Client by resolving the tmux binary path.
//
// The tmux path is resolved eagerly at construction time, allowing for fail-fast
// behavior if tmux is not installed. Returns an error if tmux is not found on PATH.
func NewClient() (*Client, error) {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return nil, fmt.Errorf("tmux not found on PATH: %w", err)
	}

	return &Client{
		tmuxPath: tmuxPath,
		runCmd:   defaultCommandRunner,
	}, nil
}

// defaultCommandRunner executes a tmux subcommand using exec.CommandContext.
func defaultCommandRunner(ctx context.Context, tmuxPath string, args ...string) error {
	//nolint:gosec // tmuxPath is validated via exec.LookPath in NewClient
	cmd := exec.CommandContext(ctx, tmuxPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, bytes.TrimSpace(output))
	}

	return nil
}

// ListSessions retrieves all tmux sessions using the list-sessions command.
//
// Returns an empty slice (not an error) if tmux is not running or has no sessions.
// This distinguishes expected "no sessions" state from actual command failures.
func (c *Client) ListSessions(ctx context.Context) ([]gomux.Session, error) {
	//nolint:gosec // tmuxPath is validated via exec.LookPath in NewClient
	cmd := exec.CommandContext(ctx, c.tmuxPath, "list-sessions", "-F", formatString)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// tmux exits with non-zero when no server is running or no sessions exist
		// This is expected - return empty slice, not an error
		stderrStr := stderr.String()
		noSessionsRunning := strings.Contains(stderrStr, "no server running") ||
			strings.Contains(stderrStr, "no sessions") ||
			strings.Contains(stderrStr, "error connecting to")
		if noSessionsRunning {
			return []gomux.Session{}, nil
		}

		return nil, fmt.Errorf("list sessions: %w", err)
	}

	sessions, err := parseSessions(stdout.String())
	if err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}

	return sessions, nil
}

// ExecIntoSession replaces the current process with tmux attach-or-create for the given session.
//
// Uses syscall.Exec to replace the current process with tmux. On success, this function
// never returns (the process is replaced). The error return path only executes on failure.
//
// The -A flag causes tmux to attach if the session exists, or create it if it doesn't.
func (c *Client) ExecIntoSession(_ context.Context, sessionName string) error {
	args := []string{"tmux", "new-session", "-A", "-s", sessionName}
	environ := os.Environ()

	//nolint:gosec // tmuxPath is validated via exec.LookPath in NewClient
	err := syscall.Exec(c.tmuxPath, args, environ)
	if err != nil {
		return fmt.Errorf("exec into session %q: %w", sessionName, err)
	}

	// Never reached on success (process replaced)
	return nil
}

// CreateSessionFromProfile creates a tmux session configured according to the given profile.
//
// The session is built by issuing a sequence of tmux subcommands:
//  1. new-session for the first window
//  2. new-window for each subsequent window
//  3. split-window for each pane within a window
//  4. send-keys for any startup command on a window or pane
//  5. select-window to focus the first window at the end
//
// If any command fails, a best-effort kill-session is attempted to clean up the
// partial session before returning the error.
func (c *Client) CreateSessionFromProfile(ctx context.Context, profile gomux.Profile) error {
	cmds := buildSessionCommands(profile)

	for _, args := range cmds {
		if err := c.runCmd(ctx, c.tmuxPath, args...); err != nil {
			// Best-effort cleanup: include kill error as secondary context if it occurs.
			killErr := c.runCmd(ctx, c.tmuxPath, "kill-session", "-t", profile.Name)
			return fmt.Errorf("create session %q: %w", profile.Name, errors.Join(err, killErr))
		}
	}

	return nil
}

const (
	splitHorizontal = "-h"
	splitVertical   = "-v"
)

// buildSessionCommands constructs the ordered sequence of tmux argument lists for creating a session.
func buildSessionCommands(profile gomux.Profile) [][]string {
	var cmds [][]string

	for i, window := range profile.Windows {
		winTarget := fmt.Sprintf("%s:%d", profile.Name, i)

		if i == 0 {
			// First window: new-session creates the session and first window together
			cmds = append(cmds, appendWindowFlags(
				[]string{"new-session", "-d", "-s", profile.Name},
				window,
			))
		} else {
			// Subsequent windows: new-window adds a window to an existing session
			cmds = append(cmds, appendWindowFlags(
				[]string{"new-window", "-t", profile.Name},
				window,
			))
		}

		// Window startup command
		if window.Command != "" {
			cmds = append(cmds, []string{"send-keys", "-t", winTarget, window.Command, "Enter"})
		}

		// Panes within this window
		for _, pane := range window.Panes {
			splitArgs := []string{"split-window", splitFlagFor(pane.Split), "-t", winTarget}
			if pane.Dir != "" {
				splitArgs = append(splitArgs, "-c", pane.Dir)
			}
			cmds = append(cmds, splitArgs)

			if pane.Command != "" {
				cmds = append(cmds, []string{"send-keys", "-t", winTarget, pane.Command, "Enter"})
			}
		}
	}

	// Focus first window
	cmds = append(cmds, []string{"select-window", "-t", profile.Name + ":0"})

	return cmds
}

// appendWindowFlags appends the optional -n (name) and -c (dir) flags for a window to args.
func appendWindowFlags(args []string, window gomux.Window) []string {
	if window.Name != "" {
		args = append(args, "-n", window.Name)
	}

	if window.Dir != "" {
		args = append(args, "-c", window.Dir)
	}

	return args
}

// splitFlagFor returns the tmux split-window flag for the given split direction.
// Empty or unrecognized values default to vertical.
func splitFlagFor(split string) string {
	if split == "horizontal" {
		return splitHorizontal
	}

	return splitVertical
}
