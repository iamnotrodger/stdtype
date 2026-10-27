package stdtype

import (
	"strings"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
)

func mustTypeAll(tb testing.TB, test *TypeTest, s string) {
	tb.Helper()

	for _, r := range s {
		test.Type(r)
	}
}

func mustTypedTestFinished(
	tb testing.TB,
	target []string,
	typed string,
) *TypeTest {
	tb.Helper()

	typeTest := NewTypeTest(target, WordsMode)
	clock := clockwork.NewFakeClockAt(time.Unix(1000, 0))
	typeTest.clock = clock.Now

	mustTypeAll(tb, typeTest, typed)

	clock.Advance(time.Minute)
	typeTest.Finish()

	return typeTest
}

func inputs(test *TypeTest) []string {
	out := make([]string, len(test.Inputs))
	for i, in := range test.Inputs {
		out[i] = string(in)
	}

	return out
}

func TestNewTypingTest(t *testing.T) {
	t.Run("should start on the first word with nothing typed", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"the", "quick", "fox"}, WordsMode)

		assert.Len(t, typeTest.Words, 3)
		assert.Equal(t, 0, typeTest.ActiveWord())
		assert.Equal(t, []string{""}, inputs(typeTest))
	})
}

func TestTypingTest_Type(t *testing.T) {
	t.Run("should append typed runes to the active word", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		mustTypeAll(t, typeTest, "cat")

		assert.Equal(t, []string{"cat"}, inputs(typeTest))
	})

	t.Run("should count every keystroke, including incorrect ones", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		mustTypeAll(t, typeTest, "cxt")

		assert.Equal(t, 3, typeTest.totalKeystrokes)
		assert.Equal(t, 2, typeTest.correctKeyStrokes)
	})

	t.Run("should move to the next word on space", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"the", "cat"}, WordsMode)
		mustTypeAll(t, typeTest, "the c")

		assert.Equal(t, 1, typeTest.ActiveWord())
		assert.Equal(t, []string{"the", "c"}, inputs(typeTest))
	})

	t.Run("should skip the rest of a word when space is typed early", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"the", "cat"}, WordsMode)
		mustTypeAll(t, typeTest, "th cat")

		assert.Equal(t, []string{"th", "cat"}, inputs(typeTest))
	})

	t.Run("should keep extra letters within their word", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"the", "cat"}, WordsMode)
		mustTypeAll(t, typeTest, "thee cat")

		assert.Equal(t, []string{"thee", "cat"}, inputs(typeTest))
	})

	t.Run("should ignore space at the start of a word", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"the", "cat"}, WordsMode)
		mustTypeAll(t, typeTest, " the  ")

		assert.Equal(t, []string{"the", ""}, inputs(typeTest))
		assert.Equal(t, 4, typeTest.totalKeystrokes)
	})

	t.Run("should not start the test on a leading space", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		typeTest.Type(' ')

		assert.False(t, typeTest.Started())
	})

	t.Run("should cap extra letters past the end of a word", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		mustTypeAll(t, typeTest, "cat"+strings.Repeat("x", MaxExtraChars+5))

		assert.Len(t, typeTest.Inputs[0], 3+MaxExtraChars)
	})

	t.Run("should count committing the last word as a missed keystroke", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"ab"}, TimeMode)
		mustTypeAll(t, typeTest, "ab ")

		assert.Equal(t, 3, typeTest.totalKeystrokes)
		assert.Equal(t, 2, typeTest.correctKeyStrokes)
	})

	t.Run("should ignore input once every word is committed", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"ab"}, TimeMode)
		mustTypeAll(t, typeTest, "ab cd")

		assert.Equal(t, []string{"ab", ""}, inputs(typeTest))
	})
}

func TestTypingTest_Backspace(t *testing.T) {
	cases := []struct {
		name  string
		typed string
		want  []string
	}{
		{
			name:  "should delete the last rune of the active word",
			typed: "the ca",
			want:  []string{"the", "c"},
		},
		{
			name:  "should not go back into a correct word",
			typed: "the ",
			want:  []string{"the", ""},
		},
		{
			name:  "should go back into an incorrect word with its input intact",
			typed: "thx ",
			want:  []string{"thx"},
		},
		{
			name:  "should go back into a word that was committed early",
			typed: "th ",
			want:  []string{"th"},
		},
		{
			name:  "should be no-op on the first word when empty",
			typed: "",
			want:  []string{""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typeTest := NewTypeTest([]string{"the", "cat"}, WordsMode)
			mustTypeAll(t, typeTest, tc.typed)

			typeTest.Backspace()
			assert.Equal(t, tc.want, inputs(typeTest))
		})
	}
}

func TestTypingTest_DeleteWord(t *testing.T) {
	cases := []struct {
		name  string
		typed string
		want  []string
	}{
		{
			name:  "should clear the active word",
			typed: "the ca",
			want:  []string{"the", ""},
		},
		{
			name:  "should not clear a correct previous word",
			typed: "the ",
			want:  []string{"the", ""},
		},
		{
			name:  "should go back and clear an incorrect previous word",
			typed: "thx ",
			want:  []string{""},
		},
		{
			name:  "should be no-op on empty",
			typed: "",
			want:  []string{""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typeTest := NewTypeTest([]string{"the", "cat"}, WordsMode)
			mustTypeAll(t, typeTest, tc.typed)

			typeTest.DeleteWord()
			assert.Equal(t, tc.want, inputs(typeTest))
		})
	}
}

func TestTypeTest_Completed(t *testing.T) {
	t.Run("when in words mode", func(t *testing.T) {
		cases := []struct {
			name  string
			typed string
			want  bool
		}{
			{
				name:  "then it is not completed while the last word is partial",
				typed: "ab c",
				want:  false,
			},
			{
				name:  "then it is not completed when the last word is full length but wrong",
				typed: "ab cx",
				want:  false,
			},
			{
				name:  "then it is not completed when an earlier word is typed correctly",
				typed: "ab",
				want:  false,
			},
			{
				name:  "then it is completed once the last word is typed correctly",
				typed: "ab cd",
				want:  true,
			},
			{
				name:  "then it is completed when the last word is committed with mistakes",
				typed: "ab cx ",
				want:  true,
			},
			{
				name:  "then it is completed despite mistakes in earlier words",
				typed: "xx cd",
				want:  true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				typeTest := NewTypeTest([]string{"ab", "cd"}, WordsMode)
				mustTypeAll(t, typeTest, tc.typed)

				assert.Equal(t, tc.want, typeTest.Completed())
			})
		}
	})

	t.Run("when in time mode", func(t *testing.T) {
		t.Run("then it should return true when finished", func(t *testing.T) {
			typeTest := NewTypeTest([]string{"cat"}, TimeMode)
			typeTest.Finish()
			assert.True(t, typeTest.Completed())
		})
	})
}

func TestTypeTest_Elapsed(t *testing.T) {
	t.Run("should not start the block until the first keystroke", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, TimeMode)
		assert.False(t, typeTest.Started())
		typeTest.Type('c')
		assert.True(t, typeTest.Started())
	})

	t.Run("should measure elapsed time from the first keystroke", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, TimeMode)
		clock := clockwork.NewFakeClockAt(time.Unix(1000, 0))
		typeTest.clock = clock.Now

		typeTest.Type('c')

		duration := 3 * time.Second
		clock.Advance(duration)

		assert.Equal(t, duration, typeTest.Elapsed())
	})

	t.Run("should freeze elapsed time at the moment Finish is called", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, TimeMode)
		clock := clockwork.NewFakeClockAt(time.Unix(1000, 0))
		typeTest.clock = clock.Now

		typeTest.Type('c')

		duration := 3 * time.Second
		clock.Advance(duration)

		typeTest.Finish()
		clock.Advance(10 * time.Second)

		assert.Equal(t, duration, typeTest.Elapsed())
	})
}

func TestComputeStats(t *testing.T) {
	t.Run("should report WPM as correct characters over five per minute", func(t *testing.T) {
		target := strings.Repeat("a", 5)
		typeTest := mustTypedTestFinished(t, []string{target}, target)
		stats := ComputeStats(typeTest)

		assert.InDelta(t, 1.0, stats.WPM, 1e-6)
		assert.InDelta(t, 1.0, stats.RawWPM, 1e-6)
		assert.Equal(t, 5, stats.CorrectChars)
	})

	t.Run("should scale WPM with characters typed per minute", func(t *testing.T) {
		target := strings.Repeat("a", 50)
		typeTest := mustTypedTestFinished(t, []string{target}, target)
		stats := ComputeStats(typeTest)

		assert.InDelta(t, 10.0, stats.WPM, 1e-6)
		assert.InDelta(t, 10.0, stats.RawWPM, 1e-6)
		assert.Equal(t, 50, stats.CorrectChars)
	})

	t.Run("should compute with incorrect typed characters", func(t *testing.T) {
		target := strings.Repeat("a", 3)
		typeTest := mustTypedTestFinished(t, []string{target}, "aXa")
		stats := ComputeStats(typeTest)

		want := 2.0 / 3.0 * 100.0
		assert.InDelta(t, want, stats.Accuracy, 1e-6)
		assert.Equal(t, 0, stats.CorrectChars)
		assert.Equal(t, 1, stats.IncorrectChars)
		assert.Equal(t, 0, stats.MissedChars)
		assert.Equal(t, 2, stats.CorrectKeystrokes)
		assert.Equal(t, 3, stats.TypedChars)
	})

	t.Run("should only count fully correct words toward WPM", func(t *testing.T) {
		typeTest := mustTypedTestFinished(t, []string{"ab", "cd", "ef"}, "ab cx ef")
		stats := ComputeStats(typeTest)

		assert.Equal(t, 5, stats.CorrectChars)
		assert.InDelta(t, 1.0, stats.WPM, 1e-6)
		assert.Equal(t, 8, stats.TypedChars)
		assert.InDelta(t, 8.0/5.0, stats.RawWPM, 1e-6)
	})

	t.Run("should count letters skipped by an early space as missed", func(t *testing.T) {
		typeTest := mustTypedTestFinished(t, []string{"abc", "de"}, "a de")
		stats := ComputeStats(typeTest)

		assert.Equal(t, 2, stats.CorrectChars)
		assert.Equal(t, 1, stats.IncorrectChars)
		assert.Equal(t, 2, stats.MissedChars)
		assert.Equal(t, 4, stats.TypedChars)
	})

	t.Run("should count letters typed past a word's end as extra", func(t *testing.T) {
		typeTest := mustTypedTestFinished(t, []string{"ab", "cd"}, "abxy cd")
		stats := ComputeStats(typeTest)

		assert.Equal(t, 2, stats.CorrectChars)
		assert.Equal(t, 2, stats.ExtraChars)
		assert.Equal(t, 1, stats.IncorrectChars)
	})

	t.Run("should report 100% accuracy when no keys has been pressed", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, TimeMode)
		stats := ComputeStats(typeTest)
		assert.InDelta(t, 100.0, stats.Accuracy, 1e-6)
	})
}

func TestCountChars(t *testing.T) {
	cases := []struct {
		name          string
		input         string
		target        string
		creditPartial bool
		want          charCounts
	}{
		{
			name:   "should credit every char of an exact word",
			input:  "the ",
			target: "the ",
			want:   charCounts{allCorrect: 4, correctWord: 4},
		},
		{
			name:   "should count a mistyped word's space as extra and credit nothing",
			input:  "thx ",
			target: "the ",
			want:   charCounts{allCorrect: 2, incorrect: 1, extra: 1},
		},
		{
			name:   "should count an early space as incorrect and the rest as missed",
			input:  "th ",
			target: "the ",
			want:   charCounts{allCorrect: 2, incorrect: 1, missed: 1},
		},
		{
			name:   "should count letters past the end as incorrect then extra",
			input:  "thee ",
			target: "the ",
			want:   charCounts{allCorrect: 3, incorrect: 1, extra: 1},
		},
		{
			name:          "should credit a correct prefix when partial credit is on",
			input:         "th",
			target:        "the ",
			creditPartial: true,
			want:          charCounts{allCorrect: 2, correctWord: 2},
		},
		{
			name:          "should not credit a wrong prefix even with partial credit",
			input:         "tx",
			target:        "the ",
			creditPartial: true,
			want:          charCounts{allCorrect: 1, incorrect: 1},
		},
		{
			name:          "should count letters typed over the space as extra",
			input:         "thex",
			target:        "the ",
			creditPartial: true,
			want:          charCounts{allCorrect: 3, extra: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := countChars([]rune(tc.input), []rune(tc.target), tc.creditPartial)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConsistency(t *testing.T) {
	t.Run("should score uniform sample as 100", func(t *testing.T) {
		consistency := Consistency([]float64{40, 40, 40, 40})
		assert.InDelta(t, 100.0, consistency, 1e-6)
	})

	t.Run("should score varied samples strictly between 0 and 100", func(t *testing.T) {
		consistency := Consistency([]float64{10, 90, 20, 80})
		assert.Greater(t, consistency, 0.0)
		assert.Less(t, consistency, 100.0)
	})

	t.Run("should score close to 100*(1-cov) when variation is small", func(t *testing.T) {
		consistency := Consistency([]float64{40, 60})
		assert.InDelta(t, 80.0, consistency, 0.001)
	})

	t.Run("should stay above 0 when variation exceeds the mean", func(t *testing.T) {
		consistency := Consistency([]float64{0, 0, 0, 100})
		assert.Greater(t, consistency, 0.0)
	})

	t.Run("should score a single sample as 100", func(t *testing.T) {
		assert.InDelta(t, 100.0, Consistency([]float64{40}), 1e-6)
	})

	t.Run("should score no samples as 0", func(t *testing.T) {
		assert.InDelta(t, 0.0, Consistency(nil), 1e-6)
	})
}

func TestIsAFK(t *testing.T) {
	cases := []struct {
		name  string
		afk   int
		total int
		want  bool
	}{
		{
			name:  "should flag a test idle the entire time",
			afk:   10,
			total: 10,
			want:  true,
		},
		{
			name:  "should flag a test idle for half",
			afk:   5,
			total: 10,
			want:  true,
		},
		{
			name:  "should not flag a test active past threshold",
			afk:   4,
			total: 10,
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsAFK(tc.afk, tc.total))
		})
	}
}
