package animation

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
