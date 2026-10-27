package ui

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	commandpalette "github.com/iamnotrodger/stdtype/ui/components/commadpalette"
	"github.com/iamnotrodger/stdtype/ui/components/result"
	"github.com/iamnotrodger/stdtype/ui/components/stats"
	"github.com/iamnotrodger/stdtype/ui/components/typing"
	"github.com/iamnotrodger/stdtype/ui/theme"
)

func typingStyles(t theme.Theme) typing.Styles {
	faint := lipgloss.NewStyle().Foreground(t.Faint)

	return typing.Styles{
		ConfigActive:    lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		ConfigInactive:  faint,
		ConfigSeparator: faint,
		Footer:          faint,
		TextArea: typing.TextAreaStyles{
			Untyped:    faint,
			Correct:    lipgloss.NewStyle().Foreground(t.Success),
			Incorrect:  lipgloss.NewStyle().Foreground(t.Error).Underline(true),
			Cursor:     lipgloss.NewStyle().Reverse(true),
			CursorText: faint,
		},
	}
}

func resultStyles(t theme.Theme) result.Styles {
	primary := lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	faint := lipgloss.NewStyle().Foreground(t.Faint)
	failure := lipgloss.NewStyle().Foreground(t.Error)

	return result.Styles{
		Title:   primary,
		Invalid: failure,
		Label:   primary,
		Value:   faint,
		Footer:  faint,
		Graph: result.GraphStyles{
			WPM: lipgloss.NewStyle().Foreground(t.Success),
			Raw: faint,
			Err: failure,
		},
	}
}

func statsStyles(t theme.Theme) stats.Styles {
	tbl := table.DefaultStyles()
	tbl.Header = tbl.Header.MarginBottom(1).Bold(true)
	tbl.Selected = tbl.Selected.Foreground(t.Primary).Background(t.Secondary)

	return stats.Styles{
		Title:  lipgloss.NewStyle(),
		Label:  lipgloss.NewStyle().Foreground(t.Success),
		Value:  lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Footer: lipgloss.NewStyle().Foreground(t.Faint),
		Table:  tbl,
	}
}

func paletteStyles(t theme.Theme) commandpalette.Styles {
	faint := lipgloss.NewStyle().Foreground(t.Faint)

	return commandpalette.Styles{
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Faint).
			Padding(1, 2),
		Divider: faint,
		Empty:   faint,
		Item:    lipgloss.NewStyle().Foreground(t.Primary),
	}
}
