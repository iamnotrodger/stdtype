package themes

import "charm.land/lipgloss/v2"

func Default() Theme {
	green := lipgloss.Color("2")
	red := lipgloss.Color("1")
	gray := lipgloss.Color("8")
	blue := lipgloss.Color("4")
	magenta := lipgloss.Color("5")
	cyan := lipgloss.Color("6")

	return Theme{
		Correct:   lipgloss.NewStyle().Foreground(green),
		Incorrect: lipgloss.NewStyle().Foreground(red).Underline(true),
		Pending:   lipgloss.NewStyle().Foreground(gray),
		Caret:     lipgloss.NewStyle().Reverse(true),
		Accent:    lipgloss.NewStyle().Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(gray),
		Stat:      lipgloss.NewStyle().Bold(true),
		StatLabel: lipgloss.NewStyle().Foreground(gray),
		Title:     lipgloss.NewStyle().Bold(true),
		Popup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderBottomForeground(gray).
			Padding(1, 2),
		Nav:     lipgloss.NewStyle().Foreground(blue),
		Command: lipgloss.NewStyle().Foreground(magenta),
		Setting: lipgloss.NewStyle().Foreground(cyan),
	}
}
