package stdtype

import "time"

type Mode int

const (
	TimeMode Mode = iota
	WordsMode
)

func (m Mode) String() string {
	if m == WordsMode {
		return "words"
	}

	return "time"
}

// Result is a persisted record of one complete test.
type Result struct {
	Timestamp   time.Time `json:"timestamp"`
	Mode        string    `json:"mode"`
	Amount      int       `json:"amount"` // seconds (time mode) or word count
	WPM         float64   `json:"wpm"`
	RawWPM      float64   `json:"raw_wpm"`
	Accuracy    float64   `json:"accuracy"`
	Consistency float64   `json:"consistency"`

	CorrectChars   int `json:"correct_chars,omitempty"`
	IncorrectChars int `json:"incorrect_chars,omitempty"`
	MissedChars    int `json:"missing_chars,omitempty"`

	Invalid       bool   `json:"invalid,omitempty"`
	InvalidReason string `json:"invalid_reason,omitempty"`
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

// Stats is a snapshot of performance for a test, compuated for its current
// state and elapsed time.
type Stats struct {
	WPM          float64
	RawWPM       float64
	Accuracy     float64
	Consistency  float64
	CorrectChars int
	TypedChars   int
	Keystrokes   int

	IncorrectChars    int
	MissedChars       int
	CorrectKeystrokes int
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
