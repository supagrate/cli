package status

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/supagrate/cli/internal/utils"
	"github.com/supagrate/cli/migrations"
)

func Run(cmd *cobra.Command, fs afero.Fs) {
	connectionString := cmd.Flag("connection").Value.String()
	var conn *utils.Connection
	if connectionString != "" {
		conn = &utils.Connection{Connection: utils.ParseConnectionString(connectionString)}
	} else {
		conn = &utils.Connection{Connection: nil}
	}
	db := utils.ConnectDatabase(*conn)

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
