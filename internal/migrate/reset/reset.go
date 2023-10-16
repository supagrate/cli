package reset

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	_init "github.com/supagrate/cli/internal/init"
	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/migrations"
)

func Run(cmd *cobra.Command, fs afero.Fs) {
	_init.EnsureInitialization()
	utils.UseDBEnvironmentVariables(cmd)

	connection := utils.Connection{
		Host:     cmd.Flag("db-host").Value.String(),
		Port:     cmd.Flag("db-port").Value.String(),
		User:     cmd.Flag("db-user").Value.String(),
		Password: cmd.Flag("db-password").Value.String(),
		Name:     cmd.Flag("db-name").Value.String(),
	}

	db := utils.ConnectDatabase(connection)

	// Resetting the Supagrate migration table
	fmt.Println("Resetting Supagrate migration table...")
	utils.ResetMigrationTable(db)
	utils.EnsureMigrationTable(db)

	// Resetting the public schema
	fmt.Println("Resetting public schema...")
	utils.ResetPublicSchema(db)

	// Running all migrations again
	migrationsToApply := migrations.ReadMigrationsFromFilesystem(fs)

	for _, migration := range migrationsToApply {
		fmt.Println("Applying " + utils.Yellow(migration.FileName) + "...")
		migration.Apply(db)
	}

	fmt.Println(utils.Green("Reset complete!"))

	db.Close()
}
