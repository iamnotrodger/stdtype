package keys

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Restart        key.Binding
	CommandPalette key.Binding
	Quit           key.Binding
}

var Keys = &KeyMap{
	Restart: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "restart"),
	),
	CommandPalette: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "menu"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}
