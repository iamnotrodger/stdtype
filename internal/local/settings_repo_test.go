package local

import (
	"os"
	"path/filepath"
	"testing"

	stdtype "github.com/iamnotrodger/stdtype"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustViper(tb testing.TB, file string) (*viper.Viper, string) {
	tb.Helper()

	path := filepath.Join(tb.TempDir(), file)

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("json")

	return v, path
}

func TestSettingsRepo(t *testing.T) {
	file := "settings.json"

	t.Run("should return defaults when no settings exist", func(t *testing.T) {
		v, path := mustViper(t, file)
		settingsRepo, err := NewSettingsRepo(v, path)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		want := stdtype.DefaultSettings()
		assert.Equal(t, want, got)
	})

	t.Run("should fall back to defaults for fields missing from file", func(t *testing.T) {
		v, path := mustViper(t, file)

		err := os.WriteFile(
			path,
			[]byte(`{"mode":"words"}`),
			0o644,
		)
		require.NoError(t, err)

		settingsRepo, err := NewSettingsRepo(v, path)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		want := stdtype.DefaultSettings()
		want.Mode = stdtype.WordsMode.String()

		assert.Equal(t, want, got)
	})

	t.Run("should update and get updated settings", func(t *testing.T) {
		v, path := mustViper(t, file)
		settingsRepo, err := NewSettingsRepo(v, path)
		require.NoError(t, err)

		want := stdtype.DefaultSettings()
		want.Mode = stdtype.WordsMode.String()
		want.Duration = 60
		want.WordCount = 50
		want.Theme = "dark"

		err = settingsRepo.UpdateSettings(want)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
