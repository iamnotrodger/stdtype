package ui

import (
	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pkg/errors"

	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/iamnotrodger/stdtype/ui/animation"
	commandpalette "github.com/iamnotrodger/stdtype/ui/components/commadpalette"
	"github.com/iamnotrodger/stdtype/ui/components/result"
	"github.com/iamnotrodger/stdtype/ui/components/stats"
	"github.com/iamnotrodger/stdtype/ui/components/typing"
	"github.com/iamnotrodger/stdtype/ui/keys"
	"github.com/iamnotrodger/stdtype/ui/theme"
)

type screen int

const (
	screenTyping screen = iota
	screenResult
	screenStats
)

type Model struct {
	wordSource   stdtype.WordSource
	statsRepo    stdtype.StatsRepo
	settingsRepo stdtype.SettingsRepo

	settings *stdtype.Settings

	typing  typing.Model
	result  result.Model
	stats   stats.Model
	palette commandpalette.Model

	screen      screen
	paletteOpen bool

	err      error
	errStyle lipgloss.Style

	width, height int
}

func NewModel(
	wordSource stdtype.WordSource,
	statsRepo stdtype.StatsRepo,
	settingsRepo stdtype.SettingsRepo,
) (Model, error) {
	settings, err := settingsRepo.GetSettings()
	if err != nil {
		return Model{}, errors.Wrap(err, "settingsRepo.GetSettings")
	}

	typingModel, err := typing.NewModel(wordSource, &settings)
	if err != nil {
		return Model{}, errors.Wrap(err, "typing.NewModel")
	}

	m := Model{
		wordSource:   wordSource,
		statsRepo:    statsRepo,
		settingsRepo: settingsRepo,
		settings:     &settings,
		screen:       screenTyping,
		typing:       typingModel,
		result:       result.NewModel(),
		stats:        stats.NewModel(),
		palette:      commandpalette.NewModel(BuildCommands(*theme.Default, wordSource.Lists())),
		errStyle:     lipgloss.NewStyle().Foreground(theme.Default.Error),
	}

	m.typing.SetStyle(typingStyles(*theme.Default))
	m.result.SetStyle(resultStyles(*theme.Default))
	m.stats.SetStyle(statsStyles(*theme.Default))
	m.palette.SetStyle(paletteStyles(*theme.Default))

	return m, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.typing.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.onResize(msg)
	case typing.TickMsg:
		return m.updateTyping(msg)
	case cursor.BlinkMsg:
		return m.updateBackground(msg)
	case animation.Msg:
		return m.updateTyping(msg)
	case typing.ResultMsg:
		return m.onComplete(msg.Result)
	case stats.ResultMsg:
		return m.onResult(msg.Result)
	case commandpalette.CommandMsg:
		return m.onCommand(msg)
	case restartMsg:
		return m.onRestart()
	case navMsg:
		return m.onNavigate(msg)
	case settingsMsg:
		return m.onSettings(msg)
	case typing.ErrMsg:
		m.err = msg.Err

		return m, nil

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
		case screenStats:
			return m.updateStats(msg)
		case screenResult:
			return m, nil
		}
	}

	return m.updateBackground(msg)
}

func (m Model) View() tea.View {
	content := m.viewScreen(m.screen)
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(content).Z(0),
	}

	if m.paletteOpen {
		palette := m.palette.View()
		paletteWidth, _ := lipgloss.Size(palette)

		layers = append(layers,
			lipgloss.NewLayer(palette).
				X((m.width-paletteWidth)/2).
				Y(m.height/4).
				Z(1),
		)
	}

	if m.err != nil {
		errView := m.errStyle.Render("error: " + m.err.Error())
		errWidth, _ := lipgloss.Size(errView)

		// The row above the footer is blank on every screen.
		layers = append(layers,
			lipgloss.NewLayer(errView).
				X(max((m.width-errWidth)/2, 0)).
				Y(max(m.height-2, 0)).
				Z(2),
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
	case screenStats:
		return m.stats.View()
	default:
		return m.typing.View()
	}
}

func (m Model) updateTyping(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.typing, cmd = m.typing.Update(msg)

	return m, cmd
}

func (m Model) updateStats(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.stats, cmd = m.stats.Update(msg)

	return m, cmd
}

func (m Model) updatePalette(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.palette, cmd = m.palette.Update(msg)

	return m, cmd
}

// Each component ignores cursor blinks it does not own.
func (m Model) updateBackground(msg tea.Msg) (tea.Model, tea.Cmd) {
	var paletteCmd, typingCmd tea.Cmd

	if m.paletteOpen {
		m.palette, paletteCmd = m.palette.Update(msg)
	}

	if m.screen == screenTyping {
		m.typing, typingCmd = m.typing.Update(msg)
	}

	return m, tea.Batch(paletteCmd, typingCmd)
}

func (m Model) onResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	w, h := msg.Width, msg.Height

	m.width, m.height = w, h
	m.typing.SetSize(w, h)
	m.result.SetSize(w, h)
	m.palette.SetSize(w, h)
	m.stats.SetSize(w, h)

	return m, nil
}

func (m Model) onComplete(result *stdtype.Result) (tea.Model, tea.Cmd) {
	if !result.Invalid {
		if err := m.statsRepo.CreateResult(result); err != nil {
			m.err = errors.Wrap(err, "statsRepo.CreateResult")
		}
	}

	return m.onResult(result)
}

func (m Model) onResult(result *stdtype.Result) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.result.SetResult(result)
	m.screen = screenResult

	return m, cmd
}

func (m Model) onRestart() (Model, tea.Cmd) {
	var cmd tea.Cmd

	m.typing, cmd = m.typing.Restart()
	m.screen = screenTyping
	m.paletteOpen = false
	m.err = nil

	return m, cmd
}

func (m Model) onSettings(msg settingsMsg) (tea.Model, tea.Cmd) {
	switch msg.field {
	case stdtype.SettingMode:
		value, ok := msg.value.(string)
		if !ok {
			return m, nil
		}
		m.settings.Mode = value
	case stdtype.SettingList:
		value, ok := msg.value.(string)
		if !ok {
			return m, nil
		}
		m.settings.WordList = value
	case stdtype.SettingDuration:
		value, ok := msg.value.(int)
		if !ok {
			return m, nil
		}
		m.settings.Mode = stdtype.TimeMode.String()
		m.settings.Duration = value
	case stdtype.SettingWordCount:
		value, ok := msg.value.(int)
		if !ok {
			return m, nil
		}
		m.settings.Mode = stdtype.WordsMode.String()
		m.settings.WordCount = value
	case stdtype.SettingPunctuation:
		value, ok := msg.value.(bool)
		if !ok {
			return m, nil
		}
		m.settings.Punctuation = value
	case stdtype.SettingNumbers:
		value, ok := msg.value.(bool)
		if !ok {
			return m, nil
		}
		m.settings.Numbers = value
	}

	err := m.settingsRepo.UpdateSettings(*m.settings)

	m, cmd := m.onRestart()
	if err != nil {
		m.err = errors.Wrap(err, "settingsRepo.UpdateSettings")
	}

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

func (m Model) onCommand(msg commandpalette.CommandMsg) (tea.Model, tea.Cmd) {
	m.paletteOpen = false

	return m, msg.Command.Run
}

func (m Model) onNavigate(msg navMsg) (tea.Model, tea.Cmd) {
	if msg.screen == screenStats {
		results, err := m.statsRepo.ListResults()
		if err != nil {
			m.err = errors.Wrap(err, "statsRepo.ListResults")

			return m, nil
		}

		m.stats.SetResults(results)
	}

	m.screen = msg.screen

	return m, nil
}
