package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rlebel12/gomux/internal/gomux"
)

// Picker provides a terminal UI for session selection.
type Picker struct{}

// NewPicker creates a new Picker.
func NewPicker() *Picker {
	return &Picker{}
}

// Pick displays a session picker and returns the selected session name.
// Returns gomux.ErrNoSelection if the user quits without selecting.
func (p *Picker) Pick(_ context.Context, sessions []gomux.Session) (string, error) {
	model := newModel(sessions)
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("run picker: %w", err)
	}

	m, ok := finalModel.(Model)
	if !ok {
		// This should never happen unless Bubbletea returns wrong type
		//nolint:err113 // Dynamic error is appropriate for unexpected internal state
		return "", fmt.Errorf("run picker: unexpected model type %T", finalModel)
	}

	if m.Quitting() {
		return "", gomux.ErrNoSelection
	}

	return m.Result(), nil
}
