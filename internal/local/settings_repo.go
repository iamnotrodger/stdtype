package local

import (
	"os"
	"path/filepath"

	"github.com/go-viper/mapstructure/v2"
	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var _ stdtype.SettingsRepo = (*SettingsRepo)(nil)

type SettingsRepo struct {
	v    *viper.Viper
	path string
}

func NewSettingsRepo(v *viper.Viper, path string) (*SettingsRepo, error) {
	err := setDefaultSettings(v, stdtype.DefaultSettings())
	if err != nil {
		return nil, errors.Wrap(err, "setDefaultSettings")
	}

	return &SettingsRepo{v, path}, nil
}

func (s *SettingsRepo) GetSettings() (stdtype.Settings, error) {
	var settings stdtype.Settings

	err := s.v.ReadInConfig()
	if err != nil && !isNotFound(err) {
		return settings, errors.Wrap(err, "viper.ReadInConfig")
	}

	decodeOpts := func(dc *mapstructure.DecoderConfig) {
		dc.WeaklyTypedInput = true
	}

	if err := s.v.Unmarshal(&settings, decodeOpts); err != nil {
		return settings, errors.Wrap(err, "viper.Unmarshal")
	}

	return settings, nil
}

func (s *SettingsRepo) UpdateSettings(settings stdtype.Settings) error {
	var values map[string]any
	if err := mapstructure.Decode(settings, &values); err != nil {
		return errors.Wrap(err, "mapstructure.Decode")
	}

	for key, val := range values {
		s.v.Set(key, val)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return errors.Wrap(err, "os.Mkdir")
	}

	if err := s.v.WriteConfig(); err != nil {
		return errors.Wrap(err, "viper.WriteConfig")
	}

	return nil
}

func setDefaultSettings(v *viper.Viper, defaults stdtype.Settings) error {
	var values map[string]any
	if err := mapstructure.Decode(defaults, &values); err != nil {
		return errors.Wrap(err, "mapstructure.Decode")
	}

	for key, value := range values {
		v.SetDefault(key, value)
	}

	return nil
}

func isNotFound(err error) bool {
	var notFound viper.ConfigFileNotFoundError

	return errors.Is(err, os.ErrNotExist) || errors.As(err, &notFound)
}
