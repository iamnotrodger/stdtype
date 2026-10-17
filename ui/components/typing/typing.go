package typing

import (
	"time"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	"github.com/iamnotrodger/stdytype/ui/common"
	appctx "github.com/iamnotrodger/stdytype/ui/context"
	"github.com/iamnotrodger/stdytype/ui/keys"
)

type TickMsg time.Time

func Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type Model struct {
	appCtx     *appctx.Context
	wordSource stdtype.WordSource

	mode    stdtype.Mode
	test    *stdtype.TypeTest
	result  *stdtype.Result
	samples stdtype.Samples

	prevStats  stdtype.Stats
	afkSeconds int

	caret cursor.Model
	keys  *keys.KeyMap
	fade  animation.Fade
}

func NewModel(appCtx *appctx.Context, wordSource stdtype.WordSource) Model {
	caret := cursor.New()
	caret.Focus()

	m := Model{
		appCtx:     appCtx,
		wordSource: wordSource,
		caret:      caret,
		keys:       keys.Keys,
		fade:       animation.NewFade(),
	}

	return m.reset()
}

func (t Model) Init() tea.Cmd {
	return tea.Batch(Tick(), t.caret.Blink())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		return m.onTick()
	case animation.Msg:
		return m.onAnimate()
	case tea.KeyPressMsg:
		return m.onKey(msg)
	}

	var cmd tea.Cmd
	m.caret, cmd = m.caret.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	settings := m.appCtx.Settings
	w, h := m.appCtx.Dimension()

	elapsed := m.test.Elapsed().Seconds()
	remaining := settings.Duration - int(elapsed)

	v := view{
		theme:       m.appCtx.Theme,
		mode:        m.test.Mode,
		target:      m.test.Target,
		typed:       m.test.Typed,
		duration:    settings.Duration,
		wordCount:   settings.WordCount,
		visibleLine: settings.Lines,
		remaining:   remaining,
		wrap:        common.WrapWidth(w),
		caretCell:   m.caretCell,
		fade:        m.fade.Value,
	}

	count := v.counter()
	main := lipgloss.JoinVertical(
		lipgloss.Center,
		count,
		"",
		v.render(),
	)

	return common.Layout{
		Theme:  m.appCtx.Theme,
		Header: v.modeBar(),
		Main:   main,
		Width:  w,
		Height: h,
	}.View()
}

func (m Model) reset() Model {
	settings := m.appCtx.Settings
	mode := stdtype.ParseMode(settings.Mode)

	count := settings.WordCount
	if mode == stdtype.TimeMode {
		count = settings.Duration*3 + 30
	}

	// TODO: log error
	words, _ := m.wordSource.Generate(
		count,
		stdtype.GenerateOptions{
			List:        settings.WordList,
			Punctuation: settings.Punctuation,
			Numbers:     settings.Numbers,
		},
	)

	m.mode = stdtype.ParseMode(settings.Mode)
	m.test = stdtype.NewTypeTest(words, m.mode)
	m.samples = stdtype.Samples{}
	m.prevStats = stdtype.Stats{}
	m.result = nil
	m.afkSeconds = 0

	return m
}

func (m Model) Restart() (Model, tea.Cmd) {
	m = m.reset()
	m.fade.Start()

	return m, animation.Tick()
}

func (m Model) Completed() bool {
	return m.test.Completed()
}

func (m Model) Result() *stdtype.Result {
	return m.result
}

func (m Model) onTick() (Model, tea.Cmd) {
	if !m.test.Started() {
		return m, Tick()
	}

	settings := m.appCtx.Settings
	stats := stdtype.ComputeStats(m.test)

	rawDelta := max(0, stats.TypedChars-m.prevStats.TypedChars)
	netDelta := max(0, stats.CorrectChars-m.prevStats.CorrectChars)

	errKeys := stats.Keystrokes - stats.CorrectKeystrokes
	prevErrKeys := m.prevStats.Keystrokes - m.prevStats.CorrectChars
	errDelta := max(0, errKeys-prevErrKeys)

	// Per-second WPM samples (chars/5 per 1/60 min).
	m.samples.Raw = append(m.samples.Raw, float64(rawDelta)/5.0*60.0)
	m.samples.WPM = append(m.samples.WPM, float64(netDelta)/5.0*60.0)
	m.samples.Err = append(m.samples.Err, errDelta)

	if stats.Keystrokes == m.prevStats.Keystrokes {
		m.afkSeconds++
	}

	m.prevStats = stats

	if m.mode == stdtype.TimeMode {
		duration := time.Duration(settings.Duration) * time.Second
		if m.test.Elapsed() >= duration {
			return m.onFinish(), Tick()
		}
	}

	return m, Tick()
}

func (m Model) caretCell(ch string) string {
	c := m.caret
	c.SetChar(ch)

	if m.test.Started() {
		c.IsBlinked = false
	}

	return c.View()
}

func (m Model) onKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.DeleteWord):
		m.test.DeleteWord()
	case key.Matches(msg, m.keys.Backspace):
		m.test.Backspace()
	case msg.Text != "":
		for _, r := range msg.Text {
			m.test.Type(r)
		}
	default:
		return m, nil
	}

	if m.test.Completed() {
		return m.onFinish(), nil
	}

	return m, nil
}

func (m Model) onFinish() Model {
	settings := m.appCtx.Settings

	m.test.Finish()
	stats := stdtype.ComputeStats(m.test)
	consistency := stdtype.Consistency(m.samples.Raw)

	mode := stdtype.ParseMode(settings.Mode)
	amount := settings.Duration
	if mode == stdtype.WordsMode {
		amount = settings.WordCount
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

	if stdtype.IsAFK(m.afkSeconds, int(m.test.Elapsed())) {
		result.Invalid = true
		result.InvalidReason = "afk detected"
	}

	m.result = result

	return m
}

func (m Model) onAnimate() (Model, tea.Cmd) {
	if m.fade.Update() {
		return m, animation.Tick()
	}

	return m, nil
}
