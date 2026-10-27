package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdtype"
	commandpalette "github.com/iamnotrodger/stdtype/ui/components/commadpalette"
	"github.com/iamnotrodger/stdtype/ui/components/stats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModel(t *testing.T) {
	t.Parallel()

	t.Run("should load settings from repo when created", func(t *testing.T) {
		t.Parallel()

		settings := stdtype.Settings{Mode: "time", Duration: 60}

		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, &fakeSettingsRepo{settings: settings})

		require.NoError(t, err)
		assert.Equal(t, settings, *m.settings)
	})

	t.Run("should fail when settings cannot be loaded", func(t *testing.T) {
		t.Parallel()

		_, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, &fakeSettingsRepo{err: errDiskFull})

		assert.ErrorIs(t, err, errDiskFull)
	})
}

func TestFinishTest(t *testing.T) {
	t.Parallel()

	t.Run("should save result and show it when typing test finished", func(t *testing.T) {
		t.Parallel()

		repo := &fakeStatsRepo{}
		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, repo, settingsRepo)
		require.NoError(t, err)

		m, cmd := typeString(m, "ab cd")
		require.NotNil(t, cmd)
		m, _ = update(m, cmd())

		assert.Equal(t, screenResult, m.screen)
		assert.Len(t, repo.results, 1)
	})

	t.Run("should show result without saving when picked from stats", func(t *testing.T) {
		t.Parallel()

		repo := &fakeStatsRepo{}
		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, repo, settingsRepo)
		require.NoError(t, err)
		m.screen = screenStats

		m, _ = update(m, stats.ResultMsg{Result: &stdtype.Result{Mode: "words", Amount: 2}})

		assert.Equal(t, screenResult, m.screen)
		assert.Empty(t, repo.results)
	})
}

func TestKeys(t *testing.T) {
	t.Parallel()

	t.Run("should quit when ctrl+c pressed", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)

		_, cmd := update(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

		require.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("should return to typing and close palette when tab pressed", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)
		m.screen = screenResult
		m.paletteOpen = true

		m, _ = update(m, tea.KeyPressMsg{Code: tea.KeyTab})

		assert.Equal(t, screenTyping, m.screen)
		assert.False(t, m.paletteOpen)
	})

	t.Run("should ignore keys when on result screen", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)
		m.screen = screenResult

		m, cmd := update(m, keyPress('a'))

		assert.Nil(t, cmd)
		assert.Equal(t, screenResult, m.screen)
	})
}

func TestPalette(t *testing.T) {
	t.Parallel()

	t.Run("should toggle palette when esc pressed", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)

		m, _ = update(m, tea.KeyPressMsg{Code: tea.KeyEscape})
		assert.True(t, m.paletteOpen)

		m, _ = update(m, tea.KeyPressMsg{Code: tea.KeyEscape})
		assert.False(t, m.paletteOpen)
	})

	t.Run("should send keys to palette instead of typing when palette open", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)
		m, _ = update(m, tea.KeyPressMsg{Code: tea.KeyEscape})

		m, _ = typeString(m, "ab")

		assert.Equal(t, "ab", m.palette.Query())
	})

	t.Run("should close palette and run command when command chosen", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)
		m.paletteOpen = true

		m, cmd := update(m, commandpalette.CommandMsg{
			Command: commandpalette.Command{Run: func() tea.Msg { return restartMsg{} }},
		})

		assert.False(t, m.paletteOpen)
		require.NotNil(t, cmd)
		assert.Equal(t, restartMsg{}, cmd())
	})

	t.Run("should show stats when stats command chosen from palette", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)
		m, _ = update(m, tea.KeyPressMsg{Code: tea.KeyEscape})
		m, _ = typeString(m, "stats")

		// Each cmd yields the next msg: CommandMsg, then navMsg.
		m, cmd := update(m, tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		m, cmd = update(m, cmd())
		require.NotNil(t, cmd)
		m, _ = update(m, cmd())

		assert.Equal(t, screenStats, m.screen)
		assert.False(t, m.paletteOpen)
	})
}

func TestNavigate(t *testing.T) {
	t.Parallel()

	t.Run("should load saved results when stats screen opened", func(t *testing.T) {
		t.Parallel()

		repo := &fakeStatsRepo{results: []*stdtype.Result{
			{Mode: "words", Amount: 25, Stats: stdtype.Stats{WPM: 87}},
		}}
		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, repo, settingsRepo)
		require.NoError(t, err)
		m, _ = update(m, tea.WindowSizeMsg{Width: 100, Height: 40})

		m, _ = update(m, navMsg{screen: screenStats})

		assert.Equal(t, screenStats, m.screen)
		assert.Contains(t, ansi.Strip(m.View().Content), "87")
	})

	t.Run("should stay on current screen when results fail to load", func(t *testing.T) {
		t.Parallel()

		repo := &fakeStatsRepo{err: errDiskFull}
		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, repo, settingsRepo)
		require.NoError(t, err)

		m, cmd := update(m, navMsg{screen: screenStats})

		assert.Equal(t, screenTyping, m.screen)
		assert.NotNil(t, cmd)
	})
}

func TestSettings(t *testing.T) {
	t.Parallel()

	initial := stdtype.Settings{Mode: "words", WordCount: 25, Duration: 30, WordList: "english"}

	tests := []struct {
		name string
		msg  settingsMsg
		want stdtype.Settings
	}{
		{
			name: "should switch to time mode when duration chosen",
			msg:  settingsMsg{field: stdtype.SettingDuration, value: 60},
			want: stdtype.Settings{Mode: "time", WordCount: 25, Duration: 60, WordList: "english"},
		},
		{
			name: "should switch to words mode when word count chosen",
			msg:  settingsMsg{field: stdtype.SettingWordCount, value: 50},
			want: stdtype.Settings{Mode: "words", WordCount: 50, Duration: 30, WordList: "english"},
		},
		{
			name: "should set mode when mode chosen",
			msg:  settingsMsg{field: stdtype.SettingMode, value: "time"},
			want: stdtype.Settings{Mode: "time", WordCount: 25, Duration: 30, WordList: "english"},
		},
		{
			name: "should set word list when language chosen",
			msg:  settingsMsg{field: stdtype.SettingList, value: "spanish"},
			want: stdtype.Settings{Mode: "words", WordCount: 25, Duration: 30, WordList: "spanish"},
		},
		{
			name: "should enable punctuation when punctuation turned on",
			msg:  settingsMsg{field: stdtype.SettingPunctuation, value: true},
			want: stdtype.Settings{Mode: "words", WordCount: 25, Duration: 30, WordList: "english", Punctuation: true},
		},
		{
			name: "should enable numbers when numbers turned on",
			msg:  settingsMsg{field: stdtype.SettingNumbers, value: true},
			want: stdtype.Settings{Mode: "words", WordCount: 25, Duration: 30, WordList: "english", Numbers: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settingsRepo := &fakeSettingsRepo{settings: initial}
			m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
			require.NoError(t, err)

			m, _ = update(m, tt.msg)

			assert.Equal(t, tt.want, *m.settings)
			assert.Equal(t, []stdtype.Settings{tt.want}, settingsRepo.saved)
		})
	}

	t.Run("should keep and not save settings when value has wrong type", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: initial}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)

		m, _ = update(m, settingsMsg{field: stdtype.SettingDuration, value: "60"})

		assert.Equal(t, initial, *m.settings)
		assert.Empty(t, settingsRepo.saved)
	})

	t.Run("should apply settings when save fails", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: initial, saveErr: errDiskFull}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)

		m, _ = update(m, settingsMsg{field: stdtype.SettingWordCount, value: 50})

		assert.Equal(t, 50, m.settings.WordCount)
	})

	t.Run("should restart on typing screen when setting changed", func(t *testing.T) {
		t.Parallel()

		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, &fakeSettingsRepo{settings: initial})
		require.NoError(t, err)
		m.screen = screenResult

		m, _ = update(m, settingsMsg{field: stdtype.SettingWordCount, value: 50})

		assert.Equal(t, screenTyping, m.screen)
	})
}

func TestView(t *testing.T) {
	t.Parallel()

	t.Run("should draw palette over screen when palette open", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)
		m, _ = update(m, tea.WindowSizeMsg{Width: 100, Height: 40})
		assert.NotContains(t, ansi.Strip(m.View().Content), "restart")

		m, _ = update(m, tea.KeyPressMsg{Code: tea.KeyEscape})

		assert.Contains(t, ansi.Strip(m.View().Content), "restart")
	})

	t.Run("should render before window size known", func(t *testing.T) {
		t.Parallel()

		settingsRepo := &fakeSettingsRepo{settings: stdtype.Settings{Mode: "words", WordCount: 2}}
		m, err := NewModel(fixedWords{"ab", "cd"}, &fakeStatsRepo{}, settingsRepo)
		require.NoError(t, err)

		assert.NotPanics(t, func() { m.View() })
	})
}

var errDiskFull = errors.New("disk full")

type fixedWords []string

func (w fixedWords) Generate(int, stdtype.GenerateOptions) ([]string, error) {
	return w, nil
}

func (w fixedWords) Lists() []string {
	return nil
}

type fakeStatsRepo struct {
	results []*stdtype.Result
	err     error
}

func (r *fakeStatsRepo) ListResults() ([]*stdtype.Result, error) {
	return r.results, r.err
}

func (r *fakeStatsRepo) CreateResult(result *stdtype.Result) error {
	r.results = append(r.results, result)

	return r.err
}

type fakeSettingsRepo struct {
	settings stdtype.Settings
	saved    []stdtype.Settings
	err      error
	saveErr  error
}

func (r *fakeSettingsRepo) GetSettings() (stdtype.Settings, error) {
	return r.settings, r.err
}

func (r *fakeSettingsRepo) UpdateSettings(settings stdtype.Settings) error {
	r.saved = append(r.saved, settings)

	return r.saveErr
}

// update unwraps tea.Model so tests can inspect the next state.
func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	next, cmd := m.Update(msg)

	return next.(Model), cmd
}

func keyPress(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// typeString returns the cmd from the last key, which carries the typing result.
func typeString(m Model, s string) (Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, r := range s {
		m, cmd = update(m, keyPress(r))
	}

	return m, cmd
}
