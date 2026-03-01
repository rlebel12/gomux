package tui

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rlebel12/gomux/internal/gomux"
)

// errUnexpectedModel indicates Bubbletea returned an unexpected model type.
var errUnexpectedModel = errors.New("unexpected model type")

// Picker provides a terminal UI for session selection.
type Picker struct{}

// NewPicker creates a new Picker.
func NewPicker() *Picker {
	return &Picker{}
}

// Pick displays a session and profile picker and returns the selected item.
// Returns gomux.ErrNoSelection if the user quits without selecting.
func (p *Picker) Pick(_ context.Context, sessions []gomux.Session, profiles []gomux.Profile) (gomux.Selection, error) {
	model := newModel(sessions, profiles)
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		return gomux.Selection{}, fmt.Errorf("run picker: %w", err)
	}

	m, ok := finalModel.(*Model)
	if !ok {
		return gomux.Selection{}, fmt.Errorf("run picker: %w: %T", errUnexpectedModel, finalModel)
	}

	if m.Quitting() {
		return gomux.Selection{}, gomux.ErrNoSelection
	}

	return m.Result(), nil
}
