package utils

import (
	"errors"
	"os"

	"github.com/spf13/afero"
)

const SupagrateDirectory = "supagrate"

func MkdirIfNotExist(fs afero.Fs, path string) error {
	if err := fs.MkdirAll(path, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	return nil
}
