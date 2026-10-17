package typing

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/iamnotrodger/stdytype/ui/animation"
	"github.com/iamnotrodger/stdytype/ui/common"
	"github.com/iamnotrodger/stdytype/ui/themes"
)

type view struct {
	theme themes.Theme
	mode  stdtype.Mode

	target []rune
	typed  []rune

	duration    int
	wordCount   int
	visibleLine int
	wrap        int
	remaining   int

	caretCell func(ch string) string

	fade float64
}

func (v view) render() string {
	spans := wrapLineSpans(v.target, v.wrap)
	strip := make([]string, 0, len(spans))

	for _, s := range spans {
		ln := v.line(s)
		if pad := v.wrap - lipgloss.Width(ln); pad > 0 {
			ln += strings.Repeat(" ", pad)
		}

		strip = append(strip, ln)
	}

	lines := visibleLines(v.visibleLine)
	offset := lineOffset(spans, lines, v.caret())

	maxOffset := len(strip) - lines
	offset = max(0, min(offset, maxOffset))

	end := min(offset+lines, len(strip))
	out := append([]string{}, strip[offset:end]...)

	for len(out) < lines {
		out = append(out, "")
	}

	return strings.Join(out, "\n")
}

func (v view) modeBar() string {
	opt := func(label string, on bool) string {
		if on {
			return v.theme.Accent.Render(label)
		}

		return v.theme.Dim.Render(label)
	}

	isTimeMode := v.mode == stdtype.TimeMode
	modes := opt("time", isTimeMode) + " " + opt("words", !isTimeMode)

	selected := v.wordCount
	options := common.WordOptions
	if isTimeMode {
		selected = v.duration
		options = common.TimeOptions
	}

	labels := make([]string, 0, len(options))
	for _, d := range options {
		label := strconv.Itoa(d)
		on := d == selected
		labels = append(labels, opt(label, on))
	}

	sep := v.theme.Dim.Render(" | ")

	return modes + sep + strings.Join(labels, " ")
}

func (v view) line(s lineSpan) string {
	var b strings.Builder

	if v.fade < 1 {
		fs := animation.FadeStyle(v.fade)
		for i := s.start; i < s.end; i++ {
			b.WriteString(fs.Render(string(v.target[i])))
		}

		return b.String()
	}

	for i := s.start; i < s.end; i++ {
		b.WriteString(v.character(i))
	}

	if v.caret() == s.end {
		b.WriteString(v.caretCell(" "))
	}

	return b.String()
}

func (v view) character(index int) string {
	target := v.target[index]

	if index == v.caret() {
		return v.caretCell(string(target))
	}

	if index > v.caret() {
		return v.theme.Pending.Render(string(target))
	}

	if v.typed[index] == target {
		return v.theme.Correct.Render(string(target))
	}

	if target == ' ' {
		target = '_'
	}

	return v.theme.Incorrect.Render(string(target))
}

func (v view) caret() int {
	return len(v.typed)
}

func (v view) counter() string {
	if v.mode == stdtype.TimeMode {
		remaining := strconv.Itoa(max(v.remaining, 0))

		return animation.FadeRender(v.fade, remaining)
	}

	count := 0
	typed := strings.TrimRight(string(v.typed), " ")

	if typed != "" {
		count = strings.Count(typed, " ")
	}

	progress := fmt.Sprintf("%d/%d", count, v.wordCount)

	return animation.FadeRender(v.fade, progress)
}
