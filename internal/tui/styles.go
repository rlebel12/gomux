package tui

import "github.com/charmbracelet/lipgloss"

// titleStyle renders the "gomux" header text.
func titleStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("240")) }

// selectedStyle renders the selected session name.
func selectedStyle() lipgloss.Style { return lipgloss.NewStyle().Bold(true) }

// normalStyle renders unselected session names.
func normalStyle() lipgloss.Style { return lipgloss.NewStyle() }

// metadataStyle renders secondary information (windows, attached, time).
func metadataStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("240")) }

// attachedStyle renders the "attached" badge.
func attachedStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("2")) }

// footerStyle renders the help bar at the bottom.
func footerStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("240")) }

// errorStyle renders validation error messages.
func errorStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("1")) }

// warningStyle renders validation warning messages.
func warningStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) }

const (
	// cursorGlyphSelected is the cursor indicator for the selected item.
	cursorGlyphSelected = "▸ "

	// cursorGlyphUnselected is the spacing for unselected items.
	cursorGlyphUnselected = "  "
)
