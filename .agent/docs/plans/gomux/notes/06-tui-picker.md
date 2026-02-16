# Phase 06: TUI Picker and Styling

## Summary

Implemented the TUI picker and full view rendering with lipgloss styling. Created the Picker type that wraps the Bubbletea Model and manages the program lifecycle. Implemented complete View() rendering for both list and input modes with proper formatting, relative time display, and styled UI elements.

## Files

- `internal/tui/picker.go` - Created Picker struct with Pick method that runs Bubbletea program
- `internal/tui/styles.go` - Created lipgloss style definitions for all UI elements
- `internal/tui/model.go` - Implemented View() method with renderListView, renderInputView, formatMetadata, and relativeTime helpers

## Tests

Existing test suite verified before implementation:
- Location: `internal/tui/model_test.go`
- Tests passing: 16 (4 test groups with subtests)
- Covers: List navigation, list actions (select/quit/new), input view behavior, name validation

## Implementation Notes

- Extracted renderTitle and renderFooter helper methods to eliminate duplication between list and input view rendering
- Used strings.Builder for efficient string concatenation in view rendering
- relativeTime function kept simple and readable with clear threshold checks
- Picker.Pick includes type assertion safety check after Bubbletea program run
- All style variables are package-level constants for consistency across the UI

## Issues

<!-- Problems encountered and how they were resolved -->
