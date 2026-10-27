package stats

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	t.Run("should send ResultMsg with highlighted result when enter pressed", func(t *testing.T) {
		t.Parallel()

		first := &stdtype.Result{Mode: "words", Amount: 25}
		second := &stdtype.Result{Mode: "time", Amount: 30}

		m := NewModel()
		m.SetResults([]*stdtype.Result{first, second})
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		require.NotNil(t, cmd)
		assert.Equal(t, ResultMsg{Result: second}, cmd())
	})

	t.Run("should do nothing when enter pressed with no results", func(t *testing.T) {
		t.Parallel()

		m := NewModel()
		m.SetResults(nil)

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Nil(t, cmd)
	})
}

func TestModelView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		results   []*stdtype.Result
		wantShown []string
	}{
		{
			name: "should show summary of valid results when results set",
			results: []*stdtype.Result{
				{Mode: "words", Amount: 25, Duration: 30 * time.Second, Stats: stdtype.Stats{WPM: 60}},
				{Mode: "time", Amount: 30, Duration: 90 * time.Second, Stats: stdtype.Stats{WPM: 80}},
				{Mode: "time", Amount: 15, Duration: time.Hour, Stats: stdtype.Stats{WPM: 200}, Invalid: true},
			},
			wantShown: []string{"completed", "2", "time typing", "00:02:00", "best wpm", "80", "avg wpm", "70"},
		},
		{
			name: "should list each result as a row when results set",
			results: []*stdtype.Result{
				{
					Mode:        "words",
					Amount:      25,
					Consistency: 81,
					Timestamp:   time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
					Stats:       stdtype.Stats{WPM: 72, RawWPM: 75, Accuracy: 96},
				},
			},
			wantShown: []string{"72", "75", "96%", "81%", "words 25", "26 Oct 2026"},
		},
		{
			name:      "should show zero summary when no results",
			wantShown: []string{"completed", "0", "00:00:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewModel()
			m.SetSize(100, 40)
			m.SetResults(tt.results)

			view := ansi.Strip(m.View())

			for _, want := range tt.wantShown {
				assert.Contains(t, view, want)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results []*stdtype.Result
		want    summary
	}{
		{
			name: "should return zero summary when no results",
			want: summary{},
		},
		{
			name: "should total tests and time when results valid",
			results: []*stdtype.Result{
				{Duration: 30 * time.Second, Stats: stdtype.Stats{WPM: 60}},
				{Duration: 90 * time.Second, Stats: stdtype.Stats{WPM: 80}},
			},
			want: summary{tests: 2, timeTyping: 2 * time.Minute, bestWPM: 80, avgWPM: 70},
		},
		{
			name: "should skip result when it is invalid",
			results: []*stdtype.Result{
				{Duration: 30 * time.Second, Stats: stdtype.Stats{WPM: 60}},
				{Duration: time.Hour, Stats: stdtype.Stats{WPM: 200}, Invalid: true},
			},
			want: summary{tests: 1, timeTyping: 30 * time.Second, bestWPM: 60, avgWPM: 60},
		},
		{
			name: "should return zero summary when every result invalid",
			results: []*stdtype.Result{
				{Duration: time.Minute, Stats: stdtype.Stats{WPM: 90}, Invalid: true},
			},
			want: summary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, summarize(tt.results))
		})
	}
}
