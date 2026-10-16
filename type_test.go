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
	duration time.Duration,
) *TypeTest {
	tb.Helper()

	typeTest := NewTypeTest(target, WordsMode)
	clock := clockwork.NewFakeClockAt(time.Unix(1000, 0))
	typeTest.clock = clock.Now

	mustTypeAll(tb, typeTest, typed)

	clock.Advance(duration)
	typeTest.Finish()

	return typeTest
}

func TestNewTypingTest(t *testing.T) {
	t.Run("should join the target words with single spaces", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"the", "quick", "fox"}, WordsMode)
		assert.Equal(t, "the quick fox", string(typeTest.Target))
	})
}

func TestTypingTest_Type(t *testing.T) {
	t.Run("should append typed runes and track correct chracters", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		mustTypeAll(t, typeTest, "cat")

		assert.Equal(t, "cat", string(typeTest.Typed))
		assert.Equal(t, 3, typeTest.CorrectChars())
	})

	t.Run("should count every keystroke, including incorrect ones", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		mustTypeAll(t, typeTest, "cxt")

		assert.Equal(t, 3, typeTest.totalKeystrokes)
		assert.Equal(t, 2, typeTest.correctKeyStrokes)
	})

	t.Run("should not accept input past the end of the target", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, WordsMode)
		mustTypeAll(t, typeTest, "catcat")

		assert.Len(t, typeTest.Typed, 3)
	})
}

func TestTypingTest_DeleteWord(t *testing.T) {
	cases := []struct {
		name  string
		typed string
		want  string
	}{
		{
			name:  "should delete a partial trailing word",
			typed: "hello wor",
			want:  "hello ",
		},
		{
			name:  "should delete a trailing space and word",
			typed: "hello ",
			want:  "",
		},
		{
			name:  "should delete the only word",
			typed: "hello",
			want:  "",
		},
		{
			name:  "should be no-op on empty",
			typed: "",
			want:  "",
		},
		{
			name:  "should only delete the last of many words",
			typed: "one two three",
			want:  "one two ",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typeTest := NewTypeTest([]string{}, WordsMode)
			typeTest.Typed = []rune(tc.typed)

			typeTest.DeleteWord()
			assert.Equal(t, tc.want, string(typeTest.Typed))
		})
	}
}

func TestTypeTest_Completed(t *testing.T) {
	t.Run("when in words mode", func(t *testing.T) {
		t.Run("then it is not completed before the target is fully typed", func(t *testing.T) {
			typeTest := NewTypeTest([]string{"cat"}, WordsMode)
			mustTypeAll(t, typeTest, "ca")
			assert.False(t, typeTest.Completed())
		})

		t.Run("then it is completed once the whole target is typed", func(t *testing.T) {
			typeTest := NewTypeTest([]string{"cat"}, WordsMode)
			mustTypeAll(t, typeTest, "cat")
			assert.True(t, typeTest.Completed())
		})
	})

	t.Run("when in time mode", func(t *testing.T) {
		t.Run("then it should return true when finished", func(t *testing.T) {
			typeTest := NewTypeTest([]string{"cat"}, TimeMode)
			typeTest.Finish()
			assert.False(t, typeTest.Completed())
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
		typeTest := mustTypedTestFinished(t, []string{target}, target, time.Minute)
		stats := ComputeStats(typeTest)

		assert.InDelta(t, 1.0, stats.WPM, 1e-6)
		assert.InDelta(t, 1.0, stats.RawWPM, 1e-6)
		assert.Equal(t, 5, stats.CorrectChars)
	})

	t.Run("should scale WPM with characters typed per minute", func(t *testing.T) {
		target := strings.Repeat("a", 50)
		typeTest := mustTypedTestFinished(t, []string{target}, target, time.Minute)
		stats := ComputeStats(typeTest)

		assert.InDelta(t, 10.0, stats.WPM, 1e-6)
		assert.InDelta(t, 10.0, stats.RawWPM, 1e-6)
		assert.Equal(t, 50, stats.CorrectChars)
	})

	t.Run("should compute with incorrect typed characters", func(t *testing.T) {
		target := strings.Repeat("a", 3)
		typeTest := mustTypedTestFinished(t, []string{target}, "aXa", time.Minute)
		stats := ComputeStats(typeTest)

		want := 2.0 / 3.0 * 100.0
		assert.InDelta(t, want, stats.Accuracy, 1e-6)
		assert.Equal(t, 2, stats.CorrectChars)
		assert.Equal(t, 1, stats.IncorrectChars)
		assert.Equal(t, 0, stats.MissedChars)
		assert.Equal(t, 2, stats.CorrectKeystrokes)
		assert.Equal(t, 3, stats.TypedChars)
	})

	t.Run("should report 100% accuracy when no keys has been pressed", func(t *testing.T) {
		typeTest := NewTypeTest([]string{"cat"}, TimeMode)
		stats := ComputeStats(typeTest)
		assert.InDelta(t, 100.0, stats.Accuracy, 1e-6)
	})
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

	t.Run("should score by coefficienty of variation", func(t *testing.T) {
		consistency := Consistency([]float64{40, 60})
		assert.InDelta(t, 80.0, consistency, 1e-6)
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
