package typing

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextAreaUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		target    string
		modelFunc func(TextArea) TextArea
		want      string
	}{
		{
			name:   "should append runes when typing",
			target: "ab cd",
			modelFunc: func(m TextArea) TextArea {
				return sendString(m, "ab")
			},
			want: "ab",
		},
		{
			name:   "should remove last rune when backspace pressed",
			target: "ab cd",
			modelFunc: func(m TextArea) TextArea {
				m = sendString(m, "ax")
				m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

				return m
			},
			want: "a",
		},
		{
			name:   "should delete back to word start when delete word pressed",
			target: "ab cd",
			modelFunc: func(m TextArea) TextArea {
				m = sendString(m, "ab c")
				m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace, Mod: tea.ModAlt})

				return m
			},
			want: "ab ",
		},
		{
			name:   "should ignore keys when test finished",
			target: "ab cd",
			modelFunc: func(m TextArea) TextArea {
				m = sendString(m, "a")
				m.test.Finish()

				return sendString(m, "b")
			},
			want: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := tt.modelFunc(newTextArea(tt.target, 10, 3))

			assert.Equal(t, tt.want, string(m.test.Typed))
		})
	}
}

func TestCursorStopsBlinkingWhenTestStarts(t *testing.T) {
	m := newTextArea("ab", 10, 3)
	require.Equal(t, cursor.CursorBlink, m.cursor.Mode())

	m = sendString(m, "a")

	assert.Equal(t, cursor.CursorStatic, m.cursor.Mode())
}

func TestTextAreaView(t *testing.T) {
	t.Parallel()

	type want struct {
		// lines is the view without ANSI, including width padding.
		lines []string
		// contains is a styled fragment the raw view must include.
		contains func(TextAreaStyles) string
	}

	tests := []struct {
		name          string
		target        string
		width, height int
		modelFunc     func(TextArea) TextArea
		want          want
	}{
		{
			name:   "should pad empty lines when text shorter than height",
			target: "ab cd",
			width:  10, height: 3,
			want: want{lines: []string{"ab cd     ", "", ""}},
		},
		{
			name:   "should wrap at word boundary when next word and its space overflow",
			target: "aaa bbb ccc",
			width:  7, height: 3,
			want: want{lines: []string{"aaa    ", "bbb ccc", ""}},
		},
		{
			name:   "should keep trailing space on line when it fits",
			target: "aaa bbb ccc",
			width:  8, height: 3,
			want: want{lines: []string{"aaa bbb ", "ccc     ", ""}},
		},
		{
			name:   "should cut word without dropping runes when word wider than line",
			target: "abcdefgh ij",
			width:  4, height: 3,
			want: want{lines: []string{"abcd", "efgh ", "ij  "}},
		},
		{
			name:   "should wrap by display width when runes are wide",
			target: "a 日本",
			width:  4, height: 2,
			want: want{lines: []string{"a   ", "日本"}},
		},
		{
			name:   "should draw cursor on trailing space when caret at line end",
			target: "aaa bbb ccc",
			width:  8, height: 3,
			modelFunc: func(m TextArea) TextArea {
				return sendString(m, "aaa bbb")
			},
			want: want{contains: func(s TextAreaStyles) string {
				return s.Correct.Render("b") + s.Cursor.Inline(true).Reverse(true).Render(" ")
			}},
		},
		{
			name:   "should keep caret on second visible line when typing past first line",
			target: "aa bb cc dd",
			width:  3, height: 2,
			modelFunc: func(m TextArea) TextArea {
				return sendString(m, "aa bb ")
			},
			want: want{lines: []string{"bb ", "cc "}},
		},
		{
			name:   "should style char as untyped when after caret",
			target: "ab",
			width:  10, height: 3,
			want: want{contains: func(s TextAreaStyles) string {
				return s.Untyped.Render("b")
			}},
		},
		{
			name:   "should style char as correct when typed matches target",
			target: "ab",
			width:  10, height: 3,
			modelFunc: func(m TextArea) TextArea {
				return sendString(m, "a")
			},
			want: want{contains: func(s TextAreaStyles) string {
				return s.Correct.Render("a")
			}},
		},
		{
			name:   "should style target char as incorrect when typed differs",
			target: "ab",
			width:  10, height: 3,
			modelFunc: func(m TextArea) TextArea {
				return sendString(m, "x")
			},
			want: want{contains: func(s TextAreaStyles) string {
				return s.Incorrect.Render("a")
			}},
		},
		{
			name:   "should render underscore when space mistyped",
			target: "a b",
			width:  10, height: 3,
			modelFunc: func(m TextArea) TextArea {
				return sendString(m, "ax")
			},
			want: want{contains: func(s TextAreaStyles) string {
				return s.Incorrect.Render("_")
			}},
		},
		{
			name:   "should render faded text when fading",
			target: "ab",
			width:  10, height: 3,
			modelFunc: func(m TextArea) TextArea {
				m.SetFade(0.5)

				return m
			},
			want: want{contains: func(TextAreaStyles) string {
				return animation.FadeStyle(0.5).Render("a")
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newTextArea(tt.target, tt.width, tt.height)
			if tt.modelFunc != nil {
				m = tt.modelFunc(m)
			}

			view := m.View()

			if tt.want.lines != nil {
				assert.Equal(t, tt.want.lines, strings.Split(ansi.Strip(view), "\n"))
			}

			if tt.want.contains != nil {
				assert.Contains(t, view, tt.want.contains(m.Styles()))
			}
		})
	}
}

func newTextArea(target string, width, height int) TextArea {
	m := NewTextArea()
	m.SetTest(stdtype.NewTypeTest(strings.Fields(target), stdtype.WordsMode))
	m.SetSize(width, height)

	return m
}

func keyPress(r rune) tea.Msg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func sendString(m TextArea, s string) TextArea {
	for _, r := range s {
		m, _ = m.Update(keyPress(r))
	}

	return m
}
