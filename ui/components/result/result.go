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

	width, height int
}

func NewMode(appCtx *appctx.Context) Model {
	return Model{appCtx: appCtx}
}

func (m *Model) SetResult(result *stdtype.Result) {
	m.result = result
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) View() string {
	var b strings.Builder

	style := m.appCtx.ColorStyle

	if m.result.Invalid {
		invalid := m.result.InvalidReason + " -- result not saved"
		b.WriteString(style.Error.Render(invalid))
		b.WriteString("\n\n")
	}

	b.WriteString(m.statsView())

	return common.Layout{
		Style:  style,
		Header: style.Primary.Bold(true).Render("results"),
		Main:   b.String(),
		Width:  m.width,
		Height: m.height,
	}.View()
}

func (m Model) statsView() string {
	result := m.result
	stats := m.result.Stats
	style := m.appCtx.ColorStyle

	stat := func(label, value string) string {
		return lipgloss.JoinVertical(lipgloss.Left,
			style.Primary.Bold(true).Render(label),
			style.Faint.Render(value),
		)
	}

	gap := strings.Repeat(" ", 3)

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
		stat("consistency", fmt.Sprintf("%.0f%%", result.Consistency)), gap,
		stat("time", fmt.Sprintf("%.1fs", result.Duration.Seconds())), gap,
	)

	return lipgloss.JoinVertical(lipgloss.Left, headline, "", details)
}
