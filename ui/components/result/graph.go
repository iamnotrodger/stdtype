package result

import (
	"fmt"
	"math"
	"slices"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/canvas"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/v2/linechart"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	stdtype "github.com/iamnotrodger/stdytype"
)

const (
	graphRows = 8
	yTicks    = 4

	rawSet = "raw"
	wpmSet = "wpm"
)

type GraphStyle struct {
	wpm lipgloss.Style
	raw lipgloss.Style
	err lipgloss.Style
}

type graphModel struct {
	samples stdtype.Samples
	start   time.Time
	style   GraphStyle

	width int
}

func (m graphModel) View() string {
	n := len(m.samples.WPM)
	if n < 2 {
		return ""
	}

	xStep := max(m.width/6, 6)
	times := sampleTimes(m.start, n)

	maxWpm, maxRaw := slices.Max(m.samples.WPM), slices.Max(m.samples.Raw)

	minStep := max(maxWpm, maxRaw) / yTicks
	roundStep := math.Ceil(minStep/10) * 10
	wpmStep := max(roundStep, 10)

	chart := timeserieslinechart.New(m.width, graphRows+2,
		timeserieslinechart.WithTimeRange(times[0], times[n-1]),
		timeserieslinechart.WithYRange(0, wpmStep*yTicks),
		timeserieslinechart.WithYLabelFormatter(intLabel),
		timeserieslinechart.WithXLabelFormatter(elapsedLabel(times[0])),
	)

	chart.SetXStep(xStep)
	chart.SetYStep(graphRows / yTicks)

	chart.SetDataSetLineStyle(rawSet, runes.ArcLineStyle)
	chart.SetDataSetLineStyle(wpmSet, runes.ArcLineStyle)
	chart.SetDataSetStyle(rawSet, m.style.raw)
	chart.SetDataSetStyle(wpmSet, m.style.wpm)

	for i := range n {
		chart.PushDataSet(rawSet, timeserieslinechart.TimePoint{Time: times[i], Value: m.samples.Raw[i]})
		chart.PushDataSet(wpmSet, timeserieslinechart.TimePoint{Time: times[i], Value: m.samples.WPM[i]})
	}

	chart.DrawBrailleDataSets([]string{rawSet, wpmSet})
	m.drawErrors(&chart, times)

	return chart.View()
}

// drawErrors marks each second that had errors with a "×" on a right Y axis.
// Adding that axis narrows the plot, so the lines are redrawn to match.
func (m graphModel) drawErrors(chart *timeserieslinechart.Model, times []time.Time) {
	errs := m.samples.Err[:min(len(m.samples.Err), len(times))]

	maxErr := slices.Max(append([]int{0}, errs...))
	if maxErr == 0 {
		return
	}

	chart.SetY2Axis(linechart.Axis{
		Max:        float64(maxErr),
		LabelStyle: m.style.err,
	})
	chart.DrawBrailleDataSets([]string{rawSet, wpmSet})

	for i, e := range errs {
		if e == 0 {
			continue
		}

		f := canvas.Float64Point{
			X: float64(times[i].UnixMilli()) / 1e3,
			Y: chart.Y2ToY(float64(e)),
		}

		chart.Canvas.SetCell(
			cellAt(chart, f),
			canvas.NewCellWithStyle('×', m.style.err),
		)
	}
}

func cellAt(tsc *timeserieslinechart.Model, f canvas.Float64Point) canvas.Point {
	x := tsc.ScaleFloat64Point(f).X
	y := tsc.ScaleFloat64PointForLine(f).Y

	p := canvas.CanvasPointFromFloat64Point(
		tsc.Origin(),
		canvas.Float64Point{
			X: x,
			Y: y,
		},
	)

	p.X++

	if p.Y >= tsc.Origin().Y {
		p.Y = tsc.Origin().Y - 1
	}

	return p
}

func intLabel(_ int, v float64) string {
	return fmt.Sprintf("%.0f", v)
}

func elapsedLabel(start time.Time) linechart.LabelFormatter {
	// match the chart's own X encoding (UnixMilli/1e3) so the first label is 0s, not -0s
	base := float64(start.UnixMilli()) / 1e3

	return func(_ int, v float64) string {
		return fmt.Sprintf("%.0fs", v-base)
	}
}

func sampleTimes(start time.Time, n int) []time.Time {
	ts := make([]time.Time, n)

	for i := range ts {
		ts[i] = start.Add(time.Duration(i) * time.Second)
	}

	return ts
}
