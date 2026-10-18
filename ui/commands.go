package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	commandpalette "github.com/iamnotrodger/stdytype/ui/components/commadpalette"
	"github.com/iamnotrodger/stdytype/ui/theme"
)

type restartMsg struct{}
type navMsg struct{ screen screen }

func BuildCommands(theme theme.Theme) []commandpalette.Command {
	navStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	commandStyle := lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true)
	// settingStye := lipgloss.NewStyle().Foreground(theme.Success).Bold(true)

	return []commandpalette.Command{
		{
			Label: navStyle.Render("  stats"),
			Run:   func() tea.Msg { return navMsg{screenStats} },
		},
		{
			Label: commandStyle.Render("  restart"),
			Run:   func() tea.Msg { return restartMsg{} },
		},
		// {
		// 	Label: settingStye.Render("  mode"),
		// },
		// {
		// 	Label: settingStye.Render("  time"),
		// },
		// {
		// 	Label: settingStye.Render("󰬴  word count"),
		// },
		// {
		// 	Label: settingStye.Render("󰗊  language"),
		// },
	}
}
