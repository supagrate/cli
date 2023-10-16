package seeds

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/supagrate/cli/internal/utils"
)

const SeedsDirectory = utils.SupagrateDirectory + "/seeds"

type Seed struct {
	FileName string
	Content  string
}

func (seed Seed) Apply(db *sql.DB) {
	logrus.Info("Applying seed: " + seed.FileName)

	_, err := db.Exec(seed.Content)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	logrus.Info("Seed applied: " + seed.FileName)
}

func EnsureSeedsDirectoryExists(fs afero.Fs) string {
	seedsDir := filepath.Join(".", SeedsDirectory)

	if err := utils.MkdirIfNotExist(fs, filepath.Dir(seedsDir)); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return seedsDir
}

func FindSeed(fs afero.Fs, filename string) Seed {
	path := filepath.Join(SeedsDirectory, filename+".sql")

	if _, err := fs.Stat(path); os.IsNotExist(err) {
		logrus.Error("Seed not found: " + filename)
		os.Exit(1)
	}

	contents, err := afero.ReadFile(fs, path)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return Seed{
		FileName: filename,
		Content:  string(contents),
	}
}

func GenerateSeedFile(fs afero.Fs, filename string) string {
	path := filepath.Join(SeedsDirectory, filename+".sql")

	if err := utils.MkdirIfNotExist(fs, filepath.Dir(path)); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	f, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	defer f.Close()

	return path
}
