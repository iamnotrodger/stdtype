package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const FPS = 70

type Msg struct {
	ID   int
	Time time.Time
}

func Tick(id int) tea.Cmd {
	return tea.Tick(
		time.Second/FPS,
		func(t time.Time) tea.Msg {
			return Msg{ID: id, Time: t}
		},
	)
}
