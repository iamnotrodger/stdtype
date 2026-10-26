package typing

import (
	"strings"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
)

type KeyMap struct {
	DeleteCharacterBackward key.Binding
	DeleteWordBackward      key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		DeleteCharacterBackward: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "delete"),
		),
		DeleteWordBackward: key.NewBinding(
			key.WithKeys("alt+backspace"),
			key.WithHelp("alt+backspace", "delete word"),
		),
	}
}

type TextAreaStyles struct {
	Untyped    lipgloss.Style
	Correct    lipgloss.Style
	Incorrect  lipgloss.Style
	Cursor     lipgloss.Style
	CursorText lipgloss.Style
}

func DefaultTextAreaStyles() TextAreaStyles {
	return TextAreaStyles{
		Untyped:    lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
		Correct:    lipgloss.NewStyle().Foreground(lipgloss.Green),
		Incorrect:  lipgloss.NewStyle().Foreground(lipgloss.Red).Underline(true),
		Cursor:     lipgloss.NewStyle().Reverse(true),
		CursorText: lipgloss.NewStyle().Foreground(lipgloss.BrightBlack),
	}
}

type TextArea struct {
	KeyMap KeyMap

	test   *stdtype.TypeTest
	styles TextAreaStyles
	cursor cursor.Model

	width, height int
	fade          float64
}

func NewTextArea() TextArea {
	m := TextArea{
		KeyMap: DefaultKeyMap(),
		cursor: cursor.New(),
		fade:   1,
	}
	m.SetStyle(DefaultTextAreaStyles())
	m.cursor.Focus()

	return m
}

func (m TextArea) Update(msg tea.Msg) (TextArea, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		return m.onKey(msg), nil
	}

	var cmd tea.Cmd
	m.cursor, cmd = m.cursor.Update(msg)

	return m, cmd
}

func (m TextArea) onKey(msg tea.KeyPressMsg) TextArea {
	// Keys can still arrive before the parent handles ResultMsg.
	if m.test == nil || m.test.Finished() {
		return m
	}

	switch {
	case key.Matches(msg, m.KeyMap.DeleteWordBackward):
		m.test.DeleteWord()
	case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
		m.test.Backspace()
	default:
		for _, r := range msg.Text {
			m.test.Type(r)
		}
	}

	if m.test.Started() {
		m.cursor.SetMode(cursor.CursorStatic)
	}

	return m
}

func (m TextArea) View() string {
	if m.test == nil {
		return ""
	}

	lines := wrapLines(m.test.Target, m.width)
	top := scrollTop(lines, m.height, len(m.test.Typed))
	visible := lines[top:min(top+m.height, len(lines))]

	rows := make([]string, 0, m.height)
	for _, l := range visible {
		row := m.renderLine(l)
		if pad := m.width - lipgloss.Width(row); pad > 0 {
			row += strings.Repeat(" ", pad)
		}

		rows = append(rows, row)
	}

	for len(rows) < m.height {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func (m *TextArea) SetTest(t *stdtype.TypeTest) {
	m.test = t
	m.cursor.SetMode(cursor.CursorBlink)
}

func (m *TextArea) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *TextArea) SetFade(v float64) {
	m.fade = v
}

func (m TextArea) Styles() TextAreaStyles {
	return m.styles
}

func (m *TextArea) SetStyle(s TextAreaStyles) {
	m.styles = s
	m.cursor.Style = s.Cursor
	m.cursor.TextStyle = s.CursorText
}

func (m TextArea) renderLine(l line) string {
	target := m.test.Target

	var b strings.Builder

	if m.fade < 1 {
		fs := animation.FadeStyle(m.fade)
		for i := l.start; i < l.end; i++ {
			b.WriteString(fs.Render(string(target[i])))
		}

		return b.String()
	}

	for i := l.start; i < l.end; i++ {
		b.WriteString(m.character(i))
	}

	return b.String()
}

func (m TextArea) character(index int) string {
	target := m.test.Target[index]
	caret := len(m.test.Typed)

	if index == caret {
		m.cursor.SetChar(string(target))

		return m.cursor.View()
	}

	if index > caret {
		return m.styles.Untyped.Render(string(target))
	}

	if m.test.Typed[index] == target {
		return m.styles.Correct.Render(string(target))
	}

	if target == ' ' {
		target = '_'
	}

	return m.styles.Incorrect.Render(string(target))
}

// line is the range of Target runes [start, end) shown on one row.
type line struct {
	start int
	end   int
}

// wrapLines breaks target at word boundaries, measured in terminal columns.
func wrapLines(target []rune, width int) []line {
	var lines []line
	start, col := 0, 0

	for i := 0; i < len(target); {
		wordEnd := i
		for wordEnd < len(target) && target[wordEnd] != ' ' {
			wordEnd++
		}

		next := wordEnd
		for next < len(target) && target[next] == ' ' {
			next++
		}

		if col > 0 && col+textWidth(target[i:next]) > width {
			lines = append(lines, line{start, i})
			start, col = i, 0
		}

		// A word wider than the line is cut wherever it overflows.
		for textWidth(target[i:wordEnd]) > width {
			breakAt := cutAt(target[i:wordEnd], width) + i
			lines = append(lines, line{start, breakAt})
			start, i = breakAt, breakAt
		}

		col += textWidth(target[i:next])
		i = next
	}

	return append(lines, line{start, len(target)})
}

// cutAt returns how many runes of word fit in width, at least one so a rune
// wider than width still makes progress.
func cutAt(word []rune, width int) int {
	col := 0
	for i, r := range word {
		col += textWidth([]rune{r})
		if col > width {
			return max(i, 1)
		}
	}

	return len(word)
}

func textWidth(runes []rune) int {
	return ansi.StringWidth(string(runes))
}

func cursorLine(lines []line, caret int) int {
	for i, l := range lines {
		if caret < l.end {
			return i
		}
	}

	return len(lines) - 1
}

// scrollTop keeps the cursor on the second visible line.
func scrollTop(lines []line, height, caret int) int {
	top := cursorLine(lines, caret) - 1

	return max(0, min(top, len(lines)-height))
}
