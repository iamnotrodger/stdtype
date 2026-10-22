package stdtype

import "time"

type Mode int

const (
	TimeMode Mode = iota
	WordsMode
	InvalidMode
)

type SettingField string

const (
	SettingMode      SettingField = "MODE"
	SettingDuration  SettingField = "DURATION"
	SettingWordCount SettingField = "WORDS"
	SettingList      SettingField = "LIST"
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
	MissedChars    int `json:"missing_chars,omitempty"`

	TypedChars        int `json:"typed_chars,omitempty"`
	Keystrokes        int `json:"key_strokes,omitempty"`
	CorrectKeystrokes int `json:"correct_keystrokes,omitempty"`
}

// Per-second WPM samples (chars/5 per 1/60 min).
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
	Target []rune
	Typed  []rune
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
