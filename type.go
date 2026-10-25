package stdtype

import (
	"math"
	"strings"
	"time"
	"unicode"
)

const (
	AFKThreshold  = 0.5
	MinAFKSeconds = 7
)

func NewTypeTest(words []string, mode Mode) *TypeTest {
	return &TypeTest{
		Target: []rune(strings.Join(words, " ")),
		Mode:   mode,
		clock:  time.Now,
	}
}

func (t *TypeTest) Type(r rune) {
	if t.finished || len(t.Typed) >= len(t.Target) {
		return
	}

	if !t.started {
		t.started = true
		t.startedAt = t.clock()
	}

	idx := len(t.Typed)
	t.totalKeystrokes++

	if r == t.Target[idx] {
		t.correctKeyStrokes++
	}

	t.Typed = append(t.Typed, r)
}

func (t *TypeTest) Backspace() {
	if len(t.Typed) > 0 {
		t.Typed = t.Typed[:len(t.Typed)-1]
	}
}

func (t *TypeTest) DeleteWord() {
	i := len(t.Typed)

	for i > 0 && unicode.IsSpace(t.Typed[i-1]) {
		i--
	}

	for i > 0 && !unicode.IsSpace(t.Typed[i-1]) {
		i--
	}

	t.Typed = t.Typed[:i]
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
		return len(t.Typed) >= len(t.Target)
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

// CorrectChars counts typed characters that match the target at the same position.
func (t *TypeTest) CorrectChars() int {
	n := 0

	for i, r := range t.Typed {
		if i < len(t.Target) && r == t.Target[i] {
			n++
		}
	}

	return n
}

// ComputeStats WPM uses the standard 5-characters-per-word convention over correct characters;
// raw uses all typed characters.
func ComputeStats(t *TypeTest) Stats {
	mins := t.Elapsed().Minutes()
	correct := t.CorrectChars()
	typed := len(t.Typed)

	var wpm, raw float64

	if mins > 0 {
		wpm = (float64(correct) / 5.0) / mins
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
		CorrectChars:      correct,
		TypedChars:        typed,
		Keystrokes:        t.totalKeystrokes,
		IncorrectChars:    typed - correct,
		MissedChars:       0,
		CorrectKeystrokes: t.correctKeyStrokes,
	}
}

// Consistency scores stead-state raw-WPM samples as 100*(1-stddev/mean),
// clampled to [0, 100].
func Consistency(rawSamples []float64) float64 {
	if len(rawSamples) < 2 {
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
	score := 100.0 * (1.0 - stdDev/mean)

	return math.Max(0, math.Min(100, score))
}

func IsAFK(afkSeconds, totalSeconds int) bool {
	if totalSeconds < MinAFKSeconds {
		return false
	}

	return float64(afkSeconds)/float64(totalSeconds) >= AFKThreshold
}
