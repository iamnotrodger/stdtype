package typing

type lineSpan struct {
	start int
	end   int
}

func wrapWidth(width int) int {
	if width <= 0 {
		return 80
	}

	return max(min(width-8, 60), 20)
}

func visibleLines(n int) int {
	return min(max(n, 3), 6)
}

func wrapLineSpans(target []rune, wrap int) []lineSpan {
	var spans []lineSpan
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

			spans = append(spans, lineSpan{lineStart, brk})
			lineStart = brk + 1
			col = i - lineStart + 1
			lastSpace = -1
		}
	}

	spans = append(spans, lineSpan{lineStart, len(target)})

	return spans
}

func caretLineIndex(spans []lineSpan, caret int) int {
	for i, s := range spans {
		if caret <= s.end {
			return i
		}
	}

	return len(spans) - 1
}

func lineOffset(spans []lineSpan, visible, caret int) int {
	top := caretLineIndex(spans, caret) - 1

	if top+visible > len(spans) {
		top = max(0, len(spans)-visible)
	}

	return top
}
