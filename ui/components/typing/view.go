package typing

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	stdtype "github.com/iamnotrodger/stdytype"
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
	spans := common.WrapLineSpans(v.target, v.wrap)

	var strip []string
	for li, s := range spans {
		ln := v.line(s)
		if pad := v.wrap - lipgloss.Width(ln); pad > 0 {
			ln += strings.Repeat(" ", pad)
		}

		strip = append(strip, ln)

		if li < len(spans)-1 {
			for range common.LineGap {
				strip = append(strip, "")
			}
		}
	}

	vh := common.ViewportRows(common.VisibleLines(v.visibleLine))
	offset := common.LineOffset(v.target, v.wrap, v.visibleLine, v.caret())

	if maxOffset := len(strip) - vh; offset > maxOffset {
		offset = maxOffset
	}

	if offset < 0 {
		offset = 0
	}

	end := min(offset+vh, len(strip))
	out := append([]string{}, strip[offset:end]...)

	for len(out) < vh {
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

func (v view) line(s common.LineSpan) string {
	var b strings.Builder

	if v.fade < 1 {
		fs := fadeStyle(v.fade)
		for i := s.Start; i < s.End; i++ {
			b.WriteString(fs.Render(string(v.target[i])))
		}

		return b.String()
	}

	for i := s.Start; i < s.End; i++ {
		b.WriteString(v.character(i))
	}

	if v.caret() == s.End {
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
		return strconv.Itoa(max(v.remaining, 0))
	}

	words := 0

	typed := strings.TrimRight(string(v.typed), " ")
	if typed != "" {
		words = strings.Count(typed, " ") + 1
	}

	return fmt.Sprintf("%d/%d", words, v.wordCount)
}

func fadeStyle(t float64) lipgloss.Style {
	const gray = 0x88

	v := int(math.Round(float64(gray) * t))
	hex := fmt.Sprintf("#%02x%02x%02x", v, v, v)

	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}
