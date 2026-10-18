package commandpalette

const (
	defaultMaxResults  = 10
	defaultMaxWidth    = 60
	defaultPlaceholder = "search..."
	promptStr          = "> "
)

type option struct {
	placeholder string
	maxResults  int
	maxWidth    int
}

type Option func(*option)

func WithPlaceholder(placeholder string) Option {
	return func(o *option) {
		o.placeholder = placeholder
	}
}

func WithMaxResult(result int) Option {
	return func(o *option) {
		o.maxResults = result
	}
}

func WithMaxWidth(width int) Option {
	return func(o *option) {
		o.maxWidth = width
	}
}

func newOptions(opts ...Option) *option {
	o := &option{
		placeholder: defaultPlaceholder,
		maxResults:  defaultMaxResults,
		maxWidth:    defaultMaxWidth,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
