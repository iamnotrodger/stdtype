package typing

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinish(t *testing.T) {
	t.Parallel()

	t.Run("should send ResultMsg when last word typed in words mode", func(t *testing.T) {
		t.Parallel()

		settings := stdtype.Settings{Mode: "words", WordCount: 2}
		m := NewModel(fixedWords{"ab", "cd"}, &settings)

		_, cmd := typeString(m, "ab cd")

		msg := runCmd(cmd)
		require.IsType(t, ResultMsg{}, msg)

		result := msg.(ResultMsg).Result
		assert.Equal(t, "words", result.Mode)
		assert.Equal(t, 2, result.Amount)
	})

	t.Run("should not finish again when keys arrive after test finished", func(t *testing.T) {
		t.Parallel()

		settings := stdtype.Settings{Mode: "words", WordCount: 2}
		m := NewModel(fixedWords{"ab", "cd"}, &settings)
		m, _ = typeString(m, "ab cd")

		_, cmd := typeString(m, "x")

		assert.Nil(t, cmd)
	})

	t.Run("should send ResultMsg on tick when duration elapsed in time mode", func(t *testing.T) {
		t.Parallel()

		// Zero duration makes the first tick after starting the deadline,
		// since the test clock can't be injected from this package.
		settings := stdtype.Settings{Mode: "time", Duration: 0}
		m := NewModel(fixedWords{"ab", "cd"}, &settings)
		m, _ = typeString(m, "a")

		_, cmd := m.Update(TickMsg{ID: m.tickID})

		assert.IsType(t, ResultMsg{}, runCmd(cmd))
	})
}

// Tick cmds are only checked for nil, never run: tea.Tick blocks for its
// full interval.
func TestTick(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modelFunc   func(Model) (Model, TickMsg)
		wantTicking bool
		wantSamples int
	}{
		{
			name: "should not tick when test not started",
			modelFunc: func(m Model) (Model, TickMsg) {
				return m, TickMsg{ID: m.tickID}
			},
			wantTicking: false,
			wantSamples: 0,
		},
		{
			name: "should record a sample when tick arrives while typing",
			modelFunc: func(m Model) (Model, TickMsg) {
				m, _ = typeString(m, "ab")

				return m, TickMsg{ID: m.tickID}
			},
			wantTicking: true,
			wantSamples: 1,
		},
		{
			name: "should drop tick when it is from a stale chain",
			modelFunc: func(m Model) (Model, TickMsg) {
				stale := TickMsg{ID: m.tickID}
				m, _ = m.Restart()
				m, _ = typeString(m, "ab")

				return m, stale
			},
			wantTicking: false,
			wantSamples: 0,
		},
		{
			name: "should stop ticking when test completed",
			modelFunc: func(m Model) (Model, TickMsg) {
				m, _ = typeString(m, "ab cd")

				return m, TickMsg{ID: m.tickID}
			},
			wantTicking: false,
			wantSamples: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := stdtype.Settings{Mode: "words", WordCount: 2}
			m := NewModel(fixedWords{"ab", "cd"}, &settings)

			m, msg := tt.modelFunc(m)

			m, cmd := m.Update(msg)

			assert.Equal(t, tt.wantTicking, cmd != nil, "ticking")
			assert.Len(t, m.samples.WPM, tt.wantSamples)
		})
	}
}

func TestTickChainStart(t *testing.T) {
	t.Parallel()

	t.Run("should start ticking when first key typed", func(t *testing.T) {
		t.Parallel()

		settings := stdtype.Settings{Mode: "words", WordCount: 2}
		m := NewModel(fixedWords{"ab", "cd"}, &settings)

		_, cmd := typeString(m, "a")

		assert.NotNil(t, cmd)
	})

	t.Run("should not start another chain when later keys typed", func(t *testing.T) {
		t.Parallel()

		settings := stdtype.Settings{Mode: "words", WordCount: 2}
		m := NewModel(fixedWords{"ab", "cd"}, &settings)
		m, _ = typeString(m, "a")

		_, cmd := typeString(m, "b")

		assert.Nil(t, cmd)
	})
}

func TestModelView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		settings  stdtype.Settings
		typed     string
		wantShown []string
	}{
		{
			name:      "should show word progress when in words mode",
			settings:  stdtype.Settings{Mode: "words", WordCount: 2},
			typed:     "ab c",
			wantShown: []string{"1/2", "ab cd", "time words | 10 25 50 100"},
		},
		{
			name:      "should show remaining seconds when in time mode",
			settings:  stdtype.Settings{Mode: "time", Duration: 30},
			wantShown: []string{"30", "time words | 15 30 60 120"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewModel(fixedWords{"ab", "cd"}, &tt.settings)
			m.SetSize(80, 24)

			m, _ = typeString(m, tt.typed)
			view := ansi.Strip(m.View())

			for _, want := range tt.wantShown {
				assert.Contains(t, view, want)
			}
		})
	}
}

type fixedWords []string

func (w fixedWords) Generate(int, stdtype.GenerateOptions) ([]string, error) {
	return w, nil
}

func (w fixedWords) Lists() []string {
	return nil
}

// typeString returns the cmd from the last key, which carries ResultMsg.
func typeString(m Model, s string) (Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, r := range s {
		m, cmd = m.Update(keyPress(r))
	}

	return m, cmd
}

func runCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}

	return cmd()
}
