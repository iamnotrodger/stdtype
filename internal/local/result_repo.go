package local

import (
	"encoding/json"
	"os"
	"path/filepath"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/pkg/errors"
)

var _ stdtype.ResultRepo = (*ResultRepo)(nil)

type ResultRepoOption func(*resultRepoOptions)

func WithFile(file string) ResultRepoOption {
	return func(o *resultRepoOptions) {
		o.file = file
	}
}

type resultRepoOptions struct {
	file string
}

func newResultRepoOptions(opts ...ResultRepoOption) *resultRepoOptions {
	o := &resultRepoOptions{
		file: "results.json",
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type ResultRepo struct {
	dir  string
	file string
}

func NewResultRepo(dir string, opts ...ResultRepoOption) (*ResultRepo, error) {
	options := newResultRepoOptions(opts...)
	path := filepath.Join(dir, options.file)

	return &ResultRepo{dir, path}, nil
}

func (h *ResultRepo) ListResults() ([]stdtype.Result, error) {
	data, err := os.ReadFile(h.file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "os.ReadFile")
	}

	var results []stdtype.Result
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return results, nil
}

func (h *ResultRepo) CreateResult(result stdtype.Result) error {
	results, err := h.ListResults()
	if err != nil {
		return errors.Wrap(err, "ListResults")
	}

	results = append(results, result)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return errors.Wrap(err, "json.MarshalIndent")
	}

	if err := writeFileWithTemp(h.dir, h.file, data); err != nil {
		return errors.Wrap(err, "writeFileWithTemp")
	}

	return nil
}
