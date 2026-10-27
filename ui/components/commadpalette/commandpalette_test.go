package commandpalette

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	commands := []Command{
		{Label: "restart"},
		{Label: "stats", Keywords: []string{"history"}},
		{Label: "theme"},
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name: "should list all commands when query empty",
			want: []string{"restart", "stats", "theme"},
		},
		{
			name:  "should keep matching commands when query typed",
			query: "them",
			want:  []string{"theme"},
		},
		{
			name:  "should match keywords when label does not match",
			query: "history",
			want:  []string{"stats"},
		},
		{
			name:  "should list nothing when no command matches",
			query: "zzz",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewModel(commands)
			m.Focus()
			m = typeString(m, tt.query)

			assert.Equal(t, tt.want, labels(m.Results()))
		})
	}
}

func TestSelect(t *testing.T) {
	t.Parallel()

	t.Run("should send CommandMsg when enter pressed on command without subcommands", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{{Label: "restart"}, {Label: "stats"}})
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		require.NotNil(t, cmd)
		assert.Equal(t, CommandMsg{Command: Command{Label: "stats"}}, cmd())
	})

	t.Run("should keep cursor on last result when moving down past it", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{{Label: "restart"}, {Label: "stats"}})
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		require.NotNil(t, cmd)
		assert.Equal(t, CommandMsg{Command: Command{Label: "stats"}}, cmd())
	})

	t.Run("should show subcommands when enter pressed on command with subcommands", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{
			{Label: "mode", Subcommands: []Command{{Label: "time"}, {Label: "words"}}},
		})

		m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Nil(t, cmd)
		assert.Equal(t, []string{"time", "words"}, labels(m.Results()))
	})

	t.Run("should search selected subcommands when command picked from a filtered list", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{
			{Label: "lines", Subcommands: []Command{{Label: "2"}, {Label: "3"}}},
			{Label: "mode", Subcommands: []Command{{Label: "time"}, {Label: "words"}}},
		})
		m.Focus()
		m = typeString(m, "mode")
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		m = typeString(m, "wor")

		assert.Equal(t, []string{"words"}, labels(m.Results()))
	})

	t.Run("should do nothing when enter pressed with no matches", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{{Label: "restart"}})
		m.Focus()
		m = typeString(m, "zzz")

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Nil(t, cmd)
	})
}

func TestReset(t *testing.T) {
	t.Parallel()

	m := NewModel([]Command{
		{Label: "restart"},
		{Label: "mode", Subcommands: []Command{{Label: "time"}}},
	})
	m.Focus()
	m = typeString(m, "mode")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m.Reset()

	assert.Empty(t, m.Query())
	assert.Equal(t, []string{"restart", "mode"}, labels(m.Results()))
}

func TestPaletteView(t *testing.T) {
	t.Parallel()

	t.Run("should show empty notice when no command matches", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{{Label: "restart"}})
		m.SetSize(60, 20)
		m.Focus()
		m = typeString(m, "zzz")

		assert.Contains(t, ansi.Strip(m.View()), "no matching commands")
	})

	t.Run("should mark highlighted command when cursor on it", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{{Label: "restart"}, {Label: "stats"}})
		m.SetSize(60, 20)
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		assert.Contains(t, ansi.Strip(m.View()), "▶ stats")
	})
}

func TestPaletteLayout(t *testing.T) {
	t.Parallel()

	t.Run("should show full label when it exactly fills the inner width", func(t *testing.T) {
		t.Parallel()

		// minWidth 20 minus the default frame of 6 leaves 14 columns:
		// 2 for the "▶ " prefix and 12 for the label.
		m := NewModel([]Command{{Label: "twelve-chars"}})
		m.SetSize(0, 0)

		assert.Contains(t, ansi.Strip(m.View()), "▶ twelve-chars")
	})

	t.Run("should keep same height when no command matches", func(t *testing.T) {
		t.Parallel()

		m := NewModel([]Command{{Label: "restart"}})
		m.SetSize(60, 20)
		m.Focus()
		withResult := lipgloss.Height(m.View())

		m = typeString(m, "zzz")

		assert.Equal(t, withResult, lipgloss.Height(m.View()))
	})
}

func typeString(m Model, s string) Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	return m
}

func labels(commands []Command) []string {
	out := make([]string, 0, len(commands))
	for _, c := range commands {
		out = append(out, c.Label)
	}

	return out
}
