package words

import (
	"regexp"
	"strings"
	"testing"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	english    = stdtype.GenerateOptions{List: "english"}
	digitCheck = regexp.MustCompile(`^[0-9]+$`)
)

func TestWordSource_Lists(t *testing.T) {
	t.Run("should include english list in order", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		want := []string{"english", "english_1k", "english_5k"}
		got := wordSource.Lists()

		assert.Equal(t, want, got)
	})
}

func TestWordSource_GetList(t *testing.T) {
	t.Run("should default to english when list does not exist", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		want, err := wordSource.GetList(defaultList)
		require.NoError(t, err)

		got, err := wordSource.GetList("not exist")
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("should return given list", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		got, err := wordSource.GetList("english_1k")
		require.NoError(t, err)

		assert.NotNil(t, got)
		assert.NotEmpty(t, got.Words)
	})
}

func TestWordSource_Generate(t *testing.T) {
	t.Run("should generate requested count", func(t *testing.T) {
		cases := []struct {
			name  string
			count int
		}{
			{
				name:  "should return nil on negative int",
				count: -1,
			},
			{
				name:  "should generate zero",
				count: 0,
			},
			{
				name:  "should generate 1 word",
				count: 1,
			},
			{
				name:  "should generate 1 word",
				count: 10,
			},
			{
				name:  "should generate 1 word",
				count: 50,
			},
			{
				name:  "should generate 1 word",
				count: 100,
			},
		}

		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		for _, tc := range cases {
			got, err := wordSource.Generate(tc.count, english)
			require.NoError(t, err)

			if tc.count <= 0 {
				assert.Nil(t, got)
			} else {
				assert.Len(t, got, tc.count)
			}
		}
	})

	t.Run("should be deterministic with same seed", func(t *testing.T) {
		a, err := NewWordSource(1)
		require.NoError(t, err)

		b, err := NewWordSource(1)
		require.NoError(t, err)

		wordsA, err := a.Generate(20, english)
		require.NoError(t, err)

		wordsB, err := b.Generate(20, english)
		require.NoError(t, err)

		assert.Equal(t, wordsA, wordsB)
	})

	t.Run("should generate different words for different seeds", func(t *testing.T) {
		a, err := NewWordSource(1)
		require.NoError(t, err)

		b, err := NewWordSource(2)
		require.NoError(t, err)

		wordsA, err := a.Generate(20, english)
		require.NoError(t, err)

		wordsB, err := b.Generate(20, english)
		require.NoError(t, err)

		assert.NotEqual(t, wordsA, wordsB)
	})

	t.Run("should not generate consecutive duplicates", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		words, err := wordSource.Generate(100, english)
		require.NoError(t, err)

		for i := 1; i < len(words); i++ {
			assert.NotEqual(t, words[i-1], words[i])
		}
	})

	t.Run("should only generate words from list", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		words, err := wordSource.Generate(100, english)
		require.NoError(t, err)

		wordSet := mustBuildWordSet(t, words)

		for _, word := range words {
			assert.Contains(t, wordSet, word)
		}
	})

	t.Run("when punctuation is enabled then it should contain punctuations", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		opts := stdtype.GenerateOptions{List: "english", Punctuation: true}
		words, err := wordSource.Generate(50, opts)
		require.NoError(t, err)

		passage := strings.Join(words, " ")
		assert.True(t, strings.ContainsAny(passage, ".?!"))
	})

	t.Run("when punctuation is disabled then it should not contain punctuations", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		opts := stdtype.GenerateOptions{List: "english", Punctuation: false}
		words, err := wordSource.Generate(50, opts)
		require.NoError(t, err)

		passage := strings.Join(words, " ")
		assert.False(t, strings.ContainsAny(passage, ".?!"))
	})

	t.Run("when number is enabled then it should contain numerics", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		opts := stdtype.GenerateOptions{List: "english", Numbers: true}
		words, err := wordSource.Generate(50, opts)
		require.NoError(t, err)

		numNumeric := 0
		for _, w := range words {
			if digitCheck.MatchString(w) {
				numNumeric++
			}
		}

		assert.NotZero(t, numNumeric)
	})

	t.Run("when number is disabled then it should not contain numerics", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		opts := stdtype.GenerateOptions{List: "english", Numbers: false}
		words, err := wordSource.Generate(50, opts)
		require.NoError(t, err)

		numNumeric := 0
		for _, w := range words {
			if digitCheck.MatchString(w) {
				numNumeric++
			}
		}

		assert.Zero(t, numNumeric)
	})
}

func TestWordSource_Word(t *testing.T) {
	t.Run("should avoid repeating previous words", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		words := []string{"a", "b", "c", "d"}
		prev := []string{"e", "f"}

		for range 200 {
			word := wordSource.word(words, prev)
			assert.NotEqual(t, "e", word)
			assert.NotEqual(t, "f", word)
		}
	})
}

func TestWordSource_Generate_Numbers(t *testing.T) {
	t.Run("should produce a 1-4 digit number with a non-zero leading digit", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		for range 200 {
			num := wordSource.number()
			digitCheck.MatchString(num)
		}
	})
}

func TestWordSource_Generate_Decorate(t *testing.T) {
	t.Run("should produce a valid decorator or leave the word unchanged", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		for range 1000 {
			got := wordSource.decorate("prev", "word")
			assert.Contains(t, mustDecorate("word"), got)
		}
	})

	t.Run("when the previous word is a comma then should never attach a semicolon", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		for range 1000 {
			got := wordSource.decorate("prev,", "word")
			assert.NotEqual(t, "word;", got)
		}
	})
}

func TestWordSource_Generate_PunctuateWord(t *testing.T) {
	t.Run("should start sentences with capitalization", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		got := wordSource.punctuateWord("", "hello", 0, 1)
		assert.Equal(t, "Hello", got)
	})

	t.Run("should capitalize word after a terminator", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		terminators := []string{"end.", "end?", "end!"}
		for _, prev := range terminators {
			got := wordSource.punctuateWord(prev, "hello", 0, 1)
			assert.Equal(t, "Hello", got)
		}
	})

	t.Run("should end sentences with a terminator", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		terminators := []string{"end.", "end?", "end!"}
		got := wordSource.punctuateWord("prev", "end", 1, 2)
		assert.Contains(t, terminators, got)
	})
}

func TestWordSource_Generate_Punctuate(t *testing.T) {
	t.Run("should capitalize the first word and end the passage with a terminator", func(t *testing.T) {
		wordSource, err := NewWordSource(0)
		require.NoError(t, err)

		words := []string{"the", "quick", "brown", "fox", "jumps"}
		wordSource.punctuate(words)

		expected := []string{"The", "quick", "brown", "fox", "jumps."}
		assert.Equal(t, expected, words)
	})

	t.Run("should start a fresh sentence after a terminator with a capital", func(t *testing.T) {
		wordSource, err := NewWordSource(1)
		require.NoError(t, err)

		words := []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
		wordSource.punctuate(words)

		expected := []string{"The", "quick", "brown.", "Fox", "jumps", "over", "the", "lazy", "dog."}
		assert.Equal(t, expected, words)
	})

	t.Run("should leave an empty passage untouched", func(t *testing.T) {
		wordSource, err := NewWordSource(0)
		require.NoError(t, err)

		words := []string{}
		wordSource.punctuate(words)

		assert.Empty(t, words)
	})
}

func mustBuildWordSet(tb testing.TB, words []string) map[string]struct{} {
	tb.Helper()

	m := make(map[string]struct{})
	for _, w := range words {
		m[w] = struct{}{}
	}

	return m
}

func mustDecorate(word string) []string {
	return []string{
		word,
		`"` + word + `"`,
		`'` + word + `'`,
		`(` + word + `)`,
		word + `:`,
		word + `;`,
		word + `,`,
		word + " -",
	}
}
