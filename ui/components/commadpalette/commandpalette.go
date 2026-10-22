package commandpalette

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	appctx "github.com/iamnotrodger/stdytype/ui/context"
	"github.com/iamnotrodger/stdytype/ui/theme"
	"github.com/sahilm/fuzzy"
)

// Command represents a single executable command in the palette.
type Command struct {
	// Label is the displayed name and the primary search target.
	Label string
	// Keywords are additional searchable terms that are not displayed.
	Keywords    []string
	Run         tea.Cmd
	Subcommands []Command
}

type CommandMsg struct {
	Command Command
}

type Model struct {
	appCtx   appctx.Context
	commands []Command
	filtered []Command
	cursor   int
	index    int
	input    textinput.Model

	style      theme.ColorStyle
	maxResults int
	maxWidth   int

	width  int
	height int
}

func NewModel(appCtx appctx.Context, commands []Command, opts ...Option) Model {
	options := newOptions(opts...)

	input := textinput.New()
	input.Placeholder = options.placeholder
	input.Prompt = promptStr

	return Model{
		appCtx:     appCtx,
		commands:   commands,
		filtered:   commands,
		index:      -1,
		style:      appCtx.ColorStyle,
		maxResults: options.maxResults,
		maxWidth:   options.maxWidth,
		input:      input,
	}
}

func (m *Model) Focus() tea.Cmd {
	return m.input.Focus()
}

func (m *Model) Reset() {
	m.input.Reset()
	m.index = -1
	m.filtered = m.commands
	m.cursor = 0
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	w := min(m.maxWidth, width)
	if textW := w - 6; textW > 0 {
		m.input.SetWidth(textW)
	}
}

func (m Model) Update(msg tea.KeyMsg) (Model, tea.Cmd) {
	visibleCount := min(len(m.filtered), m.maxResults)

	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}

		return m, nil

	case "down":
		if m.cursor < visibleCount-1 {
			m.cursor++
		}

		return m, nil

	case "enter":
		if visibleCount == 0 {
			return m, nil
		}

		action := m.filtered[m.cursor]

		if len(action.Subcommands) > 0 {
			m.filtered = action.Subcommands
			m.index = m.cursor
			m.cursor = 0
			m.input.Reset()

			return m, nil
		}

		return m, func() tea.Msg {
			return CommandMsg{Command: action}
		}
	}

	prevQuery := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if q := m.input.Value(); q != prevQuery {
		m.applyFilter(q)
		m.cursor = 0
	}

	return m, cmd
}

func (m Model) View() string {
	width := max(min(m.maxWidth, m.width), 20)
	innerWidth := width - 6

	colorStyle := m.appCtx.ColorStyle
	theme := m.appCtx.Theme

	var b strings.Builder

	b.WriteString(m.input.View())
	b.WriteString("\n")

	divider := colorStyle.Faint.Render(strings.Repeat("─", innerWidth))
	b.WriteString(divider)
	b.WriteString("\n")

	results := m.filtered
	if len(results) > m.maxResults {
		results = results[:m.maxResults]
	}

	if len(results) == 0 {
		b.WriteString(colorStyle.Faint.Render("  no matching commands"))
		b.WriteString("\n")
	} else {
		for i, command := range results {
			prefix := "  "
			if i == m.cursor {
				prefix = "▶ "
			}

			label := truncate(command.Label, innerWidth-len(prefix))
			b.WriteString(colorStyle.Primary.Render(prefix + label))

			if i < len(results)-1 {
				b.WriteString("\n")
			}
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Faint).
		Padding(1, 2).
		Width(width).
		Render(b.String())
}

func (m Model) Query() string {
	return m.input.Value()
}

func (m Model) Results() []Command {
	if len(m.filtered) > m.maxResults {
		return m.filtered[:m.maxResults]
	}

	return m.filtered
}

func (m *Model) applyFilter(query string) {
	commands := m.commands
	if m.index >= 0 {
		commands = m.commands[m.index].Subcommands
	}

	if strings.TrimSpace(query) == "" {
		m.filtered = commands

		return
	}

	sources := make([]string, len(commands))
	for i, a := range commands {
		sources[i] = a.Label
		if len(a.Keywords) > 0 {
			sources[i] += " " + strings.Join(a.Keywords, " ")
		}
	}

	matches := fuzzy.Find(query, sources)
	m.filtered = make([]Command, len(matches))
	for i, match := range matches {
		m.filtered[i] = commands[match.Index]
	}
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	if n <= 1 {
		return "…"
	}

	return string(runes[:n-1]) + "…"
}
