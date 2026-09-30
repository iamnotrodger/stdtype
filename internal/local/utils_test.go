package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteFileWithTemp(t *testing.T) {
	outputFile := "out.json"

	t.Run("should write data to the target path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, outputFile)

		want := "hello"
		err := writeFileWithTemp(dir, path, []byte(want))
		assert.NoError(t, err)

		got, err := os.ReadFile(path)
		assert.NoError(t, err)

		assert.Equal(t, want, string(got))
		assertNoTempFiles(t, dir)
	})

	t.Run("should create the target directory when missing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "cfg")
		path := filepath.Join(dir, outputFile)

		want := "hello"
		err := writeFileWithTemp(dir, path, []byte(want))
		assert.NoError(t, err)

		got, err := os.ReadFile(path)
		assert.NoError(t, err)

		assert.Equal(t, want, string(got))
	})

	t.Run("should overwrite an existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, outputFile)

		err := os.WriteFile(path, []byte("old"), 0o600)
		assert.NoError(t, err)

		want := "new"
		err = writeFileWithTemp(dir, path, []byte(want))
		assert.NoError(t, err)

		got, err := os.ReadFile(path)
		assert.NoError(t, err)

		assert.Equal(t, want, string(got))
		assertNoTempFiles(t, dir)
	})
}

func assertNoTempFiles(tb testing.TB, dir string) {
	tb.Helper()

	entries, err := os.ReadDir(dir)
	assert.NoError(tb, err)

	for _, e := range entries {
		assert.False(
			tb,
			strings.HasPrefix(e.Name(), ".tmp-"),
			"leftover temp files: %s", e.Name(),
		)
	}
}
