package theme

import (
	"charm.land/lipgloss/v2"
)

type ColorStyle struct {
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Faint     lipgloss.Style
	Error     lipgloss.Style
	Success   lipgloss.Style
}

func NewTextStyle(theme Theme) ColorStyle {
	var t ColorStyle

	t.Primary = lipgloss.NewStyle().Foreground(theme.Primary)
	t.Secondary = lipgloss.NewStyle().Foreground(theme.Secondary)
	t.Faint = lipgloss.NewStyle().Foreground(theme.Faint)
	t.Error = lipgloss.NewStyle().Foreground(theme.Error)
	t.Success = lipgloss.NewStyle().Foreground(theme.Success)

	return t
}
