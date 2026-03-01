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

// Profile represents a named tmux session layout that can be created on demand.
type Profile struct {
	Name    string
	Windows []Window
}

// Window represents a tmux window within a profile.
type Window struct {
	Name    string
	Dir     string
	Command string
	Panes   []Pane
}

// Pane represents a tmux pane within a window.
type Pane struct {
	Dir     string
	Command string
	Split   string // "horizontal" or "vertical"
}

// Selection represents the result of a picker interaction.
type Selection struct {
	Name      string
	IsProfile bool
}

// DefaultSessionName is the session name used when no sessions exist and the TUI is skipped.
const DefaultSessionName = "default"

// ErrNoSelection is returned by SessionPicker when the user quits without selecting a session.
var ErrNoSelection = errors.New("no session selected")

// ErrProfileNotFound is returned when a selected profile name cannot be found in the loaded profiles.
var ErrProfileNotFound = errors.New("profile not found")

// SessionLister lists available tmux sessions.
type SessionLister interface {
	ListSessions(ctx context.Context) ([]Session, error)
}

// SessionPicker presents sessions and profiles to the user and returns the selection.
type SessionPicker interface {
	Pick(ctx context.Context, sessions []Session, profiles []Profile) (Selection, error)
}

// SessionExecutor attaches to or creates a tmux session.
type SessionExecutor interface {
	ExecIntoSession(ctx context.Context, sessionName string) error
}

// ProfileLoader loads profile definitions from persistent storage.
type ProfileLoader interface {
	LoadProfiles(ctx context.Context) ([]Profile, error)
}

// SessionCreator creates a new tmux session from a profile definition.
type SessionCreator interface {
	CreateSessionFromProfile(ctx context.Context, profile Profile) error
}

// App orchestrates session selection and attachment.
type App struct {
	lister   SessionLister
	picker   SessionPicker
	executor SessionExecutor
	profiles ProfileLoader
	creator  SessionCreator
}

// New creates a new App with the provided dependencies.
func New(
	lister SessionLister,
	picker SessionPicker,
	executor SessionExecutor,
	profiles ProfileLoader,
	creator SessionCreator,
) *App {
	return &App{
		lister:   lister,
		picker:   picker,
		executor: executor,
		profiles: profiles,
		creator:  creator,
	}
}

// Run executes the session selection and attachment workflow.
//
// Mode routing:
//   - sessionName non-empty → bypass picker, exec directly into named session
//   - sessionName empty, no sessions, no profiles → exec into DefaultSessionName (skip TUI)
//   - sessionName empty, sessions or profiles present → present picker, exec into selection
//
// If a profile is selected, the session is created before attaching.
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

	// Load profiles
	allProfiles, err := a.profiles.LoadProfiles(ctx)
	if err != nil {
		return fmt.Errorf("load profiles: %w", err)
	}

	// Empty state: exec into DefaultSessionName (skip picker)
	if len(sessions) == 0 && len(allProfiles) == 0 {
		return a.execSession(ctx, DefaultSessionName)
	}

	// Filter profiles: exclude any whose Name matches an existing session
	filteredProfiles := filterProfiles(allProfiles, sessions)

	// TUI mode: present picker and exec chosen session or create profile session
	selection, err := a.picker.Pick(ctx, sessions, filteredProfiles)
	if err != nil {
		// User quit: clean exit
		if errors.Is(err, ErrNoSelection) {
			return nil
		}

		return fmt.Errorf("pick session: %w", err)
	}

	if selection.IsProfile {
		profile, found := findProfile(allProfiles, selection.Name)
		if !found {
			return fmt.Errorf("pick session: %w: %s", ErrProfileNotFound, selection.Name)
		}

		if err := a.creator.CreateSessionFromProfile(ctx, profile); err != nil {
			return fmt.Errorf("create session: %w", err)
		}
	}

	return a.execSession(ctx, selection.Name)
}

// execSession wraps executor calls with consistent error handling.
func (a *App) execSession(ctx context.Context, sessionName string) error {
	if err := a.executor.ExecIntoSession(ctx, sessionName); err != nil {
		return fmt.Errorf("exec session: %w", err)
	}

	return nil
}

// filterProfiles returns profiles whose names do not match any existing session name.
func filterProfiles(profiles []Profile, sessions []Session) []Profile {
	sessionNames := make(map[string]struct{}, len(sessions))
	for _, s := range sessions {
		sessionNames[s.Name] = struct{}{}
	}

	filtered := make([]Profile, 0, len(profiles))
	for _, p := range profiles {
		if _, exists := sessionNames[p.Name]; !exists {
			filtered = append(filtered, p)
		}
	}

	return filtered
}

// findProfile returns the profile with the given name from the slice.
func findProfile(profiles []Profile, name string) (Profile, bool) {
	for _, p := range profiles {
		if p.Name == name {
			return p, true
		}
	}

	return Profile{}, false
}
