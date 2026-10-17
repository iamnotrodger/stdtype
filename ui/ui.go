package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	"github.com/iamnotrodger/stdytype/ui/components/typing"
	appctx "github.com/iamnotrodger/stdytype/ui/context"
	"github.com/iamnotrodger/stdytype/ui/keys"
	"github.com/iamnotrodger/stdytype/ui/themes"
)

type screen int

const (
	screenTyping screen = iota
	screenResult
	screenStats
)

type Model struct {
	appCtx     *appctx.Context
	wordSource stdtype.WordSource
	statsRepo  stdtype.StatsRepo

	screen screen

	typing typing.Model
	result *stdtype.Result
}

func NewModel(
	wordSource stdtype.WordSource,
	statsRepo stdtype.StatsRepo,
	settings stdtype.Settings,
) Model {
	appCtx := &appctx.Context{
		Settings: settings,
		Theme:    themes.Default(),
	}

	// TODO: handle this error
	typing, _ := typing.NewModel(appCtx, wordSource)

	return Model{
		appCtx:     appCtx,
		wordSource: wordSource,
		statsRepo:  statsRepo,
		screen:     screenTyping,
		typing:     typing,
	}
}

func (m Model) Init() tea.Cmd {
	return m.typing.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.appCtx.Width = msg.Width
		m.appCtx.Height = msg.Height

		return m, nil

	case typing.TickMsg:
		return m.updateTyping(msg)
	case animation.Msg:
		return m.updateTyping(msg)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Keys.Restart):
			m.typing, cmd = m.typing.Restart()
			m.screen = screenTyping

			return m, cmd
		}

		// nolint:exhaustive
		switch m.screen {
		case screenTyping:
			return m.updateTyping(msg)
		case screenResult:
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() tea.View {
	v := tea.NewView(m.typing.View())
	v.AltScreen = true

	return v
}

func (m Model) updateTyping(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.typing, cmd = m.typing.Update(msg)
	if m.typing.Completed() {
		m.result = m.typing.Result()
		// TODO: log error
		m.statsRepo.CreateResult(m.result) // nolint:errcheck,gosec
		m.screen = screenResult
	}

	return m, cmd
}
