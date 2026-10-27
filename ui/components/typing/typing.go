package typing

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	"github.com/iamnotrodger/stdytype/ui/common"
)

type TickMsg struct {
	ID   int
	Time time.Time
}

type ResultMsg struct {
	Result *stdtype.Result
}

func tick(id int, elapsed time.Duration) tea.Cmd {
	next := elapsed.Truncate(time.Second) + time.Second

	return tea.Tick(next-elapsed, func(t time.Time) tea.Msg {
		return TickMsg{ID: id, Time: t}
	})
}

type Styles struct {
	ConfigActive    lipgloss.Style
	ConfigInactive  lipgloss.Style
	ConfigSeparator lipgloss.Style
	Footer          lipgloss.Style
	TextArea        TextAreaStyles
}

func DefaultStyles() Styles {
	return Styles{
		ConfigActive:    lipgloss.NewStyle().Bold(true),
		ConfigInactive:  lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		ConfigSeparator: lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Footer:          lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		TextArea:        DefaultTextAreaStyles(),
	}
}

type Model struct {
	styles     Styles
	wordSource stdtype.WordSource
	settings   *stdtype.Settings

	mode    stdtype.Mode
	test    *stdtype.TypeTest
	samples stdtype.Samples

	prevStats   stdtype.Stats
	prevElapsed time.Duration
	afkSeconds  int
	tickID      int

	textarea TextArea
	fade     animation.Fade

	width, height int
}

func NewModel(
	wordSource stdtype.WordSource,
	settings *stdtype.Settings,
) Model {
	m := Model{
		wordSource: wordSource,
		settings:   settings,
		styles:     DefaultStyles(),
		textarea:   NewTextArea(),
		fade:       animation.NewFade(),
	}

	return m.reset()
}

func (m Model) Init() tea.Cmd {
	return cursor.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		return m.onTick(msg)
	case animation.Msg:
		return m.onAnimate()
	case tea.KeyPressMsg:
		return m.onKey(msg)
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	m.textarea.SetFade(m.fade.Value)
	m.textarea.SetSize(wrapWidth(m.width), clampLines(m.settings.Lines))

	main := lipgloss.JoinVertical(
		lipgloss.Center,
		m.renderProgress(),
		"",
		m.textarea.View(),
	)

	return common.Layout{
		Footer: m.styles.Footer.Render(common.DefaultFooter()),
		Header: m.renderConfigBar(),
		Main:   main,
		Width:  m.width,
		Height: m.height,
	}.View()
}

func (m Model) renderConfigBar() string {
	opt := func(label string, on bool) string {
		if on {
			return m.styles.ConfigActive.Render(label)
		}

		return m.styles.ConfigInactive.Render(label)
	}

	isTimeMode := m.test.Mode == stdtype.TimeMode
	modes := opt("time", isTimeMode) + " " + opt("words", !isTimeMode)

	selected := m.settings.WordCount
	options := common.WordOptions
	if isTimeMode {
		selected = m.settings.Duration
		options = common.DurationOptions
	}

	labels := make([]string, 0, len(options))
	for _, d := range options {
		label := strconv.Itoa(d)
		on := d == selected
		labels = append(labels, opt(label, on))
	}

	sep := m.styles.ConfigSeparator.Render(" | ")

	return modes + sep + strings.Join(labels, " ")
}

func (m Model) renderProgress() string {
	if m.test.Mode == stdtype.TimeMode {
		remaining := m.settings.Duration - int(m.test.Elapsed().Seconds())

		return animation.FadeRender(m.fade.Value, strconv.Itoa(max(remaining, 0)))
	}

	count := min(m.test.ActiveWord(), m.settings.WordCount)
	progress := fmt.Sprintf("%d/%d", count, m.settings.WordCount)

	return animation.FadeRender(m.fade.Value, progress)
}

func (m *Model) SetStyle(s Styles) {
	m.styles = s
	m.textarea.SetStyle(s.TextArea)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) reset() Model {
	mode := stdtype.ParseMode(m.settings.Mode)

	count := m.settings.WordCount
	if mode == stdtype.TimeMode {
		count = m.settings.Duration*3 + 30
	}

	// TODO: log error
	words, _ := m.wordSource.Generate(
		count,
		stdtype.GenerateOptions{
			List:        m.settings.WordList,
			Punctuation: m.settings.Punctuation,
			Numbers:     m.settings.Numbers,
		},
	)

	m.mode = stdtype.ParseMode(m.settings.Mode)
	m.test = stdtype.NewTypeTest(words, m.mode)
	m.samples = stdtype.Samples{}
	m.prevStats = stdtype.Stats{}
	m.prevElapsed = 0
	m.afkSeconds = 0
	m.tickID++

	m.textarea.SetTest(m.test)

	return m
}

func (m Model) Restart() (Model, tea.Cmd) {
	m = m.reset()
	m.fade.Start()

	return m, tea.Batch(animation.Tick(), cursor.Blink)
}

func (m Model) onTick(msg TickMsg) (Model, tea.Cmd) {
	if msg.ID != m.tickID || !m.test.Started() || m.test.Completed() {
		return m, nil
	}

	if m.mode == stdtype.TimeMode {
		duration := time.Duration(m.settings.Duration) * time.Second
		if m.test.Elapsed() >= duration {
			return m.onFinish()
		}
	}

	prevKeystrokes := m.prevStats.Keystrokes
	m = m.recordSample()

	if m.prevStats.Keystrokes == prevKeystrokes {
		m.afkSeconds++
	}

	return m, tick(m.tickID, m.test.Elapsed())
}

func (m Model) recordSample() Model {
	elapsed := m.test.Elapsed()
	secs := (elapsed - m.prevElapsed).Seconds()
	stats := stdtype.ComputeStats(m.test)

	rawDelta := max(0, stats.TypedChars-m.prevStats.TypedChars)
	netDelta := max(0, stats.CorrectChars-m.prevStats.CorrectChars)

	errKeys := stats.Keystrokes - stats.CorrectKeystrokes
	prevErrKeys := m.prevStats.Keystrokes - m.prevStats.CorrectKeystrokes
	errDelta := max(0, errKeys-prevErrKeys)

	m.samples.Raw = append(m.samples.Raw, wpm(rawDelta, secs))
	m.samples.WPM = append(m.samples.WPM, wpm(netDelta, secs))
	m.samples.Err = append(m.samples.Err, errDelta)

	m.prevStats = stats
	m.prevElapsed = elapsed

	return m
}

// wpm uses the standard 5 characters per word.
func wpm(chars int, secs float64) float64 {
	return float64(chars) / 5 / (secs / 60)
}

func (m Model) onKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	wasStarted := m.test.Started()

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	if !m.test.Finished() && m.test.Completed() {
		return m.onFinish()
	}

	if !wasStarted && m.test.Started() {
		cmd = tea.Batch(cmd, tick(m.tickID, m.test.Elapsed()))
	}

	return m, cmd
}

func (m Model) onFinish() (Model, tea.Cmd) {
	m.test.Finish()

	// A sliver of a second would turn one keystroke into a WPM spike.
	if m.test.Elapsed()-m.prevElapsed >= time.Second/2 {
		m = m.recordSample()
	}

	stats := stdtype.ComputeStats(m.test)
	consistency := stdtype.Consistency(m.samples.Raw)

	mode := stdtype.ParseMode(m.settings.Mode)
	amount := m.settings.Duration
	if mode == stdtype.WordsMode {
		amount = m.settings.WordCount
	}

	result := &stdtype.Result{
		Timestamp:   time.Now(),
		Mode:        m.mode.String(),
		Amount:      amount,
		Stats:       stats,
		Consistency: consistency,
		Samples:     m.samples,
		Duration:    m.test.Elapsed(),
	}

	if stdtype.IsAFK(m.afkSeconds, int(m.test.Elapsed().Seconds())) {
		result.Invalid = true
		result.InvalidReason = "afk detected"
	}

	return m, func() tea.Msg {
		return ResultMsg{Result: result}
	}
}

func (m Model) onAnimate() (Model, tea.Cmd) {
	if m.fade.Update() {
		return m, animation.Tick()
	}

	return m, nil
}

func wrapWidth(width int) int {
	if width <= 0 {
		return 80
	}

	return max(min(width-8, 60), 20)
}

func clampLines(n int) int {
	return min(max(n, 3), 6)
}
