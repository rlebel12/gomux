package tui

import "github.com/charmbracelet/lipgloss"

//nolint:gochecknoglobals // UI styles are legitimately package-level constants
var (
	// titleStyle renders the "gomux" header text.
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// selectedStyle renders the selected session name.
	selectedStyle = lipgloss.NewStyle().Bold(true)

	// normalStyle renders unselected session names.
	normalStyle = lipgloss.NewStyle()

	// metadataStyle renders secondary information (windows, attached, time).
	metadataStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// attachedStyle renders the "attached" badge.
	attachedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	// footerStyle renders the help bar at the bottom.
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// errorStyle renders validation error messages.
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// warningStyle renders validation warning messages.
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

const (
	// cursorGlyphSelected is the cursor indicator for the selected item.
	cursorGlyphSelected = "▸ "

	// cursorGlyphUnselected is the spacing for unselected items.
	cursorGlyphUnselected = "  "
)
