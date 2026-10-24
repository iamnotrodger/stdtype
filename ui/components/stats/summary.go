package stats

import (
	"time"

	stdtype "github.com/iamnotrodger/stdytype"
)

type summary struct {
	tests      int
	timeTyping time.Duration
	bestWPM    float64
	avgWPM     float64
}

// summarize skips invalid results, as they don't reflect real typing.
func summarize(results []*stdtype.Result) summary {
	var s summary
	var sumWPM float64

	for _, r := range results {
		if r.Invalid {
			continue
		}

		s.tests++
		s.timeTyping += r.Duration
		s.bestWPM = max(s.bestWPM, r.Stats.WPM)
		sumWPM += r.Stats.WPM
	}

	if s.tests > 0 {
		s.avgWPM = sumWPM / float64(s.tests)
	}

	return s
}
