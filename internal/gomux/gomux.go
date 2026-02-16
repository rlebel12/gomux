// Package gomux provides domain types and interfaces for tmux session management.
package gomux

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Session represents a tmux session.
type Session struct {
	Created  time.Time
	Name     string
	Windows  int
	Attached bool
}

// DefaultSessionName is the session name used when no sessions exist and the TUI is skipped.
const DefaultSessionName = "default"

// ErrNoSelection is returned by SessionPicker when the user quits without selecting a session.
var ErrNoSelection = errors.New("no session selected")

// SessionLister lists available tmux sessions.
type SessionLister interface {
	ListSessions(ctx context.Context) ([]Session, error)
}

// SessionPicker presents sessions to the user and returns the selected session name.
type SessionPicker interface {
	Pick(ctx context.Context, sessions []Session) (string, error)
}

// SessionExecutor attaches to or creates a tmux session.
type SessionExecutor interface {
	ExecIntoSession(ctx context.Context, sessionName string) error
}

// App orchestrates session selection and attachment.
type App struct {
	lister   SessionLister
	picker   SessionPicker
	executor SessionExecutor
}

// New creates a new App with the provided dependencies.
func New(lister SessionLister, picker SessionPicker, executor SessionExecutor) *App {
	return &App{
		lister:   lister,
		picker:   picker,
		executor: executor,
	}
}

// Run executes the session selection and attachment workflow.
//
// Mode routing:
//   - sessionName non-empty → bypass picker, exec directly into named session
//   - sessionName empty, zero sessions → exec into DefaultSessionName (skip TUI)
//   - sessionName empty, 1+ sessions → list sessions → present picker → exec into selection
//
// Returns nil if the user quits without selecting (ErrNoSelection is handled internally).
func (a *App) Run(ctx context.Context, sessionName string) error {
	// Direct mode: bypass lister and picker, exec directly
	if sessionName != "" {
		return a.execSession(ctx, sessionName)
	}

	// List existing sessions
	sessions, err := a.lister.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	// Empty state: exec into DefaultSessionName (skip picker)
	if len(sessions) == 0 {
		return a.execSession(ctx, DefaultSessionName)
	}

	// TUI mode: present picker and exec chosen session
	chosen, err := a.picker.Pick(ctx, sessions)
	if err != nil {
		// User quit: clean exit
		if errors.Is(err, ErrNoSelection) {
			return nil
		}

		return fmt.Errorf("pick session: %w", err)
	}

	return a.execSession(ctx, chosen)
}

// execSession wraps executor calls with consistent error handling.
func (a *App) execSession(ctx context.Context, sessionName string) error {
	if err := a.executor.ExecIntoSession(ctx, sessionName); err != nil {
		return fmt.Errorf("exec session: %w", err)
	}

	return nil
}
