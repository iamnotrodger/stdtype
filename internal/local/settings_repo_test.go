package local

import (
	"os"
	"path/filepath"
	"testing"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsRepo(t *testing.T) {
	file := "settings.json"

	t.Run("should return defaults when no settings exist", func(t *testing.T) {
		settingsRepo, err := NewSettingsRepo(t.TempDir(), file)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		want := stdtype.DefaultSettings()
		assert.Equal(t, want, got)
	})

	t.Run("should fall back to defaults for fields missing from file", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(
			filepath.Join(dir, file),
			[]byte(`{"mode":"words"}`),
			0o644,
		)
		require.NoError(t, err)

		settingsRepo, err := NewSettingsRepo(dir, file)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		want := stdtype.DefaultSettings()
		want.Mode = stdtype.WordsMode.String()

		assert.Equal(t, want, got)
	})

	t.Run("should update and get updated settings", func(t *testing.T) {
		settingsRepo, err := NewSettingsRepo(t.TempDir(), file)
		require.NoError(t, err)

		want := stdtype.Settings{
			Mode:      stdtype.WordsMode.String(),
			Duration:  60,
			WordCount: 50,
			Theme:     "dark",
		}

		err = settingsRepo.UpdateSettings(want)
		require.NoError(t, err)

		got, err := settingsRepo.GetSettings()
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
