package local

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

const appName = "stdtype"

func configDir() (string, error) {
	base, err := baseConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, appName), nil
}

func baseConfigDir() (string, error) {
	if runtime.GOOS == "darwin" {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return xdg, nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "os.UserHomeDir")
		}

		return filepath.Join(home, ".config"), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.Wrap(err, "os.UserConfigDir")
	}

	return dir, nil
}

func writeFileWithTemp(dir string, path string, data []byte) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errors.Wrap(err, "os.MkdirAll")
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*.json")
	if err != nil {
		return errors.Wrap(err, "os.CreateTemp")
	}

	tmpName := tmp.Name()
	defer func() {
		os.Remove(tmpName)
	}()

	_, err = tmp.Write(data)
	if err != nil {
		tmp.Close() //nolint:errcheck,gosec

		return errors.Wrap(err, "tmp.Write")
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck,gosec

		return errors.Wrap(err, "tmp.Close")
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName) //nolint:errcheck,gosec

		return errors.Wrap(err, "os.Rename")
	}

	return nil
}
