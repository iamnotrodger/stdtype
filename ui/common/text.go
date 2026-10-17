package common

const LineGap = 0

type LineSpan struct {
	Start int
	End   int
}

func WrapWidth(width int) int {
	if width <= 0 {
		return 80
	}

	wrap := min(width-8, 60)
	wrap = max(wrap, 20)

	return wrap
}

func VisibleLines(n int) int {
	return min(max(n, 3), 6)
}

func WrapLineSpans(target []rune, wrap int) []LineSpan {
	var spans []LineSpan
	lineStart, lastSpace, col := 0, -1, 0

	for i, r := range target {
		col++

		if r == ' ' {
			lastSpace = i
		}

		if col > wrap {
			brk := lastSpace
			if brk < lineStart {
				brk = i
			}

			spans = append(spans, LineSpan{lineStart, brk})
			lineStart = brk + 1
			col = i - lineStart + 1
			lastSpace = -1
		}
	}

	spans = append(spans, LineSpan{lineStart, len(target)})

	return spans
}

func CaretLineIndex(spans []LineSpan, caret int) int {
	for i, s := range spans {
		if caret <= s.End {
			return i
		}
	}

	return len(spans) - 1
}

func ViewportRows(visible int) int {
	return visible + (visible-1)*LineGap
}

func LineOffset(target []rune, wrap, visible, caret int) int {
	spans := WrapLineSpans(target, wrap)
	top := CaretLineIndex(spans, caret) - 1

	if top+visible > len(spans) {
		top = max(0, len(spans)-visible)
	}

	return top * (1 + LineGap)
}
