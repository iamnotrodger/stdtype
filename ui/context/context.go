package context

import (
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/themes"
)

type Context struct {
	Settings stdtype.Settings
	Theme    themes.Theme
	Width    int
	Height   int
}

// Dimensions returns width and height in that order.
func (c *Context) Dimension() (int, int) {
	w, h := c.Width, c.Height

	if w <= 0 {
		w = 80
	}

	if h <= 0 {
		h = 24
	}

	return w, h
}
