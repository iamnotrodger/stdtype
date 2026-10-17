package result

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/common"
	appctx "github.com/iamnotrodger/stdytype/ui/context"
)

type Model struct {
	appCtx *appctx.Context
	result *stdtype.Result
}

func NewMode(appCtx *appctx.Context) Model {
	return Model{appCtx: appCtx}
}

func (m Model) SetResult(result *stdtype.Result) Model {
	m.result = result

	return m
}

func (m Model) View() string {
	theme := m.appCtx.Theme
	width, height := m.appCtx.Dimension()

	var b strings.Builder

	if m.result.Invalid {
		invalid := m.result.InvalidReason + " -- result not saved"
		b.WriteString(theme.Incorrect.Render(invalid))
		b.WriteString("\n\n")
	}

	b.WriteString(m.statsView())

	return common.Layout{
		Theme:  theme,
		Header: theme.Title.Render("results"),
		Main:   b.String(),
		Width:  width,
		Height: height,
	}.View()
}

func (m Model) statsView() string {
	theme := m.appCtx.Theme
	result := m.result
	stats := m.result.Stats

	stat := func(label, value string) string {
		return lipgloss.JoinVertical(lipgloss.Left,
			theme.StatLabel.Render(label),
			theme.Stat.Render(value),
		)
	}

	gap := "   "

	headline := lipgloss.JoinHorizontal(lipgloss.Top,
		stat("wpm", fmt.Sprintf("%.0f", stats.WPM)), gap,
		stat("accuracy", fmt.Sprintf("%.0f", stats.Accuracy)),
	)

	testType := fmt.Sprintf("%s %d", result.Mode, result.Amount)
	chars := fmt.Sprintf("%d/%d/%d/", stats.CorrectChars, stats.IncorrectChars, stats.MissedChars)

	details := lipgloss.JoinHorizontal(lipgloss.Top,
		stat("test type", testType), gap,
		stat("raw", fmt.Sprintf("%.0f", stats.RawWPM)), gap,
		stat("characters", chars), gap,
		stat("consistency", fmt.Sprintf("%.0f", result.Consistency)), gap,
		stat("time", fmt.Sprintf("%.1fs", result.Duration.Seconds())), gap,
	)

	return lipgloss.JoinVertical(lipgloss.Left, headline, "", details)
}
