package result

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdytype"
)

func TestModelView(t *testing.T) {
	t.Parallel()

	t.Run("should render nothing when no result set", func(t *testing.T) {
		t.Parallel()

		m := NewModel()
		m.SetSize(80, 30)

		if view := m.View(); view != "" {
			t.Errorf("view = %q, want empty", view)
		}
	})

	tests := []struct {
		name        string
		result      stdtype.Result
		wantShown   []string
		wantMissing []string
	}{
		{
			name: "should show stats when result set",
			result: stdtype.Result{
				Mode:        "words",
				Amount:      25,
				Duration:    12500 * time.Millisecond,
				Consistency: 80,
				Stats:       stdtype.Stats{WPM: 72, RawWPM: 75, Accuracy: 96},
			},
			wantShown: []string{"wpm", "72", "accuracy", "96%", "words 25", "75", "80%", "12.5s"},
		},
		{
			name: "should show reason when result invalid",
			result: stdtype.Result{
				Mode:          "time",
				Amount:        30,
				Invalid:       true,
				InvalidReason: "afk detected",
			},
			wantShown: []string{"afk detected -- result not saved"},
		},
		{
			name:        "should not show not-saved notice when result valid",
			result:      stdtype.Result{Mode: "time", Amount: 30},
			wantMissing: []string{"result not saved"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewModel()
			m.SetSize(80, 30)
			m.SetResult(&tt.result)

			view := ansi.Strip(m.View())

			for _, want := range tt.wantShown {
				if !strings.Contains(view, want) {
					t.Errorf("view does not show %q:\n%s", want, view)
				}
			}

			for _, missing := range tt.wantMissing {
				if strings.Contains(view, missing) {
					t.Errorf("view shows %q:\n%s", missing, view)
				}
			}
		})
	}
}
