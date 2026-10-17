package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const FPS = 70

type Msg time.Time

func Tick() tea.Cmd {
	return tea.Tick(
		time.Second/FPS,
		func(t time.Time) tea.Msg {
			return Msg(t)
		},
	)
}
