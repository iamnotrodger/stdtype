package local

import (
	"encoding/json"
	"os"
	"path/filepath"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/pkg/errors"
)

var _ stdtype.ResultRepo = (*ResultRepo)(nil)

type ResultRepo struct {
	dir  string
	file string
}

func NewResultRepo(dir string, file string) (*ResultRepo, error) {
	return &ResultRepo{
		dir:  dir,
		file: filepath.Join(dir, file),
	}, nil
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
