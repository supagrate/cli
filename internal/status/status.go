package status

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/migrations"
)

func Run(cmd *cobra.Command, fs afero.Fs) {
	utils.UseDBEnvironmentVariables(cmd)

	connection := utils.Connection{
		Host:     cmd.Flag("db-host").Value.String(),
		Port:     cmd.Flag("db-port").Value.String(),
		User:     cmd.Flag("db-user").Value.String(),
		Password: cmd.Flag("db-password").Value.String(),
		Name:     cmd.Flag("db-name").Value.String(),
	}

	db := utils.ConnectDatabase(connection)

	utils.EnsureMigrationTable(db)

	// Fetch all existing migrations in DB
	dbMigrations := migrations.ReadMigrationsFromDB(db)

	// Fetch all existing migrations in filesystem
	fsMigrations := migrations.ReadMigrationsFromFilesystem(fs)

	// Compare the two and list status of each migration
	status := migrations.FindMigrationStatus(fsMigrations, dbMigrations)

	RenderStatus(status)

	db.Close()
}
