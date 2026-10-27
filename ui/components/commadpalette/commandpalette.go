package commandpalette

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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

const minWidth = 20

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "move down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}

type Styles struct {
	Box     lipgloss.Style
	Divider lipgloss.Style
	Empty   lipgloss.Style
	Item    lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.BrightBlack).
			Padding(1, 2),
		Divider: lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Empty:   lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Item:    lipgloss.NewStyle(),
	}
}

type Model struct {
	KeyMap KeyMap

	commands []Command
	scope    []Command
	cursor   int
	input    textinput.Model

	styles     Styles
	maxResults int
	maxWidth   int

	width int
}

func NewModel(commands []Command, opts ...Option) Model {
	options := newOptions(opts...)

	input := textinput.New()
	input.Placeholder = options.placeholder
	input.Prompt = promptStr

	return Model{
		KeyMap:     DefaultKeyMap(),
		commands:   commands,
		scope:      commands,
		styles:     DefaultStyles(),
		maxResults: options.maxResults,
		maxWidth:   options.maxWidth,
		input:      input,
	}
}

func (m *Model) SetStyle(s Styles) {
	m.styles = s
	m.input.SetWidth(m.innerWidth())
}

func (m *Model) Focus() tea.Cmd {
	return m.input.Focus()
}

func (m *Model) Reset() {
	m.input.Reset()
	m.scope = m.commands
	m.cursor = 0
}

func (m *Model) SetSize(width, _ int) {
	m.width = width
	m.input.SetWidth(m.innerWidth())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			m.cursor = max(m.cursor-1, 0)

			return m, nil
		case key.Matches(msg, m.KeyMap.Down):
			m.cursor = max(min(m.cursor+1, len(m.Results())-1), 0)

			return m, nil
		case key.Matches(msg, m.KeyMap.Select):
			return m.selectCommand()
		}
	}

	prevQuery := m.input.Value()

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if m.input.Value() != prevQuery {
		m.cursor = 0
	}

	return m, cmd
}

// selectCommand opens a command's subcommands, or runs it when it has none.
func (m Model) selectCommand() (Model, tea.Cmd) {
	results := m.Results()
	if len(results) == 0 {
		return m, nil
	}

	command := results[m.cursor]

	if len(command.Subcommands) > 0 {
		m.scope = command.Subcommands
		m.cursor = 0
		m.input.Reset()

		return m, nil
	}

	return m, func() tea.Msg {
		return CommandMsg{Command: command}
	}
}

func (m Model) View() string {
	divider := m.styles.Divider.Render(strings.Repeat("─", m.innerWidth()))
	lines := append([]string{m.input.View(), divider}, m.renderItems()...)

	return m.styles.Box.Width(m.boxWidth()).Render(strings.Join(lines, "\n"))
}

func (m Model) renderItems() []string {
	results := m.Results()
	if len(results) == 0 {
		return []string{m.styles.Empty.Render("  no matching commands")}
	}

	items := make([]string, 0, len(results))
	for i, command := range results {
		prefix := "  "
		if i == m.cursor {
			prefix = "▶ "
		}

		label := ansi.Truncate(command.Label, m.innerWidth()-lipgloss.Width(prefix), "…")
		items = append(items, m.styles.Item.Render(prefix+label))
	}

	return items
}

func (m Model) boxWidth() int {
	return max(min(m.maxWidth, m.width), minWidth)
}

// innerWidth is the space inside the box's border and padding.
func (m Model) innerWidth() int {
	return max(m.boxWidth()-m.styles.Box.GetHorizontalFrameSize(), 1)
}

func (m Model) Query() string {
	return m.input.Value()
}

func (m Model) Results() []Command {
	results := m.scope

	if query := m.input.Value(); strings.TrimSpace(query) != "" {
		targets := make([]string, len(m.scope))
		for i, command := range m.scope {
			targets[i] = strings.Join(append([]string{command.Label}, command.Keywords...), " ")
		}

		matches := fuzzy.Find(query, targets)
		results = make([]Command, len(matches))
		for i, match := range matches {
			results[i] = m.scope[match.Index]
		}
	}

	return results[:min(len(results), m.maxResults)]
}
