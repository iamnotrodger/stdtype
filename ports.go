package stdtype

type StatsRepo interface {
	ListResults() ([]Result, error)
	CreateResult(r Result) error
}

type SettingsRepo interface {
	GetSettings() (Settings, error)
	UpdateSettings(s Settings) error
}

type GenerateOptions struct {
	List        string
	Punctuation bool
	Numbers     bool
}

type WordSource interface {
	// Generate returns n words for a new test, shaped by opts.
	Generate(n int, opts GenerateOptions) ([]string, error)
	// Lists returns the available word-list ids, in display order.
	Lists() []string
}
