package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/iamnotrodger/stdtype/ui/common"
	commandpalette "github.com/iamnotrodger/stdtype/ui/components/commadpalette"
	"github.com/iamnotrodger/stdtype/ui/theme"
)

type (
	restartMsg struct{}

	navMsg struct {
		screen screen
	}

	settingsMsg struct {
		field stdtype.SettingField
		value any
	}
)

func BuildCommands(theme theme.Theme, languages []string) []commandpalette.Command {
	normalStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	commandStyle := lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true)
	settingStye := lipgloss.NewStyle().Foreground(theme.Success).Bold(true)

	return []commandpalette.Command{
		{
			Label: normalStyle.Render("  stats"),
			Run:   func() tea.Msg { return navMsg{screenStats} },
		},
		{
			Label: commandStyle.Render("  restart"),
			Run:   func() tea.Msg { return restartMsg{} },
		},
		{
			Label:       settingStye.Render("  mode"),
			Subcommands: commandSettingsMode(settingStye),
		},
		{
			Label:       settingStye.Render("  time"),
			Subcommands: buildSettingsCommand(settingStye, stdtype.SettingDuration, common.DurationOptions),
		},
		{
			Label:       settingStye.Render("󰬴  word count"),
			Subcommands: buildSettingsCommand(settingStye, stdtype.SettingWordCount, common.WordOptions),
		},
		{
			Label:       settingStye.Render("󰗊  language"),
			Subcommands: buildSettingsCommand(settingStye, stdtype.SettingList, languages),
		},
		{
			Label:       settingStye.Render("  punctuation"),
			Subcommands: buildToggleCommand(settingStye, stdtype.SettingPunctuation),
		},
		{
			Label:       settingStye.Render("  numbers"),
			Subcommands: buildToggleCommand(settingStye, stdtype.SettingNumbers),
		},
	}
}

func buildSettingsCommand[T any](
	style lipgloss.Style,
	field stdtype.SettingField,
	values []T,
) []commandpalette.Command {
	commands := make([]commandpalette.Command, 0, len(values))

	for _, v := range values {
		commands = append(commands, commandpalette.Command{
			Label: style.Render(fmt.Sprint(v)),
			Run: func() tea.Msg {
				return settingsMsg{
					field: field,
					value: v,
				}
			},
		})
	}

	return commands
}

func buildToggleCommand(style lipgloss.Style, field stdtype.SettingField) []commandpalette.Command {
	return []commandpalette.Command{
		{
			Label: style.Render("on"),
			Run:   func() tea.Msg { return settingsMsg{field: field, value: true} },
		},
		{
			Label: style.Render("off"),
			Run:   func() tea.Msg { return settingsMsg{field: field, value: false} },
		},
	}
}

func commandSettingsMode(style lipgloss.Style) []commandpalette.Command {
	return []commandpalette.Command{
		{
			Label: style.Render("  time"),
			Run: func() tea.Msg {
				return settingsMsg{
					field: stdtype.SettingMode,
					value: stdtype.TimeMode.String(),
				}
			},
		},
		{
			Label: style.Render("󰬴  word"),
			Run: func() tea.Msg {
				return settingsMsg{
					field: stdtype.SettingMode,
					value: stdtype.WordsMode.String(),
				}
			},
		},
	}
}
