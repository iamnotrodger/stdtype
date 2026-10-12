package keys

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Quit key.Binding
}

var Keys = &KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quite"),
	),
}
