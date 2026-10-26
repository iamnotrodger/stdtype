package stats

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/common"
)

// ResultMsg is sent when a result is selected.
type ResultMsg struct {
	Result *stdtype.Result
}

type KeyMap struct {
	Select key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view result"),
		),
	}
}

type Styles struct {
	Title  lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Footer lipgloss.Style
	Table  table.Styles
}

func DefaultStyles() Styles {
	t := table.DefaultStyles()
	t.Header = t.Header.MarginBottom(1).Bold(true)
	t.Selected = t.Selected.Background(lipgloss.Blue)

	return Styles{
		Title:  lipgloss.NewStyle(),
		Label:  lipgloss.NewStyle().Foreground(lipgloss.Green),
		Value:  lipgloss.NewStyle().Bold(true),
		Footer: lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Table:  t,
	}
}

type Model struct {
	KeyMap KeyMap

	styles  Styles
	table   table.Model
	results []*stdtype.Result
	summary summary

	width, height int
}

func NewModel() Model {
	columns := []table.Column{
		{Title: "wpm", Width: 5},
		{Title: "raw", Width: 5},
		{Title: "accuracy", Width: 10},
		{Title: "consistency", Width: 13},
		{Title: "mode", Width: 10},
		{Title: "date", Width: 13},
	}

	tableWidth := 0
	for _, c := range columns {
		tableWidth += c.Width + 2
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithWidth(tableWidth),
		table.WithHeight(15),
	)

	m := Model{
		KeyMap: DefaultKeyMap(),
		table:  t,
	}
	m.SetStyle(DefaultStyles())

	return m
}

func (m *Model) SetStyle(s Styles) {
	m.styles = s
	m.table.SetStyles(s.Table)
}

func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

func (m *Model) SetResults(results []*stdtype.Result) {
	m.results = results
	m.summary = summarize(results)

	rows := make([]table.Row, 0, len(results))

	for _, stat := range results {
		rows = append(rows, table.Row{
			fmt.Sprintf("%.0f", stat.Stats.WPM),
			fmt.Sprintf("%.0f", stat.Stats.RawWPM),
			fmt.Sprintf("%.0f%%", stat.Stats.Accuracy),
			fmt.Sprintf("%.0f%%", stat.Consistency),
			fmt.Sprintf("%s %d", stat.Mode, stat.Amount),
			stat.Timestamp.Format("02 Jan 2006"),
		})
	}

	m.table.SetRows(rows)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.KeyMap.Select) {
		return m, m.selectResult()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m Model) selectResult() tea.Cmd {
	i := m.table.Cursor()
	if i < 0 || i >= len(m.results) {
		return nil
	}

	result := m.results[i]

	return func() tea.Msg {
		return ResultMsg{Result: result}
	}
}

func (m Model) View() string {
	main := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderSummary(),
		"\n\n",
		m.table.View(),
	)

	return common.Layout{
		Header: m.styles.Title.Render("stats"),
		Footer: m.styles.Footer.Render(common.DefaultFooter()),
		Main:   main,
		Width:  m.width,
		Height: m.height,
	}.View()
}

func (m Model) renderSummary() string {
	sum := m.summary

	stat := func(label, value string) string {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.styles.Label.Render(label),
			m.styles.Value.Render(value),
		)
	}

	gap := strings.Repeat(" ", 3)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		stat("completed", strconv.Itoa(sum.tests)), gap,
		stat("time typing", clock(sum.timeTyping)), gap,
		stat("best wpm", fmt.Sprintf("%.0f", sum.bestWPM)), gap,
		stat("avg wpm", fmt.Sprintf("%.0f", sum.avgWPM)),
	)
}

func clock(d time.Duration) string {
	secs := int(d.Round(time.Second).Seconds())

	return fmt.Sprintf("%02d:%02d:%02d", secs/3600, secs/60%60, secs%60)
}

type summary struct {
	tests      int
	timeTyping time.Duration
	bestWPM    float64
	avgWPM     float64
}

// summarize skips invalid results, as they don't reflect real typing.
func summarize(results []*stdtype.Result) summary {
	var s summary
	var sumWPM float64

	for _, r := range results {
		if r.Invalid {
			continue
		}

		s.tests++
		s.timeTyping += r.Duration
		s.bestWPM = max(s.bestWPM, r.Stats.WPM)
		sumWPM += r.Stats.WPM
	}

	if s.tests > 0 {
		s.avgWPM = sumWPM / float64(s.tests)
	}

	return s
}
