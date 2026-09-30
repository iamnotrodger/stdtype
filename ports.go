package stdtype

type HistoryRepo interface {
	Load() ([]Result, error)
	Save(r Result) error
}

type SettingsRepo interface {
	Load() (Settings, error)
	Save(s Settings) error
}

type GenerateOptions struct {
	List        string
	Punctuation bool
	Numbers     bool
}

type WordSource interface {
	// Generate returns n words for a new test, shapted by opts.
	Generate(n int, opts GenerateOptions) []string
	// Lists returns the available word-list ids, in display order.
	Lists() []string
}
