package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/iamnotrodger/stdytype/ui/components/result"
	"github.com/iamnotrodger/stdytype/ui/components/typing"
	"github.com/iamnotrodger/stdytype/ui/theme"
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
