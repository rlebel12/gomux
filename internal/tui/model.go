// Package tui provides the terminal user interface for gomux.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rlebel12/gomux/internal/gomux"
)

// Model represents the Bubbletea TUI model with two views: list and input.
//
//nolint:recvcheck // Bubbletea interface requires value receivers; helper methods use pointers
type Model struct {
	sessions  []gomux.Session
	result    string
	err       string
	warning   string
	textInput textinput.Model
	cursor    int
	inputMode bool
	quitting  bool
}

// newModel creates a new Model with the provided sessions.
func newModel(sessions []gomux.Session) Model {
	ti := textinput.New()
	ti.Placeholder = "session name"
	ti.Focus()

	return Model{
		sessions:  sessions,
		cursor:    0,
		inputMode: false,
		textInput: ti,
	}
}

// Init implements tea.Model.
//
//nolint:gocritic // Model must match tea.Model interface signature
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
//
//nolint:gocritic,ireturn // Model must match tea.Model interface signature
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.inputMode {
			return m.handleInputKey(keyMsg)
		}

		return m.handleListKey(keyMsg)
	}

	return m, nil
}

// View implements tea.Model.
//
//nolint:gocritic // Model must match tea.Model interface signature
func (m Model) View() string {
	if m.inputMode {
		return m.renderInputView()
	}

	return m.renderListView()
}

// Result returns the selected or entered session name.
//
//nolint:gocritic // Exported accessor uses value receiver for consistency
func (m Model) Result() string {
	return m.result
}

// Quitting returns whether the user quit without selecting.
//
//nolint:gocritic // Exported accessor uses value receiver for consistency
func (m Model) Quitting() bool {
	return m.quitting
}

// handleListKey handles key messages in list view mode.
//
//nolint:gocritic,ireturn // Must use value receiver to match Update signature
func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type { //nolint:exhaustive // Only handling specific keys
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			switch msg.Runes[0] {
			case 'j':
				m.moveCursorDown()
			case 'k':
				m.moveCursorUp()
			case 'n':
				m.inputMode = true
				m.textInput.Focus()
			case 'q':
				return m.quit()
			}
		}

	case tea.KeyDown:
		m.moveCursorDown()

	case tea.KeyUp:
		m.moveCursorUp()

	case tea.KeyEnter:
		if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
			m.result = m.sessions[m.cursor].Name
			return m, tea.Quit
		}

	case tea.KeyCtrlC:
		return m.quit()
	}

	return m, nil
}

// handleInputKey handles key messages in input view mode.
//
//nolint:gocritic,ireturn // Must use value receiver to match Update signature
func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type { //nolint:exhaustive // Only handling specific keys
	case tea.KeyEnter:
		// Validate and submit if valid
		value := m.textInput.Value()
		if m.err == "" && value != "" {
			m.result = value
			return m, tea.Quit
		}
		// Blocked by error or empty value

	case tea.KeyEsc:
		m.cancelInput()

	case tea.KeyCtrlC:
		return m.quit()

	default:
		// Forward to textinput and then validate
		var cmd tea.Cmd

		m.textInput, cmd = m.textInput.Update(msg)
		m.validate()

		return m, cmd
	}

	return m, nil
}

// validate checks the current input value for errors and warnings.
func (m *Model) validate() {
	value := m.textInput.Value()
	m.err = ""
	m.warning = ""

	if value == "" {
		return
	}

	// Check for invalid characters (dots and colons).
	if strings.Contains(value, ".") {
		m.err = "Session names cannot contain dots"
		return
	}

	if strings.Contains(value, ":") {
		m.err = "Session names cannot contain colons"
		return
	}

	// Check for duplicate name.
	for _, session := range m.sessions {
		if session.Name == value {
			m.warning = "Session already exists - will attach instead of creating"
			return
		}
	}
}

// moveCursorDown moves the cursor down one position, clamped to list bounds.
func (m *Model) moveCursorDown() {
	if m.cursor < len(m.sessions)-1 {
		m.cursor++
	}
}

// moveCursorUp moves the cursor up one position, clamped to list bounds.
func (m *Model) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// quit marks the model as quitting and returns the quit command.
//
//nolint:gocritic,ireturn // Must use value receiver to match Update signature
func (m Model) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	return m, tea.Quit
}

// cancelInput returns to list mode and clears input state.
func (m *Model) cancelInput() {
	m.inputMode = false
	m.textInput.SetValue("")
	m.err = ""
	m.warning = ""
}

// renderListView renders the list view with session names.
//
//nolint:gocritic // Model uses value receiver to match tea.Model interface pattern
func (m Model) renderListView() string {
	var b strings.Builder

	m.renderTitle(&b)

	// Session list
	for i, session := range m.sessions {
		// Cursor glyph
		if i == m.cursor {
			b.WriteString(cursorGlyphSelected)
		} else {
			b.WriteString(cursorGlyphUnselected)
		}

		// Session name
		if i == m.cursor {
			b.WriteString(selectedStyle.Render(session.Name))
		} else {
			b.WriteString(normalStyle.Render(session.Name))
		}

		b.WriteString("\n")

		// Metadata line
		b.WriteString("  ")

		metadata := m.formatMetadata(session)
		b.WriteString(metadataStyle.Render(metadata))
		b.WriteString("\n\n")
	}

	m.renderFooter(&b, "↑↓ navigate · enter attach · n new · q quit")

	return b.String()
}

// renderInputView renders the input view for creating a new session.
//
//nolint:gocritic // Model uses value receiver to match tea.Model interface pattern
func (m Model) renderInputView() string {
	var b strings.Builder

	m.renderTitle(&b)

	// Input label
	b.WriteString("New session name:\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Error or warning
	if m.err != "" {
		b.WriteString(errorStyle.Render(m.err))
		b.WriteString("\n\n")
	} else if m.warning != "" {
		b.WriteString(warningStyle.Render(m.warning))
		b.WriteString("\n\n")
	}

	m.renderFooter(&b, "enter confirm · esc cancel")

	return b.String()
}

// formatMetadata creates the metadata line for a session.
//
//nolint:gocritic // Model uses value receiver to match tea.Model interface pattern
func (m Model) formatMetadata(session gomux.Session) string {
	var parts []string

	// Windows count
	if session.Windows == 1 {
		parts = append(parts, "1 window")
	} else {
		parts = append(parts, fmt.Sprintf("%d windows", session.Windows))
	}

	// Attached badge
	if session.Attached {
		parts = append(parts, attachedStyle.Render("attached"))
	}

	// Created time
	relTime := relativeTime(session.Created, time.Now())
	parts = append(parts, "created "+relTime)

	return strings.Join(parts, " · ")
}

// relativeTime returns a human-readable relative time string.
func relativeTime(created, now time.Time) string {
	const (
		hoursPerDay = 24
		daysPerWeek = 7
	)

	duration := now.Sub(created)

	if duration < time.Minute {
		return "just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	}

	if duration < hoursPerDay*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}

	if duration < daysPerWeek*hoursPerDay*time.Hour {
		days := int(duration.Hours() / hoursPerDay)
		return fmt.Sprintf("%dd ago", days)
	}

	weeks := int(duration.Hours() / (hoursPerDay * daysPerWeek))

	return fmt.Sprintf("%dw ago", weeks)
}

// renderTitle writes the "gomux" header to the builder.
//
//nolint:gocritic // Model uses value receiver to match tea.Model interface pattern
func (m Model) renderTitle(b *strings.Builder) {
	b.WriteString(titleStyle.Render("gomux"))
	b.WriteString("\n\n")
}

// renderFooter writes the footer help text to the builder.
//
//nolint:gocritic // Model uses value receiver to match tea.Model interface pattern
func (m Model) renderFooter(b *strings.Builder, text string) {
	b.WriteString(footerStyle.Render(text))
}
