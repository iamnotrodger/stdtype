package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/keys"
	"github.com/iamnotrodger/stdytype/ui/typing"
)

type screen int

const (
	screenTyping screen = iota
	screenStats
	screenResult
)

type Model struct {
	wordSource stdtype.WordSource
	resultRepo stdtype.ResultRepo

	settings stdtype.Settings
	screen   screen
	keys     *keys.KeyMap

	typing typing.Model
}

func NewModel(
	wordSource stdtype.WordSource,
	resultRepo stdtype.ResultRepo,
	settings stdtype.Settings,
) Model {
	return Model{
		wordSource: wordSource,
		resultRepo: resultRepo,
		settings:   settings,
		screen:     screenTyping,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() tea.View {
	return tea.NewView("hello world")
}
