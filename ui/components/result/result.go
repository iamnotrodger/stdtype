package result

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/iamnotrodger/stdtype/ui/common"
)

const graphHeight = 10

type Styles struct {
	Title   lipgloss.Style
	Invalid lipgloss.Style
	Label   lipgloss.Style
	Value   lipgloss.Style
	Footer  lipgloss.Style
	Graph   GraphStyles
}

func DefaultStyles() Styles {
	return Styles{
		Title:   lipgloss.NewStyle().Bold(true),
		Invalid: lipgloss.NewStyle().Foreground(lipgloss.Red),
		Label:   lipgloss.NewStyle().Bold(true),
		Value:   lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Footer:  lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Graph:   DefaultGraphStyles(),
	}
}

type Model struct {
	result *stdtype.Result
	styles Styles
	graph  Graph

	width, height int
}

func NewModel() Model {
	return Model{
		styles: DefaultStyles(),
		graph:  NewGraph(),
	}
}

func (m *Model) SetResult(result *stdtype.Result) {
	m.result = result
	m.graph.SetSamples(result.Samples)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.graph.SetSize(graphWidth(width), graphHeight)
}

func (m *Model) SetStyle(s Styles) {
	m.styles = s
	m.graph.SetStyle(s.Graph)
}

func (m Model) View() string {
	if m.result == nil {
		return ""
	}

	var b strings.Builder

	if m.result.Invalid {
		invalid := m.result.InvalidReason + " -- result not saved"
		b.WriteString(m.styles.Invalid.Render(invalid))
		b.WriteString("\n\n")
	}

	b.WriteString(m.graph.View())
	b.WriteString("\n\n")
	b.WriteString(m.renderStats())

	return common.Layout{
		Header: m.styles.Title.Render("results"),
		Main:   b.String(),
		Footer: m.styles.Footer.Render(common.DefaultFooter()),
		Width:  m.width,
		Height: m.height,
	}.View()
}

func (m Model) renderStats() string {
	result := m.result
	stats := m.result.Stats

	stat := func(label, value string) string {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.styles.Label.Render(label),
			m.styles.Value.Render(value),
		)
	}

	gap := strings.Repeat(" ", 3)

	headline := lipgloss.JoinHorizontal(lipgloss.Top,
		stat("wpm", fmt.Sprintf("%.0f", stats.WPM)), gap,
		stat("accuracy", fmt.Sprintf("%.0f%%", stats.Accuracy)),
	)

	testType := fmt.Sprintf("%s %d", result.Mode, result.Amount)
	chars := fmt.Sprintf(
		"%d/%d/%d/%d",
		stats.CorrectChars, stats.IncorrectChars, stats.ExtraChars, stats.MissedChars,
	)

	details := lipgloss.JoinHorizontal(lipgloss.Top,
		stat("test type", testType), gap,
		stat("raw", fmt.Sprintf("%.0f", stats.RawWPM)), gap,
		stat("characters", chars), gap,
		stat("consistency", fmt.Sprintf("%.0f%%", result.Consistency)), gap,
		stat("time", fmt.Sprintf("%.1fs", result.Duration.Seconds())), gap,
	)

	return lipgloss.JoinVertical(lipgloss.Left, headline, "", details)
}

func graphWidth(width int) int {
	return min(72, max(width-8, 20))
}
