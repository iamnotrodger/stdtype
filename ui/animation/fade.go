package animation

import (
	"fmt"
	"math"

	"charm.land/lipgloss/v2"
)

const fadeStep = 1.0 / 15.0

type Fade struct {
	Value  float64
	Active bool
}

func NewFade() Fade {
	return Fade{
		Value: 1.0,
	}
}

func (f *Fade) Start() {
	f.Value = 0
	f.Active = true
}

func (f *Fade) Update() bool {
	if !f.Active {
		return false
	}

	f.Value += fadeStep

	if f.Value >= 1 {
		f.Value = 1
		f.Active = false

		return false
	}

	return true
}

func FadeRender(fade float64, str string) string {
	if fade >= 1 {
		return str
	}

	return FadeStyle(fade).Render(str)
}

func FadeStyle(t float64) lipgloss.Style {
	const gray = 0x88

	v := int(math.Round(float64(gray) * t))
	hex := fmt.Sprintf("#%02x%02x%02x", v, v, v)

	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}
