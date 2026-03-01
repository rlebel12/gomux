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

	return &Client{tmuxPath: tmuxPath}, nil
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
		if strings.Contains(stderrStr, "no server running") || strings.Contains(stderrStr, "no sessions") || strings.Contains(stderrStr, "error connecting to") {
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
