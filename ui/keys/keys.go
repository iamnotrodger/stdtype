package keys

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Restart    key.Binding
	Menu       key.Binding
	Backspace  key.Binding
	DeleteWord key.Binding
	Quit       key.Binding
}

var Keys = &KeyMap{
	Restart: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "restart"),
	),
	Menu: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "menu"),
	),
	Backspace: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete"),
	),
	DeleteWord: key.NewBinding(
		key.WithKeys("alt+backspace"),
		key.WithHelp("alt+backspace", "delete word"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quite"),
	),
}
