package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	commandpalette "github.com/iamnotrodger/stdytype/ui/components/commadpalette"
	"github.com/iamnotrodger/stdytype/ui/components/result"
	"github.com/iamnotrodger/stdytype/ui/components/typing"
	appctx "github.com/iamnotrodger/stdytype/ui/context"
	"github.com/iamnotrodger/stdytype/ui/keys"
	"github.com/iamnotrodger/stdytype/ui/theme"
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

	typing  typing.Model
	result  result.Model
	palette commandpalette.Model

	screen      screen
	paletteOpen bool

	width, height int
}

func NewModel(
	wordSource stdtype.WordSource,
	statsRepo stdtype.StatsRepo,
	settings stdtype.Settings,
) Model {
	appCtx := &appctx.Context{
		Theme:      *theme.Default,
		ColorStyle: theme.NewTextStyle(*theme.Default),
	}

	return Model{
		appCtx:     appCtx,
		wordSource: wordSource,
		statsRepo:  statsRepo,
		screen:     screenTyping,
		typing:     typing.NewModel(appCtx, wordSource, &settings),
		result:     result.NewMode(appCtx),
		palette: commandpalette.NewModel(commandpalette.Config{
			Commands: Commands,
		}),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.typing.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.onResize(msg)
	case typing.TickMsg:
		return m.updateTyping(msg)
	case animation.Msg:
		return m.updateTyping(msg)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Keys.Restart):
			return m.onRestart()
		case key.Matches(msg, keys.Keys.CommandPalette):
			return m.onPalette()
		}

		if m.paletteOpen {
			return m.updatePalette(msg)
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
	content := m.viewScreen(m.screen)
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(content).Z(0),
	}

	if m.paletteOpen {
		palette := m.palette.View()
		paletteWidth, paletteHeight := lipgloss.Size(palette)

		layers = append(layers,
			lipgloss.NewLayer(palette).
				X((m.width-paletteWidth)/2).
				Y((m.height-paletteHeight)/4).
				Z(1),
		)
	}

	comp := lipgloss.NewCompositor(layers...)
	v := tea.NewView(comp.Render())
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
		return m.onComplete()
	}

	return m, cmd
}

func (m Model) updatePalette(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.palette, cmd = m.palette.Update(msg)

	return m, cmd
}

func (m Model) onResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	w, h := msg.Width, msg.Height

	m.width, m.height = w, h
	m.typing.SetSize(w, h)
	m.result.SetSize(w, h)
	m.palette.SetSize(w, h)

	return m, nil
}

func (m Model) onComplete() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	result := m.typing.Result()

	// TODO: log error.
	m.statsRepo.CreateResult(result) // nolint:errcheck,gosec
	m.result.SetResult(result)
	m.screen = screenResult

	return m, cmd
}

func (m Model) onRestart() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.typing, cmd = m.typing.Restart()
	m.screen = screenTyping

	return m, cmd
}

func (m Model) onPalette() (tea.Model, tea.Cmd) {
	if m.paletteOpen {
		m.paletteOpen = false

		return m, nil
	}

	m.paletteOpen = true
	m.palette.Reset()

	return m, m.palette.Focus()
}
