package main

import (
	"path/filepath"
	"strconv"
	"testing"

	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/iamnotrodger/stdtype/internal/local"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsConfig(t *testing.T) {
	file := "settings.json"

	t.Run("when flag is set, should override fields", func(t *testing.T) {
		v := viper.New()
		path := filepath.Join(t.TempDir(), file)

		settingsRepo, err := local.NewSettingsRepo(v, path)
		require.NoError(t, err)

		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		SetSettingsFlag(flags)
		err = v.BindPFlags(flags)
		require.NoError(t, err)

		want := stdtype.DefaultSettings()
		want.Mode = stdtype.WordsMode.String()
		want.Duration = 60
		want.WordCount = 50

		err = flags.Parse([]string{
			"--mode", want.Mode,
			"--duration", strconv.Itoa(want.Duration),
			"--words", strconv.Itoa(want.WordCount),
		})
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("when env var is set, should override fields", func(t *testing.T) {
		want := stdtype.DefaultSettings()
		want.Mode = stdtype.WordsMode.String()
		want.Duration = 60
		want.WordCount = 50

		t.Setenv("STDTYPE_TEST_MODE", want.Mode)
		t.Setenv("STDTYPE_TEST_DURATION", strconv.Itoa(want.Duration))
		t.Setenv("STDTYPE_TEST_WORDS", strconv.Itoa(want.WordCount))

		v := viper.New()
		v.SetEnvPrefix("stdtype_test")
		v.AutomaticEnv()

		path := filepath.Join(t.TempDir(), file)

		settingsRepo, err := local.NewSettingsRepo(v, path)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
