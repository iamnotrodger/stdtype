package result

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	stdtype "github.com/iamnotrodger/stdytype"
)

func TestGraphView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		samples     stdtype.Samples
		wantEmpty   bool
		wantShown   []string
		wantMissing []string
	}{
		{
			name: "should render nothing when fewer than two samples",
			samples: stdtype.Samples{
				WPM: []float64{60},
				Raw: []float64{60},
				Err: []int{0},
			},
			wantEmpty: true,
		},
		{
			name: "should label x axis in elapsed seconds from zero",
			samples: stdtype.Samples{
				WPM: []float64{40, 50, 60},
				Raw: []float64{45, 55, 65},
				Err: []int{0, 0, 0},
			},
			wantShown: []string{"0s"},
		},
		{
			name: "should mark errors when a second had errors",
			samples: stdtype.Samples{
				WPM: []float64{40, 50, 60},
				Raw: []float64{45, 55, 65},
				Err: []int{0, 2, 0},
			},
			wantShown: []string{"×"},
		},
		{
			name: "should hide error marks when no errors",
			samples: stdtype.Samples{
				WPM: []float64{40, 50, 60},
				Raw: []float64{45, 55, 65},
				Err: []int{0, 0, 0},
			},
			wantMissing: []string{"×"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := NewGraph()
			g.SetSize(40, 10)
			g.SetSamples(tt.samples)

			view := ansi.Strip(g.View())

			if tt.wantEmpty && view != "" {
				t.Errorf("view = %q, want empty", view)
			}

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
