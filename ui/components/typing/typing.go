package typing

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pkg/errors"

	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/iamnotrodger/stdtype/ui/animation"
	"github.com/iamnotrodger/stdtype/ui/common"
)

type TickMsg struct {
	ID   int
	Time time.Time
}

type ResultMsg struct {
	Result *stdtype.Result
}

type ErrMsg struct {
	Err error
}

func errCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrMsg{Err: err}
	}
}

const wordBuffer = 30

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
) (Model, error) {
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
		return m.onAnimate(msg)
	case tea.KeyPressMsg:
		return m.onKey(msg)
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	m.textarea.SetFade(m.fade.Value)

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
	m.textarea.SetSize(wrapWidth(width), clampLines(m.settings.Lines))
}

func (m Model) reset() (Model, error) {
	m.mode = stdtype.ParseMode(m.settings.Mode)

	count := m.settings.WordCount
	if m.mode == stdtype.TimeMode {
		count = m.settings.Duration*3 + 30
	}

	words, err := m.generate(count)

	m.test = stdtype.NewTypeTest(words, m.mode)
	m.samples = stdtype.Samples{}
	m.prevStats = stdtype.Stats{}
	m.prevElapsed = 0
	m.afkSeconds = 0
	m.tickID++

	m.textarea.SetSize(wrapWidth(m.width), clampLines(m.settings.Lines))
	m.textarea.SetTest(m.test)

	return m, err
}

func (m Model) generate(count int) ([]string, error) {
	words, err := m.wordSource.Generate(
		count,
		stdtype.GenerateOptions{
			List:        m.settings.WordList,
			Punctuation: m.settings.Punctuation,
			Numbers:     m.settings.Numbers,
		},
	)

	return words, errors.Wrap(err, "wordSource.Generate")
}

func (m Model) extendWords() tea.Cmd {
	if m.mode != stdtype.TimeMode || len(m.test.Words)-m.test.ActiveWord() > wordBuffer {
		return nil
	}

	words, err := m.generate(wordBuffer)
	if err != nil {
		return errCmd(err)
	}

	for _, w := range words {
		m.test.Words = append(m.test.Words, []rune(w))
	}

	return nil
}

func (m Model) Restart() (Model, tea.Cmd) {
	m, err := m.reset()
	m.fade.Start()

	cmd := tea.Batch(animation.Tick(m.tickID), cursor.Blink)
	if err != nil {
		cmd = tea.Batch(cmd, errCmd(err))
	}

	return m, cmd
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

	rawDelta := stats.Keystrokes - m.prevStats.Keystrokes

	errKeys := stats.Keystrokes - stats.CorrectKeystrokes
	prevErrKeys := m.prevStats.Keystrokes - m.prevStats.CorrectKeystrokes
	errDelta := max(0, errKeys-prevErrKeys)

	m.samples.Raw = append(m.samples.Raw, wpm(rawDelta, secs))
	m.samples.WPM = append(m.samples.WPM, stats.WPM)
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
	extendCmd := m.extendWords()

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmd = tea.Batch(cmd, extendCmd)

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

	amount := m.settings.Duration
	if m.mode == stdtype.WordsMode {
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

func (m Model) onAnimate(msg animation.Msg) (Model, tea.Cmd) {
	if msg.ID != m.tickID {
		return m, nil
	}

	if m.fade.Update() {
		return m, animation.Tick(m.tickID)
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
