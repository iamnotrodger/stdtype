package words

import (
	"cmp"
	"embed"
	"encoding/json"
	"maps"
	"math"
	"math/rand"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/pkg/errors"
)

var _ stdtype.WordSource = (*WordSource)(nil)

const (
	// defaultList  is used when options name no list or an unknown one.
	defaultList = "english"

	// numberChance is the per-word probability of injecting a number in number mode.
	numberChance = 0.1
)

var (
	//go:embed lists/*.json
	listFS embed.FS

	punctStripper = strings.NewReplacer(
		".", "",
		"?", "",
		"!", "",
		`"`, "",
		":", "",
		"-", "",
		",", "",
	)
)

var contractions = map[string][]string{
	"are":    {"aren't"},
	"can":    {"can't"},
	"could":  {"couldn't"},
	"did":    {"didn't"},
	"does":   {"doesn't"},
	"do":     {"don't"},
	"had":    {"hadn't"},
	"has":    {"hasn't"},
	"have":   {"haven't"},
	"is":     {"isn't"},
	"it":     {"it's", "it'll"},
	"i":      {"i'm", "i'll", "i've", "i'd"},
	"you":    {"you'll", "you're", "you've", "you'd"},
	"that":   {"that's", "that'll", "that'd"},
	"must":   {"mustn't", "must've"},
	"there":  {"there's", "there'll", "there'd"},
	"he":     {"he's", "he'll", "he'd"},
	"she":    {"she's", "she'll", "she'd"},
	"we":     {"we're", "we'll", "we'd"},
	"they":   {"they're", "they'll", "they'd"},
	"should": {"shouldn't", "should've"},
	"was":    {"wasn't"},
	"were":   {"weren't"},
	"will":   {"won't"},
	"would":  {"wouldn't", "would've"},
	"going":  {"goin'"},
}

type WordList struct {
	Name  string   `json:"name"`
	Words []string `json:"words"`
}

type WordSource struct {
	lists map[string]string
	ids   []string
	rng   *rand.Rand
}

func NewWordSource(seed int64) (*WordSource, error) {
	lists := make(map[string]string)

	dir := "lists"
	entries, err := listFS.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "listFS.ReadDir")
	}

	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		name := strings.TrimSuffix(e.Name(), ".json")

		lists[name] = path
	}

	// defaultList first, then ascending size.
	ids := slices.Collect(maps.Keys(lists))
	slices.SortFunc(ids, func(a, b string) int {
		if a == defaultList {
			return -1
		}
		if b == defaultList {
			return 1
		}

		return cmp.Compare(a, b)
	})

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec

	return &WordSource{lists, ids, rng}, nil
}

func (w *WordSource) Lists() []string {
	out := make([]string, len(w.ids))
	copy(out, w.ids)

	return out
}

// GetList defaults to "english" if list does not exist.
func (w *WordSource) GetList(list string) (*WordList, error) {
	file, ok := w.lists[list]
	if !ok {
		file = w.lists[defaultList]
	}

	data, err := listFS.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "listFS.ReadFile")
	}

	var wordList WordList
	if err := json.Unmarshal(data, &wordList); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return &wordList, nil
}

func (w *WordSource) Generate(n int, opts stdtype.GenerateOptions) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	wordList, err := w.GetList(opts.List)
	if err != nil {
		return nil, errors.Wrap(err, "w.GetList")
	}

	out := make([]string, 0, n)
	for range n {
		if opts.Numbers && w.rng.Float64() < numberChance {
			out = append(out, w.number())

			continue
		}

		out = append(out, w.word(wordList.Words, out))
	}

	if opts.Punctuation {
		w.punctuate(out)
	}

	return out, nil
}

func (w *WordSource) word(words, prev []string) string {
	prevWord1 := normalize(last(prev, 0))
	prevWord2 := normalize(last(prev, 1))

	isRepeat := func(w string) bool {
		k := normalize(w)

		return k == prevWord1 || k == prevWord2
	}

	word := w.randomWord(words)

	if len(words) <= 2 {
		return word
	}

	for retries := 0; retries < 100 && isRepeat(word); retries++ {
		word = w.randomWord(words)
	}

	return word
}

func (w *WordSource) randomWord(words []string) string {
	return words[w.rng.Intn(len(words))]
}

// number returns a 1 to 4 digit number with a non-zero leading digit.
func (w *WordSource) number() string {
	digits := 1 + w.rng.Intn(4)
	low := int(math.Pow10((digits - 1)))
	high := int(math.Pow10(digits))

	return strconv.Itoa(low + w.rng.Intn(high-low))
}

func (w *WordSource) punctuate(words []string) {
	for i := range words {
		prev := ""
		if i > 0 {
			prev = words[i-1] //nolint:gosec
		}

		words[i] = w.punctuateWord(prev, words[i], i, len(words))
	}
}

// punctuateWord capitalizes sentences start, ends sentences with terminator, and
// occasionally attaches a quote, bracket, colon, semicolon, comma, or dash.
func (w *WordSource) punctuateWord(prev, word string, index, maxIndex int) string {
	isLast := index == maxIndex-1
	lastChar := lastRune(prev)

	if index == 0 || lastChar == '.' || lastChar == '?' || lastChar == '!' {
		return capitalize(word)
	}

	// Never end on a one-word sentence.
	endSentence := w.rng.Float64() < 0.1 && lastChar != ',' && index != maxIndex-2
	if isLast || endSentence {
		return word + w.terminator()
	}

	return w.decorate(prev, word)
}

// terminator returns a sentence-ending marker, weighted towards the period.
func (w *WordSource) terminator() string {
	switch x := w.rng.Float64(); {
	case x <= 0.8:
		return "."
	case x <= 0.9:
		return "?"
	default:
		return "!"
	}
}

func (w *WordSource) decorate(prev, word string) string {
	r := w.rng
	last := lastRune(prev)

	switch {
	case r.Float64() < 0.01 && last != ',':
		return `"` + word + `"`
	case r.Float64() < 0.011 && last != ',':
		return `'` + word + `'`
	case r.Float64() < 0.012 && last != ',':
		return `(` + word + `)`
	case r.Float64() < 0.013 && last != ',' && last != ';' && last != ':':
		return word + `:`
	case r.Float64() < 0.014 && last != ',' && prev != "-":
		return "-"
	case r.Float64() < 0.015 && last != ',' && last != ';':
		return word + `;`
	case r.Float64() < 0.2 && last != ',':
		return word + `,`
	case r.Float64() < 0.5:
		return w.contract(word)
	}

	return word
}

func (w *WordSource) contract(word string) string {
	options, ok := contractions[strings.ToLower(word)]
	if !ok {
		return word
	}

	c := options[w.rng.Intn(len(options))]
	if unicode.IsUpper([]rune(word)[0]) {
		return capitalize(c)
	}

	return c
}

// normalize returns word without punctuation and capitalization.
func normalize(w string) string {
	return strings.ToLower(punctStripper.Replace(w))
}

func capitalize(w string) string {
	if w == "" {
		return w
	}

	runes := []rune(w)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

func last(s []string, back int) string {
	if i := len(s) - 1 - back; i >= 0 {
		return s[i]
	}

	return ""
}

func lastRune(s string) rune {
	r := []rune(s)
	if len(r) == 0 {
		return 0
	}

	return r[len(r)-1]
}
