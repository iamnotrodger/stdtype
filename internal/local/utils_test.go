package local

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir(t *testing.T) {
	t.Run("should use XDG_CONFIG_HOME with app name", func(t *testing.T) {
		tmp := t.TempDir()
		want := filepath.Join(tmp, appName)

		t.Setenv("XDG_CONFIG_HOME", tmp)

		got, err := configDir()
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("should use home dir when XDG_CONFIG_HOME is unset on macOS", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("macOs-specific fallback")
		}

		t.Setenv("XDG_CONFIG_HOME", "")

		tmp, err := os.UserHomeDir()
		require.NoError(t, err)

		want := filepath.Join(tmp, appName)

		got, err := configDir()
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}

func TestWriteFileWithTemp(t *testing.T) {
	outputFile := "out.json"

	t.Run("should write data to the target path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, outputFile)

		want := "hello"
		err := writeFileWithTemp(dir, path, []byte(want))
		require.NoError(t, err)

		got, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Equal(t, want, string(got))
		assertNoTempFiles(t, dir)
	})

	t.Run("should create the target directory when missing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "cfg")
		path := filepath.Join(dir, outputFile)

		want := "hello"
		err := writeFileWithTemp(dir, path, []byte(want))
		require.NoError(t, err)

		got, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Equal(t, want, string(got))
	})

	t.Run("should overwrite an existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, outputFile)

		err := os.WriteFile(path, []byte("old"), 0o600)
		require.NoError(t, err)

		want := "new"
		err = writeFileWithTemp(dir, path, []byte(want))
		require.NoError(t, err)

		got, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Equal(t, want, string(got))
		assertNoTempFiles(t, dir)
	})
}

func assertNoTempFiles(tb testing.TB, dir string) {
	tb.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(tb, err)

	for _, e := range entries {
		assert.False(
			tb,
			strings.HasPrefix(e.Name(), ".tmp-"),
			"leftover temp files: %s", e.Name(),
		)
	}
}
