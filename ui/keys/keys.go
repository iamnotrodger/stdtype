package keys

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	Enter          key.Binding
	Restart        key.Binding
	CommandPalette key.Binding
	Quit           key.Binding
}

var Keys = &KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "enter"),
	),
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
		key.WithHelp("ctrl+c", "quite"),
	),
}
