package stdtype

import (
	"math"
	"slices"
	"time"
)

const (
	AFKThreshold  = 0.5
	MinAFKSeconds = 7

	MaxExtraChars = 20
)

func NewTypeTest(words []string, mode Mode) *TypeTest {
	runes := make([][]rune, len(words))
	for i, w := range words {
		runes[i] = []rune(w)
	}

	return &TypeTest{
		Words:  runes,
		Inputs: [][]rune{{}},
		Mode:   mode,
		clock:  time.Now,
	}
}

func (t *TypeTest) Type(r rune) {
	active := t.ActiveWord()
	if t.finished || active >= len(t.Words) {
		return
	}

	word, input := t.Words[active], t.Inputs[active]

	if r == ' ' && len(input) == 0 {
		return
	}

	if r != ' ' && len(input) >= len(word)+MaxExtraChars {
		return
	}

	if !t.started {
		t.started = true
		t.startedAt = t.clock()
	}

	t.totalKeystrokes++

	if r == ' ' {
		// The last word has no trailing space, so committing it is a miss.
		if len(input) == len(word) && active < len(t.Words)-1 {
			t.correctKeyStrokes++
		}

		t.Inputs = append(t.Inputs, []rune{})

		return
	}

	if len(input) < len(word) && word[len(input)] == r {
		t.correctKeyStrokes++
	}

	t.Inputs[active] = append(input, r)
}

func (t *TypeTest) Backspace() {
	active := t.ActiveWord()
	if input := t.Inputs[active]; len(input) > 0 {
		t.Inputs[active] = input[:len(input)-1]

		return
	}

	if t.canGoBack() {
		t.Inputs = t.Inputs[:active]
	}
}

func (t *TypeTest) DeleteWord() {
	active := t.ActiveWord()
	if len(t.Inputs[active]) > 0 {
		t.Inputs[active] = []rune{}

		return
	}

	if t.canGoBack() {
		t.Inputs = t.Inputs[:active]
		t.Inputs[active-1] = []rune{}
	}
}

func (t *TypeTest) canGoBack() bool {
	prev := t.ActiveWord() - 1

	return prev >= 0 && !slices.Equal(t.Inputs[prev], t.Words[prev])
}

func (t *TypeTest) ActiveWord() int {
	return len(t.Inputs) - 1
}

func (t *TypeTest) Started() bool {
	return t.started
}

func (t *TypeTest) Finished() bool {
	return t.finished
}

func (t *TypeTest) Completed() bool {
	switch t.Mode {
	case WordsMode:
		last := len(t.Words) - 1
		active := t.ActiveWord()

		return active > last ||
			(active == last && slices.Equal(t.Inputs[active], t.Words[active]))
	case TimeMode:
		return t.finished
	default:
		return false
	}
}

func (t *TypeTest) Finish() {
	if !t.finished {
		t.finished = true
		t.finishedAt = t.clock()
	}
}

// Elapsed returns time since the first keystroke (frozen after finished).
func (t *TypeTest) Elapsed() time.Duration {
	if !t.started {
		return 0
	}

	end := t.clock()
	if t.finished {
		end = t.finishedAt
	}

	return end.Sub(t.startedAt)
}

type charCounts struct {
	allCorrect  int
	correctWord int
	incorrect   int
	extra       int
	missed      int
}

// Words are compared with their commit space, so a space only counts as
// correct when its word is.
func (t *TypeTest) chars() charCounts {
	var total charCounts

	active := t.ActiveWord()

	for i := 0; i < len(t.Inputs) && i < len(t.Words); i++ {
		committed := i < active
		input, target := t.Inputs[i], t.Words[i]

		if !committed && len(input) == 0 {
			break
		}

		if committed {
			input = append(slices.Clone(input), ' ')
		}

		if i < len(t.Words)-1 {
			target = append(slices.Clone(target), ' ')
		}

		// The active word is still being typed, so its untyped rest isn't missed.
		c := countChars(input, target, !committed)

		total.allCorrect += c.allCorrect
		total.correctWord += c.correctWord
		total.incorrect += c.incorrect
		total.extra += c.extra
		total.missed += c.missed
	}

	return total
}

func countChars(input, target []rune, creditPartial bool) charCounts {
	var c charCounts

	wordCorrect := slices.Equal(input, target)
	partiallyCorrect := len(input) <= len(target) && slices.Equal(input, target[:len(input)])
	inputHasSpace := slices.Contains(input, ' ')

	for i := range max(len(input), len(target)) {
		switch {
		case i < len(input) && i < len(target) && input[i] == target[i]:
			if target[i] == ' ' && !wordCorrect {
				c.extra++
			} else {
				c.allCorrect++
			}

			if wordCorrect || (creditPartial && partiallyCorrect) {
				c.correctWord++
			}
		case i >= len(input):
			if !creditPartial {
				c.missed++
			}
		case i >= len(target) || (target[i] == ' ' && !inputHasSpace):
			c.extra++
		default:
			c.incorrect++
		}
	}

	return c
}

// ComputeStats WPM counts only characters of correctly typed words; raw counts
// every typed character.
func ComputeStats(t *TypeTest) Stats {
	mins := t.Elapsed().Minutes()
	c := t.chars()
	typed := c.allCorrect + c.incorrect + c.extra

	var wpm, raw float64

	if mins > 0 {
		wpm = (float64(c.correctWord) / 5.0) / mins
		raw = (float64(typed) / 5.0) / mins
	}

	acc := 100.0
	if t.totalKeystrokes > 0 {
		acc = float64(t.correctKeyStrokes) / float64(t.totalKeystrokes) * 100.0
	}

	return Stats{
		WPM:               wpm,
		RawWPM:            raw,
		Accuracy:          acc,
		CorrectChars:      c.correctWord,
		IncorrectChars:    c.incorrect,
		ExtraChars:        c.extra,
		MissedChars:       c.missed,
		TypedChars:        typed,
		Keystrokes:        t.totalKeystrokes,
		CorrectKeystrokes: t.correctKeyStrokes,
	}
}

// Consistency maps the coefficient of variation of raw samples onto [0, 100].
func Consistency(rawSamples []float64) float64 {
	if len(rawSamples) == 0 {
		return 0
	}

	var sum float64
	for _, v := range rawSamples {
		sum += v
	}

	mean := sum / float64(len(rawSamples))
	if mean == 0 {
		return 0
	}

	var sumSquaredDiff float64
	for _, sample := range rawSamples {
		diff := sample - mean
		sumSquaredDiff += diff * diff
	}

	stdDev := math.Sqrt(sumSquaredDiff / float64(len(rawSamples)))

	return kogasa(stdDev / mean)
}

// kogasa is a tanh curve that stays close to 100*(1-cov) for small cov but
// never drops below 0.
func kogasa(cov float64) float64 {
	cov3 := cov * cov * cov

	return 100 * (1 - math.Tanh(cov+cov3/3+cov3*cov*cov/5))
}

func IsAFK(afkSeconds, totalSeconds int) bool {
	if totalSeconds < MinAFKSeconds {
		return false
	}

	return float64(afkSeconds)/float64(totalSeconds) >= AFKThreshold
}
