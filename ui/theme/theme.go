package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Primary   color.Color
	Secondary color.Color
	Faint     color.Color
	Success   color.Color
	Warning   color.Color
	Error     color.Color

	PrimaryBorder   color.Color
	FaintBorder     color.Color
	SecondaryBorder color.Color
}

var Default = &Theme{
	Primary:         lipgloss.NoColor{},
	Secondary:       lipgloss.Blue,
	Faint:           lipgloss.BrightBlack,
	Success:         lipgloss.Green,
	Warning:         lipgloss.Yellow,
	Error:           lipgloss.Red,
	PrimaryBorder:   lipgloss.BrightBlack,
	SecondaryBorder: lipgloss.BrightBlack,
	FaintBorder:     lipgloss.BrightBlack,
}
