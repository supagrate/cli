package new

import (
	"fmt"

	"github.com/spf13/afero"

	_init "github.com/supagrate/cli/internal/init"
	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/migrations"
)

func Run(migrationName string, fs afero.Fs) error {
	_init.EnsureInitialization()

	migrationDir := migrations.EnsureMigrationDirectoryExists(fs, migrationName)

	migrations.GenerateMigrationFile(fs, migrationDir, "down.sql")
	migrations.GenerateMigrationFile(fs, migrationDir, "up.sql")

	fmt.Println(utils.Green("New migration created at ") + migrationDir)

	return nil
}
