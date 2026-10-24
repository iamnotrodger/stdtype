package stats

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
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
	// TODO: add average type shiiiii.

	return common.Layout{
		Style:  m.appCtx.ColorStyle,
		Header: "stats",
		Main:   m.table.View(),
		Width:  m.width,
		Height: m.height,
	}.View()
}
