// Package gomux provides domain types and interfaces for tmux session management.
package gomux

import (
	"context"
	"errors"
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
// The full orchestration logic will be implemented in Phase 02.
func (a *App) Run(_ context.Context, _ string) error {
	// Stub implementation - Phase 02 adds the real logic
	return nil
}
