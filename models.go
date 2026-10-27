package stdtype

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

type Mode int

const (
	TimeMode Mode = iota
	WordsMode
	InvalidMode
)

type SettingField string

const (
	SettingMode        SettingField = "MODE"
	SettingDuration    SettingField = "DURATION"
	SettingWordCount   SettingField = "WORDS"
	SettingList        SettingField = "LIST"
	SettingPunctuation SettingField = "PUNCTUATION"
	SettingNumbers     SettingField = "NUMBERS"
)

func (m Mode) String() string {
	if m == WordsMode {
		return "words"
	}

	return "time"
}

func ParseMode(m string) Mode {
	switch m {
	case TimeMode.String():
		return TimeMode
	case WordsMode.String():
		return WordsMode
	default:
		return InvalidMode
	}
}

// Stats is a snapshot of performance for a test, computed for its current
// state and elapsed time.
type Stats struct {
	WPM      float64 `json:"wpm"`
	RawWPM   float64 `json:"raw_wpm"`
	Accuracy float64 `json:"accuracy"`

	CorrectChars   int `json:"correct_chars,omitempty"`
	IncorrectChars int `json:"incorrect_chars,omitempty"`
	ExtraChars     int `json:"extra_chars,omitempty"`
	MissedChars    int `json:"missing_chars,omitempty"`

	TypedChars        int `json:"typed_chars,omitempty"`
	Keystrokes        int `json:"key_strokes,omitempty"`
	CorrectKeystrokes int `json:"correct_keystrokes,omitempty"`
}

// Samples are taken once per second.
type Samples struct {
	WPM []float64 `json:"wpm,omitempty"`
	Raw []float64 `json:"raw,omitempty"`
	Err []int     `json:"err,omitempty"`
}

// Result is a persisted record of one complete test.
type Result struct {
	Timestamp   time.Time `json:"timestamp"`
	Mode        string    `json:"mode"`
	Amount      int       `json:"amount"` // seconds (time mode) or word count
	Stats       Stats     `json:"stats"`
	Consistency float64   `json:"consistency"`

	Invalid       bool          `json:"invalid,omitempty"`
	InvalidReason string        `json:"invalid_reason,omitempty"`
	Samples       Samples       `json:"samples"`
	Duration      time.Duration `json:"duration"`
}

type TypeTest struct {
	Words  [][]rune
	Inputs [][]rune
	Mode   Mode

	started    bool
	startedAt  time.Time
	finished   bool
	finishedAt time.Time

	correctKeyStrokes int
	totalKeystrokes   int

	clock func() time.Time
}

type Settings struct {
	Mode        string `mapstructure:"mode"`
	Duration    int    `mapstructure:"duration"`
	WordCount   int    `mapstructure:"words"`
	Lines       int    `mapstructure:"lines"`
	Theme       string `mapstructure:"theme"`
	WordList    string `mapstructure:"list"`
	Punctuation bool   `mapstructure:"punctuation"`
	Numbers     bool   `mapstructure:"numbers"`
}

func DefaultSettings() Settings {
	return Settings{
		Mode:      TimeMode.String(),
		Duration:  30,
		WordCount: 25,
		Lines:     3,
		Theme:     "default",
		WordList:  "english",
	}
}

var (
	ErrInvalidMode      = errors.New("invalid mode")
	ErrInvalidDuration  = errors.New("invalid duration")
	ErrInvalidWordCount = errors.New("invalid words")
	ErrInvalidLines     = errors.New("invalid lines")
	ErrInvalidList      = errors.New("invalid list")
)

func (s Settings) Validate(lists []string) error {
	var errs []error

	if ParseMode(s.Mode) == InvalidMode {
		errs = append(errs, fmt.Errorf("%w %q: must be %q or %q", ErrInvalidMode, s.Mode, TimeMode, WordsMode))
	}

	if s.Duration <= 0 {
		errs = append(errs, fmt.Errorf("%w %d: must be positive", ErrInvalidDuration, s.Duration))
	}

	if s.WordCount <= 0 {
		errs = append(errs, fmt.Errorf("%w %d: must be positive", ErrInvalidWordCount, s.WordCount))
	}

	if s.Lines <= 0 {
		errs = append(errs, fmt.Errorf("%w %d: must be positive", ErrInvalidLines, s.Lines))
	}

	if !slices.Contains(lists, s.WordList) {
		errs = append(errs, fmt.Errorf("%w %q: must be one of %v", ErrInvalidList, s.WordList, lists))
	}

	return errors.Join(errs...)
}
