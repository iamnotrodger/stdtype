package themes

import "charm.land/lipgloss/v2"

type Theme struct {
	Correct   lipgloss.Style
	Incorrect lipgloss.Style
	Pending   lipgloss.Style
	Caret     lipgloss.Style
	Accent    lipgloss.Style
	Dim       lipgloss.Style
	Stat      lipgloss.Style
	StatLabel lipgloss.Style
	Title     lipgloss.Style
	Popup     lipgloss.Style

	Nav     lipgloss.Style
	Command lipgloss.Style
	Setting lipgloss.Style
}
