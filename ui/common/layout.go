package common

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/iamnotrodger/stdytype/ui/keys"
	"github.com/iamnotrodger/stdytype/ui/themes"
)

type Layout struct {
	Theme  themes.Theme
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
		return l.Theme.Dim.Render(DefaultFooter())
	}

	return l.Footer
}

func DefaultFooter() string {
	keys := []key.Binding{keys.Keys.Restart, keys.Keys.Menu, keys.Keys.Quit}
	out := make([]string, 0, len(keys))

	for _, k := range keys {
		h := k.Help()
		out = append(out, h.Key+" "+h.Desc)
	}

	return strings.Join(out, " · ")
}
