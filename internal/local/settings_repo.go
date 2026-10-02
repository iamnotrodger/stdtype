package local

import (
	"encoding/json"
	"os"
	"path/filepath"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/pkg/errors"
)

var _ stdtype.SettingsRepo = (*SettingsRepo)(nil)

type SettingsRepo struct {
	dir  string
	file string
}

func NewSettingsRepo(dir string, file string) (*SettingsRepo, error) {
	return &SettingsRepo{
		dir:  dir,
		file: filepath.Join(dir, file),
	}, nil
}

func (s *SettingsRepo) GetSettings() (stdtype.Settings, error) {
	defaultSettings := stdtype.DefaultSettings()

	data, err := os.ReadFile(s.file)
	if errors.Is(err, os.ErrNotExist) {
		return defaultSettings, nil
	}

	if err != nil {
		return defaultSettings, errors.Wrap(err, "os.ReadFile")
	}

	if err := json.Unmarshal(data, &defaultSettings); err != nil {
		return defaultSettings, errors.Wrap(err, "json.Unmarshal")
	}

	return defaultSettings, nil
}

func (s *SettingsRepo) UpdateSettings(settings stdtype.Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return errors.Wrap(err, "json.MarshalIndent")
	}

	if err := writeFileWithTemp(s.dir, s.file, data); err != nil {
		return errors.Wrap(err, "writeFileWithTemp")
	}

	return nil
}
