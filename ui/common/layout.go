package common

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/iamnotrodger/stdytype/ui/keys"
	"github.com/iamnotrodger/stdytype/ui/theme"
)

type Layout struct {
	Style  theme.ColorStyle
	Header string
	Main   string
	Footer string
	Width  int
	Height int
}

func (l Layout) View() string {
	head := lipgloss.PlaceHorizontal(l.Width, lipgloss.Center, l.Header)
	foot := lipgloss.PlaceHorizontal(l.Width, lipgloss.Center, l.GetFooter())

	main := l.Main

	height := max(1, l.Height-lipgloss.Height(head)-lipgloss.Height(foot))
	if lines := strings.Split(l.Main, "\n"); len(lines) > height {
		main = strings.Join(lines[:height], "\n")
	}

	mid := lipgloss.Place(l.Width, height, lipgloss.Center, lipgloss.Center, main)

	return lipgloss.JoinVertical(lipgloss.Left, head, mid, foot)
}

func (l Layout) GetFooter() string {
	if l.Footer == "" {
		return l.Style.Faint.Render(DefaultFooter())
	}

	return l.Footer
}

func DefaultFooter() string {
	keys := []struct {
		binding key.Binding
		icon    string
	}{
		{
			keys.Keys.Restart,
			"",
		},
		{
			keys.Keys.CommandPalette,
			"󰍜",
		},
		{
			keys.Keys.Quit,
			"󰈆",
		},
	}

	footer := make([]string, 0, len(keys))

	for _, k := range keys {
		h := k.binding.Help()
		footer = append(footer, h.Key+" "+k.icon)
	}

	return strings.Join(footer, " · ")
}
