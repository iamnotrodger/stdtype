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
	appctx "github.com/iamnotrodger/stdytype/ui/context"
	"github.com/iamnotrodger/stdytype/ui/keys"
)

// TODO: send msg on select.
type SelectedMsg struct {
	Result *stdtype.Result
}

type Model struct {
	appCtx  *appctx.Context
	table   table.Model
	results []*stdtype.Result
	summary summary

	width, height int
}

func NewModel(appCtx *appctx.Context) Model {
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

	theme := appCtx.Theme

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithWidth(tableWidth),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		MarginBottom(1).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(theme.Primary).
		Background(theme.Secondary)
	t.SetStyles(s)

	return Model{
		appCtx: appCtx,
		table:  t,
	}
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

func (m Model) Update(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	if key.Matches(msg, keys.Keys.Enter) {
		// TODO: should we add guard rails?
		result := m.results[m.table.Cursor()]

		return m, func() tea.Msg {
			return SelectedMsg{result}
		}
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	main := lipgloss.JoinVertical(
		lipgloss.Left,
		m.summaryView(),
		"\n\n",
		m.table.View(),
	)

	return common.Layout{
		Style:  m.appCtx.ColorStyle,
		Header: "stats",
		Main:   main,
		Width:  m.width,
		Height: m.height,
	}.View()
}

func (m Model) summaryView() string {
	style := m.appCtx.ColorStyle
	sum := m.summary

	stat := func(label, value string) string {
		return lipgloss.JoinVertical(lipgloss.Left,
			style.Success.Render(label),
			style.Primary.Bold(true).Render(value),
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
