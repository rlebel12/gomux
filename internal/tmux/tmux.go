// Package tmux provides integration with the tmux terminal multiplexer.
package tmux

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rlebel12/gomux/internal/gomux"
)

var (
	// ErrInvalidFieldCount is returned when a line doesn't have the expected number of fields.
	ErrInvalidFieldCount = errors.New("invalid field count")
	// ErrInvalidAttachedFlag is returned when the attached flag is not 0 or 1.
	ErrInvalidAttachedFlag = errors.New("invalid attached flag")
)

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
