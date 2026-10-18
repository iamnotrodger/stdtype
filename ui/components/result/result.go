package result

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/common"
	appctx "github.com/iamnotrodger/stdytype/ui/context"
	"github.com/iamnotrodger/stdytype/ui/theme"
)

type Model struct {
	appCtx        *appctx.Context
	result        *stdtype.Result
	style         theme.ColorStyle
	width, height int
}

func NewMode(appCtx *appctx.Context) Model {
	return Model{appCtx: appCtx, style: appCtx.ColorStyle}
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

	if m.result.Invalid {
		invalid := m.result.InvalidReason + " -- result not saved"
		b.WriteString(m.style.Error.Render(invalid))
		b.WriteString("\n\n")
	}

	b.WriteString(m.statsView())

	return common.Layout{
		Style:  m.style,
		Header: m.style.Primary.Bold(true).Render("results"),
		Main:   b.String(),
		Width:  m.width,
		Height: m.height,
	}.View()
}

func (m Model) statsView() string {
	result := m.result
	stats := m.result.Stats

	stat := func(label, value string) string {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.style.Primary.Bold(true).Render(label),
			m.style.Faint.Render(value),
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
