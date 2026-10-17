package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	"github.com/iamnotrodger/stdytype/ui/components/result"
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

	typing typing.Model
	result result.Model

	screen screen
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

	return Model{
		appCtx:     appCtx,
		wordSource: wordSource,
		statsRepo:  statsRepo,
		typing:     typing.NewModel(appCtx, wordSource),
		result:     result.NewMode(appCtx),
		screen:     screenTyping,
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
	v := tea.NewView(m.viewScreen(m.screen))
	v.AltScreen = true

	return v
}

func (m Model) viewScreen(s screen) string {
	switch s {
	case screenResult:
		return m.result.View()
	default:
		return m.typing.View()
	}
}

func (m Model) updateTyping(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.typing, cmd = m.typing.Update(msg)

	if m.typing.Completed() {
		return m.onTestComplete()
	}

	return m, cmd
}

func (m Model) onTestComplete() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	result := m.typing.Result()

	m.statsRepo.CreateResult(result) // nolint:errcheck,gosec

	m.result = m.result.SetResult(result)
	m.screen = screenResult

	return m, cmd
}
