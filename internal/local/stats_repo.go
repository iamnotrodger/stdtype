package local

import (
	"encoding/json"
	"os"
	"path/filepath"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/pkg/errors"
)

var _ stdtype.StatsRepo = (*StatsRepo)(nil)

type StatsRepo struct {
	dir  string
	file string
}

func NewStatsRepo(dir string, file string) (*StatsRepo, error) {
	return &StatsRepo{
		dir:  dir,
		file: filepath.Join(dir, file),
	}, nil
}

func (s *StatsRepo) ListResults() ([]stdtype.Result, error) {
	data, err := os.ReadFile(s.file)
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

func (s *StatsRepo) CreateResult(result stdtype.Result) error {
	results, err := s.ListResults()
	if err != nil {
		return errors.Wrap(err, "ListResults")
	}

	results = append(results, result)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return errors.Wrap(err, "json.MarshalIndent")
	}

	if err := writeFileWithTemp(s.dir, s.file, data); err != nil {
		return errors.Wrap(err, "writeFileWithTemp")
	}

	return nil
}
